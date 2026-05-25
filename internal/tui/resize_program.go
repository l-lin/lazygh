package tui

import "github.com/jesseduffield/gocui"

func (program *Program) growFocusedPane(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}

	return program.dispatch(gui, MsgAdjustFocusedPane{Delta: 1})
}

func (program *Program) shrinkFocusedPane(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}

	return program.dispatch(gui, MsgAdjustFocusedPane{Delta: -1})
}
