package tui

func (program *Program) updateModalEditorState(transition func(modalEditorState) modalEditorState) {
	if program == nil {
		return
	}
	program.overlayState.modalEditor = transition(program.overlayState.modalEditor)
}

func (program *Program) setModalEditorErrorMessage(message string) {
	program.updateModalEditorState(func(state modalEditorState) modalEditorState {
		return state.withErrorMessage(message)
	})
}

func (program *Program) clearModalEditorErrorMessage() {
	program.updateModalEditorState(func(state modalEditorState) modalEditorState {
		return state.withoutErrorMessage()
	})
}
