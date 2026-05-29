package tui

func (program *Program) updateModalEditorState(transition func(modalEditorState) modalEditorState) {
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		return state.withModalEditor(transition(state.modalEditor))
	})
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
