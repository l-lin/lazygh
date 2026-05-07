package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	pullRequestBuildRunUnavailableMessage  = "󰅚 Build run unavailable"
	pullRequestBuildLogsUnavailableMessage = "󰅚 Build logs unavailable"
	pullRequestBuildRunActionTitle         = "View build run"
	pullRequestBuildRunLogsActionTitle     = "View job logs"
	actionsPopupBuildActionIcon            = ""
	actionsPopupBuildRunIcon               = actionsPopupBuildActionIcon
	actionsPopupBuildRunLogsIcon           = actionsPopupBuildActionIcon
)

type pullRequestBuildRunTarget struct {
	summary      githubcli.PullRequest
	check        githubcli.PullRequestStatusCheck
	popupContent pullRequestBuildRunPopupContent
}

func (program *Program) detailCursorHasBuildLink() bool {
	_, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	return ok
}

func (program *Program) currentPullRequestBuildRunTargetAtDetailCursor() (pullRequestBuildRunTarget, bool) {
	if program.reviewSession.active || !program.shouldShowPullRequestDetailTabs() || program.activeDetailTab != DescriptionDetailTab {
		return pullRequestBuildRunTarget{}, false
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestBuildRunTarget{}, false
	}

	actualView := program.resolveView(program.gui, nil, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	entry, ok := program.browserOverviewBuildEntryAtDetailCursor(actualView)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return pullRequestBuildRunTarget{}, false
	}
	check, ok := pullRequestStatusCheckMatchingEntry(result.detail.StatusCheckRollup, entry)
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

func (program *Program) startPullRequestBuildRunLoad(gui *gocui.Gui, summary githubcli.PullRequest, check githubcli.PullRequestStatusCheck) error {
	if program.githubLoader == nil || program.pullRequestBuildRunLoad != nil {
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
	program.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: githubcli.FormatPullRequestBuildRunCommand(repository, check)}
	program.asyncRunner.Go(func() {
		program.loadPullRequestBuildRun(gui, repository, target)
	})
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) loadPullRequestBuildRun(gui *gocui.Gui, repository string, target pullRequestBuildRunTarget) {
	rawRunOutput, err := program.githubLoader.GetPullRequestBuildRun(repository, target.check)
	jobs := []githubcli.PullRequestBuildRunJob(nil)
	if err == nil {
		jobs, _ = program.githubLoader.GetPullRequestBuildRunJobs(repository, target.check)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.pullRequestBuildRunLoad = nil
		if err != nil {
			program.setFeedback(program.model.Focus(), pullRequestBuildRunUnavailableMessage)
			return program.refreshViews(gui)
		}

		target.popupContent.body = rawRunOutput
		target.popupContent.jobs = jobs
		return program.openPullRequestBuildRunPopup(gui, target.popupContent)
	})
}

func (program *Program) startPullRequestBuildRunJobLogLoad(gui *gocui.Gui, summary githubcli.PullRequest, check githubcli.PullRequestStatusCheck) error {
	if program.githubLoader == nil || program.pullRequestBuildRunLoad != nil {
		return nil
	}

	repository := pullRequestRepositoryName(summary.Repository)
	if repository == "" || repository == "-" {
		return nil
	}

	program.feedbackMessage = ""
	program.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: githubcli.FormatPullRequestBuildRunJobsCommand(repository, check)}
	program.asyncRunner.Go(func() {
		program.loadPullRequestBuildRunJobLog(gui, repository, check)
	})
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) loadPullRequestBuildRunJobLog(gui *gocui.Gui, repository string, check githubcli.PullRequestStatusCheck) {
	job, rawLogOutput, err := program.githubLoader.GetPullRequestBuildRunJobLogForCheck(repository, check)
	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.pullRequestBuildRunLoad = nil
		if err != nil {
			program.setFeedback(program.model.Focus(), pullRequestBuildLogsUnavailableMessage)
			return program.refreshViews(gui)
		}

		return program.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
			title:         pullRequestBuildRunLogsPopupTitle(job.Name),
			runURL:        strings.TrimSpace(job.URL),
			repository:    repository,
			body:          sanitizePullRequestBuildRunLog(rawLogOutput),
			widthPercent:  pullRequestBuildLogsPopupWidthPercent,
			heightPercent: pullRequestBuildLogsPopupHeightPercent,
		})
	})
}

func (program *Program) pullRequestBuildRunActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "view-build-run",
		title:    pullRequestBuildRunActionTitle,
		icon:     actionsPopupBuildRunIcon,
		keywords: []string{"build", "run", "workflow", "checks", "details"},
		execute:  program.executePullRequestBuildRunAction,
	}
}

func (program *Program) pullRequestBuildRunLogsActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "view-build-run-job-logs",
		title:    pullRequestBuildRunLogsActionTitle,
		icon:     actionsPopupBuildRunLogsIcon,
		keywords: []string{"job", "logs", "build", "run", "workflow"},
		execute:  program.executePullRequestBuildRunLogsAction,
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

func pullRequestStatusCheckMatchingEntry(checks []githubcli.PullRequestStatusCheck, entry pullRequestOverviewEntry) (githubcli.PullRequestStatusCheck, bool) {
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
		return githubcli.PullRequestStatusCheck{}, false
	}
	for _, check := range checks {
		if strings.EqualFold(strings.TrimSpace(buildPullRequestBuildEntry(check).Label), trimmedLabel) {
			return check, true
		}
	}
	return githubcli.PullRequestStatusCheck{}, false
}

func checkTitleForPullRequestBuildRunPopup(check githubcli.PullRequestStatusCheck) string {
	return pullRequestOverviewCheckDisplayName(check)
}

func (program *Program) browserOverviewBuildEntryAtDetailCursor(view *gocui.View) (pullRequestOverviewEntry, bool) {
	document := program.currentDetailDocument(view)
	program.syncDetailViewState(document, viewPageSize(view))
	return program.browserOverviewBuildEntryAtDetailCursorDocument(document)
}

func (program *Program) browserOverviewBuildEntryAtDetailCursorDocument(document detailDocument) (pullRequestOverviewEntry, bool) {
	if program.reviewSession.active || !program.shouldShowPullRequestDetailTabs() || program.activeDetailTab != DescriptionDetailTab {
		return pullRequestOverviewEntry{}, false
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return pullRequestOverviewEntry{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return pullRequestOverviewEntry{}, false
	}

	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, result.detail, document.width, program.detailViewState.cursor.line)
	if !ok || !sectionAtCursor.inBody || !strings.EqualFold(strings.TrimSpace(sectionAtCursor.section.overviewBlockTitle), "Builds") {
		return pullRequestOverviewEntry{}, false
	}
	entry, ok := pullRequestOverviewEntryAtBodyLine(sectionAtCursor.section, sectionAtCursor.bodyLine)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return pullRequestOverviewEntry{}, false
	}
	return entry, true
}
