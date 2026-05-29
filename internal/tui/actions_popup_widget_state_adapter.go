package tui

func (program *Program) updateActionsPopupWidgetState(transition func(actionsPopupWidgetState) actionsPopupWidgetState) {
	if program == nil {
		return
	}
	program.actionsPopupWidget = transition(program.actionsPopupWidget)
}

func (program *Program) openActionsPopupSearchEditor(text string) {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withSearchEditorOpened(text)
	})
}

func (program *Program) clearActionsPopupSearchEditor() {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withSearchEditorCleared()
	})
}

func (program *Program) applyActionsPopupSearchEditorIntent(intent lineEditorIntent) bool {
	applied := false
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		updated, ok := state.withSearchEditorIntentApplied(intent)
		applied = ok
		return updated
	})
	return applied
}

func (program *Program) setActionsPopupErrorMessage(message string) {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withErrorMessage(message)
	})
}

func (program *Program) clearActionsPopupErrorMessage() {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withoutErrorMessage()
	})
}

func (program *Program) setActionsPopupPendingConfirmation(actionID string) {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withPendingConfirmation(actionID)
	})
}

func (program *Program) clearActionsPopupPendingConfirmation() {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withPendingConfirmationCleared()
	})
}

func (program *Program) resetActionsPopupWidgetChrome() {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withPopupClosed()
	})
}

func (program *Program) openReactionPicker(target pullRequestReactionActionTarget) {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withReactionPickerOpened(target)
	})
}

func (program *Program) openThemePicker() {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withThemePickerOpened()
	})
}

func (program *Program) openAssigneePicker(target pullRequestAssigneePickerTarget) {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withAssigneePickerOpened(target, program.currentConnectedUserLogin(), program.currentConnectedUserName())
	})
}

func (program *Program) setActionsPopupAssigneePickerState(picker *assigneePickerState) {
	program.updateActionsPopupWidgetState(func(state actionsPopupWidgetState) actionsPopupWidgetState {
		return state.withAssigneePickerState(picker)
	})
}
