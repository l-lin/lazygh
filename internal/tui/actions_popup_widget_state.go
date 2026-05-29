package tui

import "strings"

func newActionsPopupWidgetState() actionsPopupWidgetState {
	return actionsPopupWidgetState{assigneePickerSearchDebounceDelay: defaultAssigneePickerSearchDebounceDelay}
}

func (state actionsPopupWidgetState) withErrorMessage(message string) actionsPopupWidgetState {
	state.errorMessage = strings.TrimSpace(message)
	return state
}

func (state actionsPopupWidgetState) withoutErrorMessage() actionsPopupWidgetState {
	state.errorMessage = ""
	return state
}

func (state actionsPopupWidgetState) withPendingConfirmation(actionID string) actionsPopupWidgetState {
	state.pendingConfirmationActionID = strings.TrimSpace(actionID)
	state.errorMessage = ""
	return state
}

func (state actionsPopupWidgetState) withPendingConfirmationCleared() actionsPopupWidgetState {
	state.pendingConfirmationActionID = ""
	return state
}

func (state actionsPopupWidgetState) withSearchEditorCleared() actionsPopupWidgetState {
	state.searchEditor = lineEditor{}
	state.searchEditorVisible = false
	return state
}

func (state actionsPopupWidgetState) withReactionPickerOpened(target pullRequestReactionActionTarget) actionsPopupWidgetState {
	state = state.withSearchEditorCleared()
	state.pendingConfirmationActionID = ""
	state.errorMessage = ""
	state.reactionPicker = &reactionPickerState{target: target}
	state.themePicker = nil
	state.assigneePicker = nil
	state.assigneePickerLoad = nil
	return state
}

func (state actionsPopupWidgetState) withThemePickerOpened() actionsPopupWidgetState {
	state = state.withSearchEditorCleared()
	state.pendingConfirmationActionID = ""
	state.errorMessage = ""
	state.reactionPicker = nil
	state.themePicker = &themePickerState{}
	state.assigneePicker = nil
	state.assigneePickerLoad = nil
	return state
}

func (state actionsPopupWidgetState) withAssigneePickerOpened(target pullRequestAssigneePickerTarget, viewerLogin string, viewerName string) actionsPopupWidgetState {
	state = state.withSearchEditorCleared()
	state.pendingConfirmationActionID = ""
	state.errorMessage = ""
	state.reactionPicker = nil
	state.themePicker = nil
	state.assigneePicker = newAssigneePickerState(target, viewerLogin, viewerName)
	state.assigneePickerLoad = nil
	return state
}

func (state actionsPopupWidgetState) withAssigneePickerState(picker *assigneePickerState) actionsPopupWidgetState {
	state.assigneePicker = picker
	return state
}

func (state actionsPopupWidgetState) withPopupClosed() actionsPopupWidgetState {
	state = state.withSearchEditorCleared()
	state.pendingConfirmationActionID = ""
	state.errorMessage = ""
	state.reactionPicker = nil
	state.themePicker = nil
	state.assigneePicker = nil
	state.assigneePickerLoad = nil
	return state
}
