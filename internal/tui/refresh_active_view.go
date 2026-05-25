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
	pendingOperations := 0
	if gui != nil && !program.reviewModeActive() && program.hasNotificationQueries() && program.markManualNotificationRefresh() {
		pendingOperations++
	}
	program.beginManualRefresh(notificationsRefreshSuccessMessage, pendingOperations)
	program.reloadNotifications(gui)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) markManualNotificationRefresh() bool {
	if program == nil {
		return false
	}
	program.manualRefreshState.notificationPending = true
	return true
}

func (program *Program) consumeManualNotificationRefresh() bool {
	if program == nil {
		return false
	}
	pending := program.manualRefreshState.notificationPending
	program.manualRefreshState.notificationPending = false
	return pending
}
