package tui

import "github.com/jesseduffield/gocui"

const notificationsRefreshSuccessMessage = "Notifications refreshed"

func (program *Program) refreshActiveView(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()

	state := program.screenState()
	switch state.ActiveView().Number {
	case sidePanelUserViewNumber:
		return nil
	case sidePanelPullRequestsViewNumber:
		if state.Mode != ScreenModeBrowser {
			return nil
		}
		return program.handleActionsPopupActionError(gui, program.executeRefreshPullRequestListAction(gui))
	case sidePanelNotificationsViewNumber:
		return program.handleActionsPopupActionError(gui, program.executeRefreshNotificationsAction(gui))
	case mainPanelViewNumber:
		if !program.actionContext().IsPullRequestContext() {
			return nil
		}
		return program.handleActionsPopupActionError(gui, program.executeRefreshPullRequestAction(gui))
	default:
		return nil
	}
}

func (program *Program) executeRefreshNotificationsAction(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgRefreshNotificationsRequested{})
}
