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
	context, ok := program.currentPullRequestDescriptionCursorContext()
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}

	entry, ok := program.browserOverviewBuildEntryAtDetailCursorDocument(context.selection.document)
	if !ok || strings.TrimSpace(entry.Link) == "" {
		return pullRequestBuildRunTarget{}, false
	}
	check, ok := pullRequestStatusCheckMatchingEntry(context.detail.StatusCheckRollup, entry)
	if !ok {
		return pullRequestBuildRunTarget{}, false
	}
	repository := pullRequestRepositoryName(context.summary.Repository)
	return pullRequestBuildRunTarget{
		summary: context.summary,
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
	return program.dispatch(gui, MsgPullRequestBuildRunLoadRequested{Target: target})
}

func (program *Program) startPullRequestBuildRunJobLogLoad(gui *gocui.Gui, summary githubdomain.PullRequest, check githubdomain.PullRequestStatusCheck) error {
	if !program.hasBuildQueries() || program.pullRequestBuildRunLoad != nil {
		return nil
	}

	return program.dispatch(gui, MsgPullRequestBuildRunJobLogLoadRequested{Summary: summary, Check: check})
}

func (program *Program) pullRequestBuildRunActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "view-build-run",
		title:   pullRequestBuildRunActionTitle,
		icon:    actionsPopupBuildRunIcon,
		execute: actionsPopupExecuteErr(program.executePullRequestBuildRunAction),
	}
}

func (program *Program) pullRequestBuildRunLogsActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "view-build-run-job-logs",
		title:   pullRequestBuildRunLogsActionTitle,
		icon:    actionsPopupBuildRunLogsIcon,
		execute: actionsPopupExecuteErr(program.executePullRequestBuildRunLogsAction),
	}
}

func (program *Program) executePullRequestBuildRunAction(gui *gocui.Gui) error {
	target, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	if err := program.startPullRequestBuildRunLoad(gui, target.summary, target.check); err != nil {
		return err
	}
	return program.closeActionsPopupIfVisible(gui)
}

func (program *Program) executePullRequestBuildRunLogsAction(gui *gocui.Gui) error {
	target, ok := program.currentPullRequestBuildRunTargetAtDetailCursor()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	if err := program.startPullRequestBuildRunJobLogLoad(gui, target.summary, target.check); err != nil {
		return err
	}
	return program.closeActionsPopupIfVisible(gui)
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

func (program *Program) browserOverviewBuildEntryAtDetailCursorDocument(document detailDocument) (pullRequestOverviewEntry, bool) {
	summary, detail, ok := program.currentPullRequestDescriptionSummaryAndDetail()
	if !ok {
		return pullRequestOverviewEntry{}, false
	}

	sectionAtCursor, ok := program.browserOverviewSectionAtCursor(summary, detail, document.width, program.detailState.viewState.cursor.line)
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
