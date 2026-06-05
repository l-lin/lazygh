package tui

import "github.com/jesseduffield/gocui"

func (program *Program) refreshViewsIfGUI(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	return program.afterStateChange(gui)
}

func (program *Program) refreshCurrentViewFocus(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	return program.syncCurrentView(gui)
}

func (program *Program) applyScreenCompositionAndSyncView(gui *gocui.Gui, composition screenComposition) error {
	if actualErr := program.applyScreenComposition(gui, composition); actualErr != nil {
		return actualErr
	}
	return program.syncCurrentView(gui)
}

func (program *Program) recenterListSelection(gui *gocui.Gui, view *gocui.View, fallbackName string, selectedVisibleLine int, lineCount int) error {
	return program.placeListSelection(gui, view, fallbackName, selectedVisibleLine, lineCount, viewportPlacementCenter)
}

func (program *Program) placeListSelection(gui *gocui.Gui, view *gocui.View, fallbackName string, selectedVisibleLine int, lineCount int, placement viewportPlacement) error {
	if lineCount < 1 {
		return nil
	}

	actualView := program.resolveView(gui, view, fallbackName)
	viewName := fallbackName
	if actualView != nil && actualView.Name() != "" {
		viewName = actualView.Name()
	}
	program.setPendingListViewportPlacement(viewName, placement)
	program.placeListLine(actualView, selectedVisibleLine, lineCount, placement)
	return program.refreshViewsIfGUI(gui)
}
