package tui

import "time"

type searchWidgetState struct {
	editor         lineEditor
	editorVisible  bool
	detailReversed bool
}

func (state *searchWidgetState) openEditor(text string) {
	if state == nil {
		return
	}
	state.editor = newLineEditor(text)
	state.editorVisible = true
}

func (state *searchWidgetState) clearEditor() {
	if state == nil {
		return
	}
	state.editor = lineEditor{}
	state.editorVisible = false
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
