package tui

import "github.com/jesseduffield/gocui"

func (program *Program) growFocusedPane(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}

	program.model.GrowFocusedPane()
	return program.layout(gui)
}

func (program *Program) shrinkFocusedPane(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return nil
	}

	program.model.ShrinkFocusedPane()
	return program.layout(gui)
}
