package tui

import "github.com/jesseduffield/gocui"

func (program *Program) dispatch(gui *gocui.Gui, msg Msg) error {
	if program != nil {
		program.gui = gui
	}
	program.executeCmds(gui, Update(program, msg))
	return program.afterStateChange(gui)
}

func (program *Program) dispatchAsync(gui *gocui.Gui, msg Msg) {
	if program == nil || msg == nil {
		return
	}
	if gui == nil || program.uiUpdater == nil {
		_ = program.dispatch(gui, msg)
		return
	}
	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		return program.dispatch(gui, msg)
	})
}
