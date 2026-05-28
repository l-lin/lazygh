package tui

import "github.com/jesseduffield/gocui"

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
	return program.dispatch(gui, MsgRepeatActionsPopupSearch{Direction: searchRepeatForward})
}

func (program *Program) previousActionsPopupSearchMatch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgRepeatActionsPopupSearch{Direction: searchRepeatBackward})
}
