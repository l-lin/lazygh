package tui

func (state startupStateModel) withAppStarted() startupStateModel {
	state.appStarted = true
	return state
}

func (state startupStateModel) withAdvancedLoadingSpinnerFrame(frameCount int) startupStateModel {
	if frameCount <= 0 {
		return state
	}
	state.loadingSpinnerFrameIndex = (state.loadingSpinnerFrameIndex + 1) % frameCount
	return state
}
