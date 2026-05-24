package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	pullRequestBuildRunUnavailableMessage  = iconWarning + " Build run unavailable"
	pullRequestBuildLogsUnavailableMessage = iconWarning + " Build logs unavailable"
	pullRequestBuildRunActionTitle         = "View build run"
	pullRequestBuildRunLogsActionTitle     = "View job logs"
)

type pullRequestBuildRunTarget struct {
	summary      githubdomain.PullRequest
	check        githubdomain.PullRequestStatusCheck
	popupContent pullRequestBuildRunPopupContent
}

func (program *Program) detailCursorHasBuildLink() bool {
	_, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	return ok
}

func (program *Program) currentPullRequestBuildRunTargetAtDetailCursor() (pullRequestBuildRunTarget, bool) {
	summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}

	actualView := program.resolveView(program.gui, nil, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	entry, ok := program.browserOverviewBuildEntryAtDetailCursor(actualView)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return pullRequestBuildRunTarget{}, false
	}
	check, ok := pullRequestStatusCheckMatchingEntry(detail.StatusCheckRollup, entry)
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}
	repository := pullRequestRepositoryName(summary.Repository)
	return pullRequestBuildRunTarget{
		summary: summary,
		check:   check,
		popupContent: pullRequestBuildRunPopupContent{
			checkTitle: checkTitleForPullRequestBuildRunPopup(check),
			runURL:     strings.TrimSpace(check.Link),
			repository: repository,
		},
	}, true
}

func (program *Program) startPullRequestBuildRunLoad(gui *gocui.Gui, summary githubdomain.PullRequest, check githubdomain.PullRequestStatusCheck) error {
	if !program.hasBuildQueries() || program.pullRequestBuildRunLoad != nil {
		return nil
	}

	repository := pullRequestRepositoryName(summary.Repository)
	target := pullRequestBuildRunTarget{
		summary: summary,
		check:   check,
		popupContent: pullRequestBuildRunPopupContent{
			checkTitle: checkTitleForPullRequestBuildRunPopup(check),
			runURL:     strings.TrimSpace(check.Link),
			repository: repository,
		},
	}
	program.feedbackMessage = ""
	program.pullRequestBuildRunPopup = nil
	program.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: formatPullRequestBuildRunCommand(repository, check)}
	program.asyncRunner.Go(func() {
		program.loadPullRequestBuildRun(gui, repository, target)
	})
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) loadPullRequestBuildRun(gui *gocui.Gui, repository string, target pullRequestBuildRunTarget) {
	rawRunOutput, err := program.buildQueries.GetPullRequestBuildRun(repository, target.check)
	jobs := []githubdomain.PullRequestBuildRunJob(nil)
	jobsErr := error(nil)
	if err == nil {
		jobs, jobsErr = program.buildQueries.GetPullRequestBuildRunJobs(repository, target.check)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.pullRequestBuildRunLoad = nil
		if err != nil {
			program.reportError(gui, strings.TrimSpace(normalizeGHCommandError(err).Error()))
			return program.afterStateChange(gui)
		}

		target.popupContent.body = rawRunOutput
		target.popupContent.jobs = jobs
		if actualErr := program.openPullRequestBuildRunPopup(gui, target.popupContent); actualErr != nil {
			return actualErr
		}
		if jobsErr != nil {
			program.reportError(gui, strings.TrimSpace(normalizeGHCommandError(jobsErr).Error()))
		}
		return program.afterStateChange(gui)
	})
}

func (program *Program) startPullRequestBuildRunJobLogLoad(gui *gocui.Gui, summary githubdomain.PullRequest, check githubdomain.PullRequestStatusCheck) error {
	if !program.hasBuildQueries() || program.pullRequestBuildRunLoad != nil {
		return nil
	}

	repository := pullRequestRepositoryName(summary.Repository)
	if repository == "" || repository == "-" {
		return nil
	}

	program.feedbackMessage = ""
	program.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: formatPullRequestBuildRunJobsCommand(repository, check)}
	program.asyncRunner.Go(func() {
		program.loadPullRequestBuildRunJobLog(gui, repository, check)
	})
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) loadPullRequestBuildRunJobLog(gui *gocui.Gui, repository string, check githubdomain.PullRequestStatusCheck) {
	job, rawLogOutput, err := program.buildQueries.GetPullRequestBuildRunJobLogForCheck(repository, check)
	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.pullRequestBuildRunLoad = nil
		if err != nil {
			program.reportError(gui, strings.TrimSpace(normalizeGHCommandError(err).Error()))
			return program.afterStateChange(gui)
		}

		if actualErr := program.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
			title:         pullRequestBuildRunLogsPopupTitle(job.Name),
			runURL:        strings.TrimSpace(job.URL),
			repository:    repository,
			body:          sanitizePullRequestBuildRunLog(rawLogOutput),
			widthPercent:  pullRequestBuildLogsPopupWidthPercent,
			heightPercent: pullRequestBuildLogsPopupHeightPercent,
		}); actualErr != nil {
			return actualErr
		}
		return program.afterStateChange(gui)
	})
}

