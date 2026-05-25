package tui

import "github.com/jesseduffield/gocui"

func (program *Program) start(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	program.configureGUI(gui)
	gui.SetManagerFunc(program.layout)
	if err := program.setKeybindings(gui); err != nil {
		return err
	}
	return program.refreshViews(gui)
}

func (program *Program) syncShellState(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}
	return program.reloadRegisteredKeybindings(gui)
}
