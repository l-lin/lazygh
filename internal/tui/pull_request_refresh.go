package tui

import "github.com/jesseduffield/gocui"

const (
	pullRequestRefreshActionTitle        = "Refresh current PR information"
	pullRequestRefreshSuccessMessage     = "Pull request refreshed"
	pullRequestListRefreshActionTitle    = "Refresh PR list"
	pullRequestListRefreshSuccessMessage = "Pull request list refreshed"
)

func (program *Program) refreshPullRequestAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	target, targetOK := program.selectedPullRequestActionTarget()
	summary, summaryOK := program.currentPullRequestSummary()
	if targetOK && summaryOK {
		requested = MsgRefreshPullRequestRequested{Target: target, Summary: summary}
	}
	return actionsPopupAction{
		id:        "refresh-current-pull-request-information",
		title:     pullRequestRefreshActionTitle,
		icon:      actionsPopupRefreshPullRequestIcon,
		requested: requested,
	}
}

func (program *Program) refreshPullRequestListAction() actionsPopupAction {
	return actionsPopupAction{
		id:        "refresh-pull-request-list",
		title:     pullRequestListRefreshActionTitle,
		icon:      actionsPopupRefreshPullRequestIcon,
		requested: MsgRefreshPullRequestListRequested{},
	}
}

func (program *Program) requestRefreshPullRequestList(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgRefreshPullRequestListRequested{})
}

func (program *Program) requestRefreshCurrentPullRequest(gui *gocui.Gui) error {
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
