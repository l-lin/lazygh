package tui

import "github.com/jesseduffield/gocui"

func (program *Program) actionsPopupSelectionLineState() (int, int) {
	return program.currentActionsPopupSelectedRenderedLine(), program.currentActionsPopupRenderedLineCount()
}

func (program *Program) openActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	actions := program.currentActionsPopupActions()
	return program.dispatch(gui, MsgOpenActionsPopup{ActionCount: len(actions)})
}

func (program *Program) closeActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgCloseActionsPopup{})
}

func (program *Program) focusActionsPopupSearch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgFocusActionsPopupSearch{})
}

func (program *Program) focusActionsPopupList(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgFocusActionsPopupList{})
}

func (program *Program) moveActionsPopupSelectionDown(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveActionsPopupSelection{Delta: 1})
}

func (program *Program) moveActionsPopupSelectionUp(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveActionsPopupSelection{Delta: -1})
}

func (program *Program) pageActionsPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	actualView := program.resolveView(gui, view, viewActionsPopupName)
	program.clearActionsPopupPendingConfirmation()
	program.model.PageActionsPopupDown(viewPageSize(actualView))
	program.actionsPopupWidget.errorMessage = ""
	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.recenterListSelection(gui, actualView, viewActionsPopupName, selectedLine, lineCount)
}

func (program *Program) pageActionsPopupUp(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	actualView := program.resolveView(gui, view, viewActionsPopupName)
	program.clearActionsPopupPendingConfirmation()
	program.model.PageActionsPopupUp(viewPageSize(actualView))
	program.actionsPopupWidget.errorMessage = ""
	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.recenterListSelection(gui, actualView, viewActionsPopupName, selectedLine, lineCount)
}

func (program *Program) fullPageActionsPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	actualView := program.resolveView(gui, view, viewActionsPopupName)
	program.clearActionsPopupPendingConfirmation()
	program.model.FullPageActionsPopupDown(viewPageSize(actualView))
	program.actionsPopupWidget.errorMessage = ""
	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.recenterListSelection(gui, actualView, viewActionsPopupName, selectedLine, lineCount)
}

func (program *Program) fullPageActionsPopupUp(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	actualView := program.resolveView(gui, view, viewActionsPopupName)
	program.clearActionsPopupPendingConfirmation()
	program.model.FullPageActionsPopupUp(viewPageSize(actualView))
	program.actionsPopupWidget.errorMessage = ""
	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.recenterListSelection(gui, actualView, viewActionsPopupName, selectedLine, lineCount)
}

func (program *Program) recenterActionsPopupSelection(gui *gocui.Gui, view *gocui.View) error {
	if !program.model.ActionsPopupVisible() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.recenterListSelection(gui, view, viewActionsPopupName, selectedLine, lineCount)
}

func (program *Program) moveActionsPopupSelectionToViewportTop(gui *gocui.Gui, view *gocui.View) error {
	if !program.model.ActionsPopupVisible() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.placeListSelection(gui, view, viewActionsPopupName, selectedLine, lineCount, viewportPlacementTop)
}

func (program *Program) moveActionsPopupSelectionToViewportBottom(gui *gocui.Gui, view *gocui.View) error {
	if !program.model.ActionsPopupVisible() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	selectedLine, lineCount := program.actionsPopupSelectionLineState()
	return program.placeListSelection(gui, view, viewActionsPopupName, selectedLine, lineCount, viewportPlacementBottom)
}

func (program *Program) moveActionsPopupSelectionToTop(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveActionsPopupSelectionToTop{})
}

func (program *Program) moveActionsPopupSelectionToBottom(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgMoveActionsPopupSelectionToBottom{})
}

func (program *Program) executeSelectedActionsPopupAction(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.ghCommandLoading() {
		return nil
	}

	if program.assigneePickerVisible() {
		action, ok := program.selectedActionsPopupAction()
		if !ok {
			return nil
		}
		return program.handleActionsPopupActionError(gui, action.execute(gui))
	}

	action, ok := program.selectedActionsPopupAction()
	if !ok {
		return nil
	}
	return program.handleActionsPopupActionError(gui, action.execute(gui))
}

func (program *Program) submitSelectedActionsPopupAction(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	if !program.assigneePickerVisible() {
		return program.executeSelectedActionsPopupAction(gui, nil)
	}
	return program.handleActionsPopupActionError(gui, program.executeSubmitAssigneePickerAction(gui))
}

func (program *Program) handleActionsPopupActionError(gui *gocui.Gui, err error) error {
	return program.dispatch(gui, MsgActionsPopupActionErrorHandled{Err: err})
}

func (program *Program) editActionsPopupSearch(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyEnter || key == gocui.KeyEsc {
		return false
	}
	if program.actionsPopupWidget.searchEditor == nil {
		program.actionsPopupWidget.searchEditor = newLineEditor(program.model.ActionsPopupSearchQuery())
	}
	if !program.actionsPopupWidget.searchEditor.HandleKey(key, ch, mod) {
		return false
	}

	return program.dispatchEditorMessage(MsgActionsPopupSearchEdited{Query: program.actionsPopupWidget.searchEditor.Text()})
}
