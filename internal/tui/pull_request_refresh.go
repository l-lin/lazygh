package tui

import "github.com/jesseduffield/gocui"

const (
	pullRequestRefreshActionTitle    = "Refresh current PR information"
	pullRequestRefreshSuccessMessage = "Pull request refreshed"
)

func (program *Program) refreshPullRequestAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "refresh-current-pull-request-information",
		title:    pullRequestRefreshActionTitle,
		icon:     actionsPopupRefreshPullRequestIcon,
		keywords: []string{"refresh", "reload", "sync", "current", "pull request", "stale"},
		execute:  program.executeRefreshPullRequestAction,
	}
}

func (program *Program) executeRefreshPullRequestAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}

	program.invalidatePullRequestDetail(target.repository, target.number)
	if program.reviewSession.active {
		program.invalidatePullRequestDiff(target.repository, target.number)
	} else {
		program.reloadActivePullRequestsTab(gui)
	}
	program.setFeedback(program.model.Focus(), pullRequestRefreshSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}
