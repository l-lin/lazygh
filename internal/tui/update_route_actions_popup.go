package tui

func (program *Program) routeActionsPopupChromeLifecycle(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgOpenActionsPopup:
		program.clearPendingSelectionPrefix()
		program.detailState.viewState.clearPendingPrefix()
		program.clearActionsPopupPendingConfirmation()
		if program.overlayState.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
			return handledUpdate(nil)
		}
		if actual.ActionCount <= 0 {
			return handledUpdate(nil)
		}
		program.actionsPopupWidget.reactionPicker = nil
		program.actionsPopupWidget.themePicker = nil
		program.actionsPopupWidget.assigneePicker = nil
		program.actionsPopupWidget.assigneePickerLoad = nil
		program.model.OpenActionsPopup(actual.ActionCount)
		program.actionsPopupWidget.clearSearchEditor()
		program.actionsPopupWidget.errorMessage = ""
		return handledUpdate(nil)
	case MsgCloseActionsPopup:
		program.clearPendingSelectionPrefix()
		program.closeActionsPopupState()
		return handledUpdate(nil)
	case MsgFocusActionsPopupSearch:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return handledUpdate(nil)
		}
		program.model.ClearPaneSearchQueries()
		program.clearActionsPopupPendingConfirmation()
		program.actionsPopupWidget.openSearchEditor("")
		program.updateActionsPopupSearch("")
		program.model.FocusActionsPopupSearch()
		program.actionsPopupWidget.errorMessage = ""
		return handledUpdate(nil)
	case MsgFocusActionsPopupList:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return handledUpdate(nil)
		}
		program.clearActionsPopupPendingConfirmation()
		program.model.BlurActionsPopupSearch()
		return handledUpdate(nil)
	case MsgExecuteSelectedActionsPopupActionRequested:
		return handledUpdate(program.applyExecuteSelectedActionsPopupActionRequested())
	case MsgSubmitSelectedActionsPopupActionRequested:
		return handledUpdate(program.applySubmitSelectedActionsPopupActionRequested())
	case MsgActionsPopupPageRequested:
		return handledUpdate(program.applyActionsPopupPageRequested(actual))
	case MsgActionsPopupPageResolved:
		return handledUpdate(program.applyActionsPopupPageResolved(actual))
	case MsgActionsPopupViewportRequested:
		return handledUpdate(program.applyActionsPopupViewportRequested(actual))
	case MsgMoveActionsPopupSelection:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return handledUpdate(nil)
		}
		program.syncVisibleActionsPopupSearchSelection()
		program.clearActionsPopupPendingConfirmation()
		program.moveActionsPopupSelection(actual.Delta)
		program.actionsPopupWidget.errorMessage = ""
		return handledUpdate(nil)
	case MsgMoveActionsPopupSelectionToTop:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return handledUpdate(nil)
		}
		program.syncVisibleActionsPopupSearchSelection()
		program.clearActionsPopupPendingConfirmation()
		program.model.MoveActionsPopupSelectionToTop()
		program.actionsPopupWidget.errorMessage = ""
		return handledUpdate(nil)
	case MsgMoveActionsPopupSelectionToBottom:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return handledUpdate(nil)
		}
		program.syncVisibleActionsPopupSearchSelection()
		program.clearActionsPopupPendingConfirmation()
		program.model.MoveActionsPopupSelectionToBottom()
		program.actionsPopupWidget.errorMessage = ""
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) moveActionsPopupSelection(delta int) {
	if program == nil || delta == 0 {
		return
	}
	if delta > 0 {
		for range delta {
			program.model.MoveActionsPopupSelectionDown()
		}
		return
	}
	for range -delta {
		program.model.MoveActionsPopupSelectionUp()
	}
}
