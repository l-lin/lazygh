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

func newPullRequestRefreshDispatcher(program *Program, gui *gocui.Gui) func(Msg) bool {
	if program == nil || gui == nil || program.uiUpdater == nil {
		return nil
	}
	updater := program.uiUpdater
	dispatch := program.dispatch
	return func(message Msg) bool {
		if message == nil {
			return false
		}
		update := func(gui *gocui.Gui) error {
			return dispatch(gui, message)
		}
		if tryUpdater, ok := updater.(interface {
			TryApply(*gocui.Gui, func(*gocui.Gui) error) bool
		}); ok {
			return tryUpdater.TryApply(gui, update)
		}
		updater.Apply(gui, update)
		return true
	}
}
