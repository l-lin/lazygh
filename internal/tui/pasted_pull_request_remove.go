package tui

import "github.com/jesseduffield/gocui"

const removePastedPullRequestActionTitle = "Remove from pasted PRs"

func (program *Program) canRemovePastedPullRequest() bool {
	if program == nil || program.model == nil || program.reviewModeActive() {
		return false
	}
	if program.model.Focus() != FocusPullRequestsView {
		return false
	}
	if !program.isPastedPullRequestTab(program.model.ActivePullRequestTab()) {
		return false
	}
	_, ok := program.model.SelectedPullRequestSummary()
	return ok
}

func (program *Program) removePastedPullRequestActionsPopupAction() actionsPopupAction {
	var requested Msg = actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if program.canRemovePastedPullRequest() {
		requested = MsgRemovePastedPullRequestRequested{}
	}
	return actionsPopupAction{
		id:        "remove-pasted-pull-request",
		title:     removePastedPullRequestActionTitle,
		icon:      iconDelete,
		keywords:  []string{"remove", "delete", "pasted"},
		requested: requested,
	}
}

func (program *Program) removePastedPullRequestShortcut(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgRemovePastedPullRequestRequested{})
}

func (program *Program) applyRemovePastedPullRequestRequested() {
	if !program.canRemovePastedPullRequest() {
		return
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return
	}

	program.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestRemoved(summary)
	})
	program.queueSavePastedPullRequestsPersistentCache()
	if program.navigationState.openedPullRequestSummary != nil && samePullRequestIdentity(*program.navigationState.openedPullRequestSummary, summary) {
		program.clearOpenedPullRequestSummaryState()
	}
	program.closeActionsPopupForAcceptedRequest()
	program.syncPastedPullRequestTab()
}
