package tui

import "github.com/jesseduffield/gocui"

const (
	pullRequestRefreshActionTitle        = "Refresh current PR information"
	pullRequestRefreshSuccessMessage     = "Pull request refreshed"
	pullRequestListRefreshActionTitle    = "Refresh PR list"
	pullRequestListRefreshSuccessMessage = "Pull request list refreshed"
)

func (program *Program) refreshPullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "refresh-current-pull-request-information",
		title:   pullRequestRefreshActionTitle,
		icon:    actionsPopupRefreshPullRequestIcon,
		execute: program.executeRefreshPullRequestAction,
	}
}

func (program *Program) refreshPullRequestListAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "refresh-pull-request-list",
		title:   pullRequestListRefreshActionTitle,
		icon:    actionsPopupRefreshPullRequestIcon,
		execute: program.executeRefreshPullRequestListAction,
	}
}

func (program *Program) executeRefreshPullRequestListAction(gui *gocui.Gui) actionsPopupActionResult {
	program.markManualPullRequestListRefresh(program.model.ActivePullRequestTab())
	program.setFeedback(program.model.Focus(), pullRequestListRefreshSuccessMessage)
	program.reloadActivePullRequestsTab(gui)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executeRefreshPullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	program.markManualPullRequestDetailRefresh(summary)
	program.setFeedback(program.model.Focus(), pullRequestRefreshSuccessMessage)
	program.markPullRequestDetailNeedsRefresh(summary)
	if program.reviewModeActive() {
		program.markManualPullRequestDiffRefresh(summary)
		program.markPullRequestDiffNeedsRefresh(summary)
	} else {
		program.markManualPullRequestListRefresh(program.model.ActivePullRequestTab())
		program.reloadActivePullRequestsTab(gui)
	}
	program.invalidatePersistentPullRequest(target.repository, target.number)
	return actionsPopupActionResult{closePopup: true}
}
