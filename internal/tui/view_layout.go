package tui

import "github.com/jesseduffield/gocui"

func setPaneView(gui *gocui.Gui, viewName string, visible bool, frame paneFrame) (*gocui.View, error) {
	if !visible {
		return nil, deleteViewIfPresent(gui, viewName)
	}

	view, err := gui.SetView(viewName, frame.x0, frame.y0, frame.x1, frame.y1, 0)
	if err != nil && !isUnknownViewError(err) {
		return nil, err
	}

	return view, nil
}

func deleteViewIfPresent(gui *gocui.Gui, viewName string) error {
	if gui == nil {
		return nil
	}

	if err := gui.DeleteView(viewName); err != nil && !isUnknownViewError(err) {
		return err
	}

	return nil
}
