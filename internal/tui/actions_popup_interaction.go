package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) actionsPopupSelectionLineState() (int, int) {
	return program.currentActionsPopupSelectedRenderedLine(), program.currentActionsPopupRenderedLineCount()
}

func (program *Program) openActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	program.clearActionsPopupPendingConfirmation()
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}

	actions := program.currentActionsPopupActions()
	if len(actions) == 0 {
		return nil
	}

	program.reactionPicker = nil
	program.themePicker = nil
	program.assigneePicker = nil
	program.assigneePickerLoad = nil
	program.model.OpenActionsPopup(len(actions))
	program.actionsPopupSearchEditor = nil
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) closeActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.model.CloseActionsPopup()
	program.actionsPopupSearchEditor = nil
	program.clearActionsPopupPendingConfirmation()
	program.actionsPopupErrorMessage = ""
	program.reactionPicker = nil
	program.themePicker = nil
	program.assigneePicker = nil
	program.assigneePickerLoad = nil
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) focusActionsPopupSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.model.ClearPaneSearchQueries()
	program.clearActionsPopupPendingConfirmation()
	program.actionsPopupSearchEditor = newLineEditor("")
	program.updateActionsPopupSearch("")
	program.model.FocusActionsPopupSearch()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) focusActionsPopupList(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	program.model.BlurActionsPopupSearch()
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionDown(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	program.model.MoveActionsPopupSelectionDown()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionUp(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	program.model.MoveActionsPopupSelectionUp()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) pageActionsPopupDown(gui *gocui.Gui, view *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	actualView := program.resolveView(gui, view, viewActionsPopupName)
	program.clearActionsPopupPendingConfirmation()
	program.model.PageActionsPopupDown(viewPageSize(actualView))
	program.actionsPopupErrorMessage = ""
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
	program.actionsPopupErrorMessage = ""
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
	program.actionsPopupErrorMessage = ""
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
	program.actionsPopupErrorMessage = ""
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
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	program.model.MoveActionsPopupSelectionToTop()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionToBottom(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	program.model.MoveActionsPopupSelectionToBottom()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
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
		return program.handleActionsPopupActionResult(gui, action.execute(gui))
	}

	action, ok := program.selectedActionsPopupAction()
	if !ok {
		return nil
	}
	return program.handleActionsPopupActionResult(gui, action.execute(gui))
}

func (program *Program) submitSelectedActionsPopupAction(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	if !program.assigneePickerVisible() {
		return program.executeSelectedActionsPopupAction(gui, nil)
	}
	return program.handleActionsPopupActionResult(gui, program.executeSubmitAssigneePickerAction(gui))
}

func (program *Program) handleActionsPopupActionResult(gui *gocui.Gui, result actionsPopupActionResult) error {
	if result.err != nil {
		if message, ok := transientErrorPopupActionMessage(result.err); ok {
			program.actionsPopupErrorMessage = ""
			program.reportError(gui, message)
			if gui == nil {
				return nil
			}
			return program.refreshViews(gui)
		}
		if message := strings.TrimSpace(result.feedbackMessage); message != "" {
			program.actionsPopupErrorMessage = ""
			program.setFeedback(result.feedbackTarget, message)
			if gui == nil {
				return nil
			}
			return program.refreshViews(gui)
		}
		program.actionsPopupErrorMessage = strings.TrimSpace(result.err.Error())
		program.reportError(gui, program.actionsPopupErrorMessage)
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}

	if result.closePopup {
		return program.closeActionsPopup(gui, nil)
	}
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) editActionsPopupSearch(view *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if key == gocui.KeyEnter || key == gocui.KeyEsc {
		return false
	}
	if program.actionsPopupSearchEditor == nil {
		program.actionsPopupSearchEditor = newLineEditor(program.model.ActionsPopupSearchQuery())
	}
	if !program.actionsPopupSearchEditor.HandleKey(key, ch, mod) {
		return false
	}

	program.clearActionsPopupPendingConfirmation()
	query := program.actionsPopupSearchEditor.Text()
	requestID := 0
	if program.assigneePickerVisible() {
		requestID = program.resetAssigneePickerSearch(query)
	}
	program.updateActionsPopupSearch(query)
	if program.assigneePickerVisible() {
		program.queueAssigneePickerSearch(program.gui, requestID, query)
	}
	program.actionsPopupErrorMessage = ""
	if program.gui != nil {
		_ = program.refreshViews(program.gui)
		return true
	}

	program.configureActionsPopupSearchView(view)
	program.renderActionsPopupSearchView(view)
	return true
}
