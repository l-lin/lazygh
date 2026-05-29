package tui

func (program *Program) markManualNotificationRefresh() bool {
	if program == nil {
		return false
	}

	marked := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualMarked := state.withNotificationPending()
		if !actualMarked {
			return state
		}
		marked = true
		return updatedState
	})
	return marked
}

func (program *Program) consumeManualNotificationRefresh() bool {
	if program == nil {
		return false
	}

	pending := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualPending := state.withoutNotificationPending()
		if !actualPending {
			return state
		}
		pending = true
		return updatedState
	})
	return pending
}
