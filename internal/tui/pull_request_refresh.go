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
		execute: actionsPopupExecuteErr(program.executeRefreshPullRequestAction),
	}
}

func (program *Program) refreshPullRequestListAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "refresh-pull-request-list",
		title:   pullRequestListRefreshActionTitle,
		icon:    actionsPopupRefreshPullRequestIcon,
		execute: actionsPopupExecuteErr(program.executeRefreshPullRequestListAction),
	}
}

func (program *Program) executeRefreshPullRequestListAction(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgRefreshPullRequestListRequested{})
}

func (program *Program) executeRefreshPullRequestAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return errActionsPopupActionUnavailable
	}

	return program.dispatch(gui, MsgRefreshPullRequestRequested{Target: target, Summary: summary})
}
