package tui

func (program *Program) updateManualRefreshState(transition func(manualRefreshStateModel) manualRefreshStateModel) {
	if program == nil {
		return
	}
	program.manualRefreshState = transition(program.manualRefreshState)
}
