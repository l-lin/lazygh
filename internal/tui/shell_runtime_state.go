package tui

func newListViewportRuntimeState() listViewportRuntimeState {
	return listViewportRuntimeState{pendingPlacements: map[string]viewportPlacement{}}
}

func (state listViewportRuntimeState) withPendingPlacement(viewName string, placement viewportPlacement) listViewportRuntimeState {
	if viewName == "" {
		return state
	}
	updated := cloneListViewportPlacements(state.pendingPlacements)
	updated[viewName] = placement
	state.pendingPlacements = updated
	return state
}

func (state listViewportRuntimeState) pendingPlacement(viewName string) (viewportPlacement, bool) {
	if viewName == "" || len(state.pendingPlacements) == 0 {
		return 0, false
	}
	placement, ok := state.pendingPlacements[viewName]
	if !ok {
		return 0, false
	}
	return placement, true
}

func (state listViewportRuntimeState) withoutPendingPlacement(viewName string) listViewportRuntimeState {
	if viewName == "" || len(state.pendingPlacements) == 0 {
		return state
	}
	if _, ok := state.pendingPlacements[viewName]; !ok {
		return state
	}
	updated := cloneListViewportPlacements(state.pendingPlacements)
	delete(updated, viewName)
	state.pendingPlacements = updated
	return state
}

func cloneListViewportPlacements(placements map[string]viewportPlacement) map[string]viewportPlacement {
	if len(placements) == 0 {
		return map[string]viewportPlacement{}
	}
	copied := make(map[string]viewportPlacement, len(placements))
	for viewName, placement := range placements {
		copied[viewName] = placement
	}
	return copied
}

func (state keybindingRuntimeState) withRegisteredFingerprint(fingerprint string) keybindingRuntimeState {
	state.registeredFingerprint = fingerprint
	return state
}

func (state keybindingRuntimeState) registeredFingerprintValue() string {
	return state.registeredFingerprint
}

func (program *Program) updateListViewportRuntime(update func(listViewportRuntimeState) listViewportRuntimeState) {
	if program == nil || update == nil {
		return
	}
	program.listViewportRuntime = update(program.listViewportRuntime)
}

func (program *Program) updateKeybindingRuntime(update func(keybindingRuntimeState) keybindingRuntimeState) {
	if program == nil || update == nil {
		return
	}
	program.keybindingRuntime = update(program.keybindingRuntime)
}

func (program *Program) registeredKeybindingFingerprint() string {
	if program == nil {
		return ""
	}
	return program.keybindingRuntime.registeredFingerprintValue()
}
