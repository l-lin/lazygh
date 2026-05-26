package tui

import "github.com/jesseduffield/gocui"

func (program *Program) nextDetailSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgRepeatDetailSearchRequested{View: view, Direction: searchRepeatForward})
}

func (program *Program) previousDetailSearchMatch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgRepeatDetailSearchRequested{View: view, Direction: searchRepeatBackward})
}
