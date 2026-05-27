package tui

import "github.com/jesseduffield/gocui"

func (program *Program) refreshViewsIfGUI(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	return program.afterStateChange(gui)
}

func (program *Program) refreshShell(gui *gocui.Gui) error {
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) refreshDetailView(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}
	if actualErr := program.refreshExistingView(gui, viewDetailName, program.configureDetailView, program.renderDetailView); actualErr != nil {
		return actualErr
	}

	return program.syncShellState(gui)
}

func (program *Program) applyScreenCompositionAndSyncView(gui *gocui.Gui, composition screenComposition) error {
	if actualErr := program.applyScreenComposition(gui, composition); actualErr != nil {
		return actualErr
	}
	return program.syncCurrentView(gui)
}

func (program *Program) mutateDetailViewState(gui *gocui.Gui, view *gocui.View, mutate func(detailDocument, int)) error {
	if actualErr := program.mutateDetailViewStateWithoutRefresh(gui, view, mutate); actualErr != nil {
		return actualErr
	}

	return program.refreshDetailView(gui)
}

func (program *Program) mutateDetailViewStateWithoutRefresh(gui *gocui.Gui, view *gocui.View, mutate func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	actualView := view
	if actualView == nil && gui != nil {
		if detailView, actualErr := gui.View(viewDetailName); actualErr == nil {
			actualView = detailView
		}
	}

	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)
	mutate(detailDocument, viewportHeight)
	program.syncDetailViewState(detailDocument, viewportHeight)
	return nil
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
