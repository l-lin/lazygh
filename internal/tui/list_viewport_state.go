package tui

func (program *Program) setPendingListViewportPlacement(viewName string, placement viewportPlacement) {
	program.updateListViewportRuntime(func(state listViewportRuntimeState) listViewportRuntimeState {
		return state.withPendingPlacement(viewName, placement)
	})
}

func (program *Program) pendingListViewportPlacement(viewName string) (viewportPlacement, bool) {
	if program == nil {
		return 0, false
	}
	return program.listViewportRuntime.pendingPlacement(viewName)
}

func (program *Program) clearPendingListViewportPlacement(viewName string) {
	program.updateListViewportRuntime(func(state listViewportRuntimeState) listViewportRuntimeState {
		return state.withoutPendingPlacement(viewName)
	})
}

func (program *Program) clearVisibleListViewportPlacements(layout ScreenLayout) {
	for _, frame := range layout.PanelFrames {
		if !frame.Visible {
			continue
		}
		program.clearPendingListViewportPlacement(frame.ViewName)
	}
}
