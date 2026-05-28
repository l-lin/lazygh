package tui

import "github.com/jesseduffield/gocui"

func (program *Program) dispatchRuntimeMessage(msg Msg) error {
	return program.applyRuntimeMessage(nil, msg, true)
}

func (program *Program) executeRuntimeMessage(gui *gocui.Gui, msg Msg) error {
	return program.applyRuntimeMessage(gui, msg, false)
}

func (program *Program) applyRuntimeMessage(gui *gocui.Gui, msg Msg, runAfterStateChange bool) error {
	if program == nil || msg == nil {
		return nil
	}
	gui = program.captureGUI(gui)
	program.executeCmds(gui, Update(program, msg))
	if !runAfterStateChange {
		return nil
	}
	return program.afterStateChange(gui)
}
