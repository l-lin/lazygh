package tui

import "time"

type searchWidgetPromptKind int

const (
	searchWidgetPromptNone searchWidgetPromptKind = iota
	searchWidgetPromptSearch
	searchWidgetPromptCommand
)

type searchWidgetState struct {
	editor         lineEditor
	editorVisible  bool
	promptKind     searchWidgetPromptKind
	detailReversed bool
}

func (state searchWidgetState) withEditorOpened(text string) searchWidgetState {
	state.editor = newLineEditor(text)
	state.editorVisible = true
	state.promptKind = searchWidgetPromptSearch
	return state
}

func (state searchWidgetState) withCommandModeOpened() searchWidgetState {
	state.editor = newLineEditor("")
	state.editorVisible = true
	state.promptKind = searchWidgetPromptCommand
	state.detailReversed = false
	return state
}

func (state searchWidgetState) withEditorCleared() searchWidgetState {
	state.editor = lineEditor{}
	state.editorVisible = false
	state.promptKind = searchWidgetPromptNone
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

func (state searchWidgetState) commandModeActive() bool {
	return state.editorVisible && state.promptKind == searchWidgetPromptCommand
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

func (state actionsPopupWidgetState) hasSearchEditor() bool {
	return state.searchEditorVisible
}
