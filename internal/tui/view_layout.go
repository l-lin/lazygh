package tui

import "github.com/jesseduffield/gocui"

func deleteViewIfPresent(gui *gocui.Gui, viewName string) error {
	if gui == nil {
		return nil
	}

	if err := gui.DeleteView(viewName); err != nil && !isUnknownViewError(err) {
		return err
	}

	return nil
}

func deleteViewsIfPresent(gui *gocui.Gui, viewNames ...string) error {
	for _, viewName := range viewNames {
		if err := deleteViewIfPresent(gui, viewName); err != nil {
			return err
		}
	}

	return nil
}
