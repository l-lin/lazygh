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
	program.reloadActivePullRequestsTab(gui)
	program.setFeedback(program.model.Focus(), pullRequestListRefreshSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executeRefreshPullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	if program.reviewModeActive() {
		program.invalidatePullRequestDiff(target.repository, target.number)
	} else {
		program.reloadActivePullRequestsTab(gui)
	}
	program.setFeedback(program.model.Focus(), pullRequestRefreshSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}
