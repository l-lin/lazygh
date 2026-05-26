package tui

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
