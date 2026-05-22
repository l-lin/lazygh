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
	pendingOperations := 0
	if gui != nil && program.hasPullRequestListQueries() && program.markManualPullRequestListRefresh(program.model.ActivePullRequestTab()) {
		pendingOperations++
	}
	program.beginManualRefresh(pullRequestListRefreshSuccessMessage, pendingOperations)
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

	pendingOperations := 0
	if program.hasDetailQueries() {
		if program.markManualPullRequestDetailRefresh(summary) {
			pendingOperations++
		}
		program.markPullRequestDetailNeedsRefresh(summary)
	}
	if program.reviewModeActive() {
		if program.hasDetailQueries() {
			if program.markManualPullRequestDiffRefresh(summary) {
				pendingOperations++
			}
			program.markPullRequestDiffNeedsRefresh(summary)
		}
	} else {
		if gui != nil && program.hasPullRequestListQueries() && program.markManualPullRequestListRefresh(program.model.ActivePullRequestTab()) {
			pendingOperations++
		}
	}
	program.beginManualRefresh(pullRequestRefreshSuccessMessage, pendingOperations)
	if !program.reviewModeActive() {
		program.reloadActivePullRequestsTab(gui)
	}
	program.invalidatePersistentPullRequest(target.repository, target.number)
	return actionsPopupActionResult{closePopup: true}
}
