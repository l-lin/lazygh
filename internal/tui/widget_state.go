package tui

import "time"

type searchWidgetState struct {
	editor         lineEditor
	editorVisible  bool
	detailReversed bool
}

func (state searchWidgetState) withEditorOpened(text string) searchWidgetState {
	state.editor = newLineEditor(text)
	state.editorVisible = true
	return state
}

func (state searchWidgetState) withEditorCleared() searchWidgetState {
	state.editor = lineEditor{}
	state.editorVisible = false
	return state
}

func (state searchWidgetState) withEditorIntentApplied(intent lineEditorIntent) (searchWidgetState, bool) {
	if !state.editorVisible {
		return state, false
	}

	updated := state
	if !updated.editor.ApplyIntent(intent) {
		return state, false
	}
	return updated, true
}

func (state searchWidgetState) withDetailSearchDirection(reverse bool) searchWidgetState {
	state.detailReversed = reverse
	return state
}

func (state searchWidgetState) hasEditor() bool {
	return state.editorVisible
}

type actionsPopupWidgetState struct {
	searchEditor                      lineEditor
	searchEditorVisible               bool
	errorMessage                      string
	pendingConfirmationActionID       string
	reactionPicker                    *reactionPickerState
	themePicker                       *themePickerState
	assigneePicker                    *assigneePickerState
	assigneePickerSearchDebounceDelay time.Duration
	assigneePickerLoad                *assigneePickerLoadState
}

func (state *actionsPopupWidgetState) openSearchEditor(text string) {
	if state == nil {
		return
	}
	state.searchEditor = newLineEditor(text)
	state.searchEditorVisible = true
}

func (state *actionsPopupWidgetState) clearSearchEditor() {
	if state == nil {
		return
	}
	state.searchEditor = lineEditor{}
	state.searchEditorVisible = false
}

func (state actionsPopupWidgetState) hasSearchEditor() bool {
	return state.searchEditorVisible
}
