package tui

func (program *Program) updateOverlayState(transition func(overlayStateModel) overlayStateModel) {
	if program == nil {
		return
	}
	program.overlayState = transition(program.overlayState)
}

func (program *Program) setHelpVisible(visible bool) {
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		return state.withHelpVisible(visible)
	})
}

func (program *Program) setTransientErrorPopupState(popup transientErrorPopupState) {
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		return state.withTransientErrorPopup(popup)
	})
}

func (program *Program) clearTransientErrorPopupState() {
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		return state.withClearedTransientErrorPopup()
	})
}

func (program *Program) setModalEditorState(state modalEditorState) {
	program.updateOverlayState(func(overlay overlayStateModel) overlayStateModel {
		return overlay.withModalEditor(state)
	})
}

func (program *Program) clearModalEditorState() {
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		return state.withClearedModalEditor()
	})
}
