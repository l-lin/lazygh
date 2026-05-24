package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (model *Model) followActionsPopupSearchMatch(choose searchMatchIndexChooser) bool {
	if !model.actionsPopup.visible {
		return false
	}

	matchIndexes := model.actionsPopup.filteredActionIndexes
	matchIndex := choose(matchIndexes, model.actionsPopup.selectedActionIndex)
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return false
	}

	model.actionsPopup.selectedActionIndex = matchIndexes[matchIndex]
	return true
}

func (program *Program) nextActionsPopupSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.repeatActionsPopupSearch(gui, searchMatchIndexAfter)
}

func (program *Program) previousActionsPopupSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.repeatActionsPopupSearch(gui, searchMatchIndexBefore)
}

func (program *Program) repeatActionsPopupSearch(gui *gocui.Gui, choose searchMatchIndexChooser) error {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return nil
	}
	if strings.TrimSpace(program.model.ActionsPopupSearchQuery()) == "" {
		return nil
	}

	program.clearActionsPopupPendingConfirmation()
	if !program.model.followActionsPopupSearchMatch(choose) {
		return nil
	}

	program.actionsPopupWidget.errorMessage = ""
	return program.refreshViewsIfGUI(gui)
}
