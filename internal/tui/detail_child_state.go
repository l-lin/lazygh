package tui

func (state detailStateModel) withWrapWidth(width int) detailStateModel {
	state.wrapWidth = width
	return state
}

func (state detailStateModel) withActiveTab(tab DetailTab) detailStateModel {
	state.activeTab = tab
	return state
}

func (state detailStateModel) withProjectedScreenStateApplication(application projectedScreenStateApplication) detailStateModel {
	if !application.hasDetailTab {
		return state
	}
	return state.withActiveTab(application.activeDetailTab)
}

func (state detailStateModel) withAdvancedActiveTab(delta int, count int) detailStateModel {
	if count <= 0 || delta == 0 {
		return state
	}

	index := int(state.activeTab)
	if delta > 0 {
		index = (index + 1) % count
	} else {
		index = (index + count - 1) % count
	}
	state.activeTab = DetailTab(index)
	return state
}

func (state detailStateModel) withResetViewState() detailStateModel {
	state.viewState.reset()
	return state
}

func (state detailStateModel) withSyncedViewport(detailDocument detailDocument, viewportHeight int) detailStateModel {
	state.viewState.sync(detailDocument, viewportHeight)
	return state
}

func (state detailStateModel) withCursorAtLine(detailDocument detailDocument, line int) detailStateModel {
	state.viewState.cursor = detailDocument.clampPosition(detailPosition{line: line, column: 0})
	state.viewState.preferredColumn = 0
	return state
}

func (state detailStateModel) withFocusedLine(detailDocument detailDocument, viewportHeight int, line int) detailStateModel {
	state = state.withCursorAtLine(detailDocument, line)
	return state.withSyncedViewport(detailDocument, viewportHeight)
}

func (state detailStateModel) synced(detailIdentity string, detailDocument detailDocument, viewportHeight int, searchQuery string) detailStateModel {
	if detailIdentity != state.lastIdentity {
		state.lastIdentity = detailIdentity
		state = state.withResetViewState()
	}

	state = state.withSyncedViewport(detailDocument, viewportHeight)
	state.viewState.syncSearch(detailDocument, searchQuery)
	return state
}

func (state detailStateModel) syncedForRender(detailIdentity string, detailDocument detailDocument, wrapWidth int, viewportHeight int, searchQuery string) detailStateModel {
	state = state.withWrapWidth(wrapWidth)
	return state.synced(detailIdentity, detailDocument, viewportHeight, searchQuery)
}
