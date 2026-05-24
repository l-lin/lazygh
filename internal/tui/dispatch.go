package tui

import "github.com/jesseduffield/gocui"

func (program *Program) dispatch(gui *gocui.Gui, msg Msg) error {
	program.executeCmds(gui, Update(program, msg))
	if gui == nil {
		return nil
	}
	return program.layout(gui)
}
