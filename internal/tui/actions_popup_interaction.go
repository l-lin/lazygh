package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) openActionsPopup(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return nil
	}

	actions := program.currentActionsPopupActions()
	if len(actions) == 0 {
		return nil
	}

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
	program.actionsPopupErrorMessage = ""
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

	program.model.BlurActionsPopupSearch()
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionDown(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return nil
	}

	program.model.MoveActionsPopupSelectionDown()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionUp(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return nil
	}

	program.model.MoveActionsPopupSelectionUp()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) moveActionsPopupSelectionToTop(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	return program.armOrHandleGoToTopPrefix(viewActionsPopupName, func() error {
		program.model.MoveActionsPopupSelectionToTop()
		program.actionsPopupErrorMessage = ""
		if gui == nil {
			return nil
		}

		return program.refreshViews(gui)
	})
}

func (program *Program) moveActionsPopupSelectionToBottom(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return nil
	}

	program.model.MoveActionsPopupSelectionToBottom()
	program.actionsPopupErrorMessage = ""
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) executeSelectedActionsPopupAction(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}

	action, ok := program.selectedActionsPopupAction()
	if !ok {
		return nil
	}

	result := action.execute(gui)
	if result.err != nil {
		program.actionsPopupErrorMessage = strings.TrimSpace(result.err.Error())
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
	if key == gocui.KeyEnter || key == gocui.KeyEsc || key == gocui.KeyCtrlLsqBracket {
		return false
	}
	if program.actionsPopupSearchEditor == nil {
		program.actionsPopupSearchEditor = newLineEditor(program.model.ActionsPopupSearchQuery())
	}
	if !program.actionsPopupSearchEditor.HandleKey(key, ch, mod) {
		return false
	}

	program.updateActionsPopupSearch(program.actionsPopupSearchEditor.Text())
	program.actionsPopupErrorMessage = ""
	if program.gui != nil {
		_ = program.refreshViews(program.gui)
		return true
	}

	program.configureActionsPopupSearchView(view)
	program.renderActionsPopupSearchView(view)
	return true
}
