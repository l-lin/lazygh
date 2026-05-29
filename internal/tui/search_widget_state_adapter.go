package tui

func (program *Program) updateSearchWidget(transition func(searchWidgetState) searchWidgetState) {
	if program == nil {
		return
	}
	program.searchWidget = transition(program.searchWidget)
}

func (program *Program) openSearchWidgetEditor(text string) {
	program.updateSearchWidget(func(state searchWidgetState) searchWidgetState {
		return state.withEditorOpened(text)
	})
}

func (program *Program) clearSearchWidgetEditor() {
	program.updateSearchWidget(func(state searchWidgetState) searchWidgetState {
		return state.withEditorCleared()
	})
}

func (program *Program) applySearchWidgetEditorIntent(intent lineEditorIntent) bool {
	applied := false
	program.updateSearchWidget(func(state searchWidgetState) searchWidgetState {
		updated, ok := state.withEditorIntentApplied(intent)
		applied = ok
		return updated
	})
	return applied
}

func (program *Program) setSearchWidgetDetailDirection(reverse bool) {
	program.updateSearchWidget(func(state searchWidgetState) searchWidgetState {
		return state.withDetailSearchDirection(reverse)
	})
}
