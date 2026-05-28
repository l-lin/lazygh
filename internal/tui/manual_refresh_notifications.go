package tui

func (program *Program) markManualNotificationRefresh() bool {
	if program == nil {
		return false
	}
	updatedState, marked := program.manualRefreshState.withNotificationPending()
	if !marked {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}

func (program *Program) consumeManualNotificationRefresh() bool {
	if program == nil {
		return false
	}
	updatedState, pending := program.manualRefreshState.withoutNotificationPending()
	if !pending {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}
