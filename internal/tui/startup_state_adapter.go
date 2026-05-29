package tui

func (program *Program) updateStartupState(transition func(startupStateModel) startupStateModel) {
	if program == nil {
		return
	}
	program.startupState = transition(program.startupState)
}

func (program *Program) markAppStarted() {
	program.updateStartupState(func(state startupStateModel) startupStateModel {
		return state.withAppStarted()
	})
}

func (program *Program) advanceStartupLoadingSpinnerFrame(frameCount int) {
	program.updateStartupState(func(state startupStateModel) startupStateModel {
		return state.withAdvancedLoadingSpinnerFrame(frameCount)
	})
}
