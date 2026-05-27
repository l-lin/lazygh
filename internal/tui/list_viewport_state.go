package tui

func (program *Program) setPendingListViewportPlacement(viewName string, placement viewportPlacement) {
	if viewName == "" {
		return
	}
	if program.listViewportRuntime.pendingPlacements == nil {
		program.listViewportRuntime.pendingPlacements = map[string]viewportPlacement{}
	}
	program.listViewportRuntime.pendingPlacements[viewName] = placement
}

func (program *Program) pendingListViewportPlacement(viewName string) (viewportPlacement, bool) {
	if viewName == "" || len(program.listViewportRuntime.pendingPlacements) == 0 {
		return 0, false
	}

	placement, ok := program.listViewportRuntime.pendingPlacements[viewName]
	if !ok {
		return 0, false
	}
	return placement, true
}

func (program *Program) clearPendingListViewportPlacement(viewName string) {
	if viewName == "" || len(program.listViewportRuntime.pendingPlacements) == 0 {
		return
	}
	delete(program.listViewportRuntime.pendingPlacements, viewName)
}

func (program *Program) clearVisibleListViewportPlacements(layout ScreenLayout) {
	for _, frame := range layout.PanelFrames {
		if !frame.Visible {
			continue
		}
		program.clearPendingListViewportPlacement(frame.ViewName)
	}
}
