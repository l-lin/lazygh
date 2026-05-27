package tui

import "github.com/jesseduffield/gocui"

func (program *Program) captureGUI(gui *gocui.Gui) *gocui.Gui {
	if program == nil {
		return gui
	}
	if gui != nil {
		program.gui = gui
	}
	return program.gui
}

func (program *Program) dispatch(gui *gocui.Gui, msg Msg) error {
	gui = program.captureGUI(gui)
	program.executeCmds(gui, Update(program, msg))
	return program.afterStateChange(gui)
}

func (program *Program) dispatchAsyncMessage(msg Msg) {
	if program == nil || msg == nil {
		return
	}
	gui := program.captureGUI(nil)
	if gui == nil || program.uiUpdater == nil {
		_ = program.dispatch(gui, msg)
		return
	}
	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		return program.dispatch(gui, msg)
	})
}
