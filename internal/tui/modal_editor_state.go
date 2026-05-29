package tui

import "strings"

func (state modalEditorState) withErrorMessage(message string) modalEditorState {
	state.errorMessage = strings.TrimSpace(message)
	return state
}

func (state modalEditorState) withoutErrorMessage() modalEditorState {
	state.errorMessage = ""
	return state
}

func (state modalEditorState) withLineEditorIntentApplied(intent lineEditorIntent) (modalEditorState, bool) {
	updated := state.clone()
	if !updated.applyLineEditorIntent(intent) {
		return state, false
	}
	updated.errorMessage = ""
	return updated, true
}

func (state modalEditorState) withMultilineEditorIntentApplied(intent multilineEditorIntent) (modalEditorState, bool) {
	updated := state.clone()
	if !updated.applyMultilineEditorIntent(intent) {
		return state, false
	}
	updated.errorMessage = ""
	return updated, true
}

func (state modalEditorState) withTextFromExternalEditor(text string) modalEditorState {
	if !state.visible() {
		return state
	}

	updated := state.clone()
	if updated.isLineEditor() {
		updated.lineEditor.SetText(normalizeSingleLineExternalEditorText(text))
	} else {
		updated.editor.SetText(text)
	}
	updated.errorMessage = ""
	return updated
}
