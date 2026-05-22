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
		return program.handleActionsPopupActionResult(gui, program.executeRefreshPullRequestListAction(gui))
	case sidePanelNotificationsViewNumber:
		return program.handleActionsPopupActionResult(gui, program.executeRefreshNotificationsAction(gui))
	case mainPanelViewNumber:
		if !program.actionContext().IsPullRequestContext() {
			return nil
		}
		return program.handleActionsPopupActionResult(gui, program.executeRefreshPullRequestAction(gui))
	default:
		return nil
	}
}

func (program *Program) executeRefreshNotificationsAction(gui *gocui.Gui) actionsPopupActionResult {
	program.markManualNotificationRefresh()
	program.setFeedback(program.model.Focus(), notificationsRefreshSuccessMessage)
	program.reloadNotifications(gui)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) markManualNotificationRefresh() {
	if program == nil {
		return
	}
	program.manualNotificationRefreshError = true
}

func (program *Program) consumeManualNotificationRefresh() bool {
	if program == nil {
		return false
	}
	pending := program.manualNotificationRefreshError
	program.manualNotificationRefreshError = false
	return pending
}
