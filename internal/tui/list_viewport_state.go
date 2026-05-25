package tui

func (program *Program) setPendingListViewportPlacement(viewName string, placement viewportPlacement) {
	if viewName == "" {
		return
	}
	if program.navigationState.pendingListViewportPlacements == nil {
		program.navigationState.pendingListViewportPlacements = map[string]viewportPlacement{}
	}
	program.navigationState.pendingListViewportPlacements[viewName] = placement
}

func (program *Program) pendingListViewportPlacement(viewName string) (viewportPlacement, bool) {
	if viewName == "" || len(program.navigationState.pendingListViewportPlacements) == 0 {
		return 0, false
	}

	placement, ok := program.navigationState.pendingListViewportPlacements[viewName]
	if !ok {
		return 0, false
	}
	return placement, true
}

func (program *Program) clearPendingListViewportPlacement(viewName string) {
	if viewName == "" || len(program.navigationState.pendingListViewportPlacements) == 0 {
		return
	}
	delete(program.navigationState.pendingListViewportPlacements, viewName)
}

func (program *Program) clearVisibleListViewportPlacements(layout ScreenLayout) {
	for _, frame := range layout.PanelFrames {
		if !frame.Visible {
			continue
		}
		program.clearPendingListViewportPlacement(frame.ViewName)
	}
}