func (program *Program) pullRequestBuildRunActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "view-build-run",
		title:   pullRequestBuildRunActionTitle,
		icon:    actionsPopupBuildRunIcon,
		execute: program.executePullRequestBuildRunAction,
	}
}

func (program *Program) pullRequestBuildRunLogsActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "view-build-run-job-logs",
		title:   pullRequestBuildRunLogsActionTitle,
		icon:    actionsPopupBuildRunLogsIcon,
		execute: program.executePullRequestBuildRunLogsAction,
	}
}

func (program *Program) executePullRequestBuildRunAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.startPullRequestBuildRunLoad(gui, target.summary, target.check); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executePullRequestBuildRunLogsAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.startPullRequestBuildRunJobLogLoad(gui, target.summary, target.check); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func pullRequestStatusCheckMatchingEntry(checks []githubdomain.PullRequestStatusCheck, entry pullRequestOverviewEntry) (githubdomain.PullRequestStatusCheck, bool) {
	trimmedLink := strings.TrimSpace(entry.Link)
	if trimmedLink != "" {
		for _, check := range checks {
			if strings.EqualFold(strings.TrimSpace(check.Link), trimmedLink) {
				return check, true
			}
		}
	}

	trimmedLabel := strings.TrimSpace(entry.Label)
	if trimmedLabel == "" {
		return githubdomain.PullRequestStatusCheck{}, false
	}
	for _, check := range checks {
		if strings.EqualFold(strings.TrimSpace(buildPullRequestBuildEntry(check).Label), trimmedLabel) {
			return check, true
		}
	}
	return githubdomain.PullRequestStatusCheck{}, false
}

func checkTitleForPullRequestBuildRunPopup(check githubdomain.PullRequestStatusCheck) string {
	return pullRequestOverviewCheckDisplayName(check)
}

func (program *Program) browserOverviewBuildEntryAtDetailCursor(view *gocui.View) (pullRequestOverviewEntry, bool) {
	document := program.currentDetailDocument(view)
	program.syncDetailViewState(document, viewPageSize(view))
	return program.browserOverviewBuildEntryAtDetailCursorDocument(document)
}

func (program *Program) browserOverviewBuildEntryAtDetailCursorDocument(document detailDocument) (pullRequestOverviewEntry, bool) {
	summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestOverviewEntry{}, false
	}

	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, detail, document.width, program.detailViewState.cursor.line)
	if !ok || !sectionAtCursor.inBody || !strings.EqualFold(strings.TrimSpace(sectionAtCursor.section.overviewBlockTitle), "Builds") {
		return pullRequestOverviewEntry{}, false
	}
	entry, ok := pullRequestOverviewEntryAtBodyLine(sectionAtCursor.section, sectionAtCursor.bodyLine)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return pullRequestOverviewEntry{}, false
	}
	return entry, true
}

func (program *Program) currentPullRequestDescriptionSummaryAndDetail() (githubdomain.PullRequest, githubdomain.PullRequestDetail, bool) {
	actionContext := program.actionContext()
	if actionContext.IsReviewContext() {
		summary, detail, ok := program.reviewSessionDescriptionSummaryAndDetail()
		if !ok {
			return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
		}
		return summary, detail, true
	}
	if !actionContext.ShowsPullRequestDescription() {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return githubdomain.PullRequest{}, githubdomain.PullRequestDetail{}, false
	}
	return summary, result.detail, true
}

func (program *Program) detailCursorActionsAvailable() bool {
	actionContext := program.actionContext()
	if actionContext.ActiveView.Focus == FocusDetailView {
		return true
	}
	return actionContext.IsReviewContext() && actionContext.ActiveView.Focus == FocusUserView && actionContext.ShowsPullRequestDescription()
}
