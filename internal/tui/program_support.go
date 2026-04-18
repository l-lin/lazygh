package tui

import "github.com/jesseduffield/gocui"

type asyncRunner interface {
	Go(run func())
}

type goroutineAsyncRunner struct{}

func (goroutineAsyncRunner) Go(run func()) {
	go run()
}

type uiUpdater interface {
	Apply(gui *gocui.Gui, update func(*gocui.Gui) error)
}

type queuedUIUpdater struct{}

func (queuedUIUpdater) Apply(gui *gocui.Gui, update func(*gocui.Gui) error) {
	gui.Update(update)
}
