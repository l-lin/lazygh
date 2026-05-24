package tui

import "github.com/jesseduffield/gocui"

func (program *Program) dispatch(gui *gocui.Gui, msg Msg) error {
	program.executeCmds(gui, Update(program, msg))
	return program.afterStateChange(gui)
}
