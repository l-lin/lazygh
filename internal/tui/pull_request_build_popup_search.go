package tui

func (program *Program) pullRequestBuildRunPopupSearchActive() bool {
	return program != nil && program.pullRequestBuildRunPopup != nil && program.pullRequestBuildRunPopup.searchActive
}

func (program *Program) searchPromptVisible() bool {
	return program.model.SearchActive() || program.pullRequestBuildRunPopupSearchActive()
}

func (program *Program) startPullRequestBuildRunPopupSearch() {
	if program == nil || program.pullRequestBuildRunPopup == nil {
		return
	}
	program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		return state.withSearchOpened()
	})
}
