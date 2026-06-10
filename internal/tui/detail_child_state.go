package tui

type detailClipboardResult struct {
	state              detailStateModel
	selection          detailSelectionRange
	text               string
	hasVisualSelection bool
}

func (state detailStateModel) withWrapWidth(width int) detailStateModel {
	state.wrapWidth = width
	return state
}

func (state detailStateModel) wordWrapEnabled() bool {
	return !state.wordWrapDisabled
}

func (state detailStateModel) withWordWrapEnabled(enabled bool) detailStateModel {
	state.wordWrapDisabled = !enabled
	return state
}

func (state detailStateModel) withWordWrapToggled() detailStateModel {
	return state.withWordWrapEnabled(!state.wordWrapEnabled())
}

func (state detailStateModel) withActiveTab(tab DetailTab) detailStateModel {
	state.activeTab = tab
	return state
}

func (state detailStateModel) withCommitDiffTabOpened(pullRequestKey string, commitOID string, shortLabel string) detailStateModel {
	state.commitDiffTab = state.commitDiffTab.withOpened(pullRequestKey, commitOID, shortLabel)
	return state
}

func (state detailStateModel) withCommitDiffTabCleared() detailStateModel {
	state.commitDiffTab = state.commitDiffTab.cleared()
	if state.activeTab == CommitChangesDetailTab {
		state.activeTab = DescriptionDetailTab
	}
	return state
}

func (state detailStateModel) withProjectedScreenStateApplication(application projectedScreenStateApplication, visibleTabs []DetailTab) detailStateModel {
	if !application.hasDetailTab {
		return state
	}
	if tab, ok := detailTabAtIndex(visibleTabs, application.activeDetailTabIndex); ok {
		return state.withActiveTab(tab)
	}
	return state
}

func (state detailStateModel) withAdvancedActiveTab(delta int, visibleTabs []DetailTab) detailStateModel {
	if len(visibleTabs) == 0 || delta == 0 {
		return state
	}

	index := detailTabIndex(visibleTabs, state.activeTab)
	if delta > 0 {
		index = (index + 1) % len(visibleTabs)
	} else {
		index = (index + len(visibleTabs) - 1) % len(visibleTabs)
	}
	state.activeTab = visibleTabs[index]
	return state
}

func detailTabIndex(visibleTabs []DetailTab, activeTab DetailTab) int {
	for index, tab := range visibleTabs {
		if tab == activeTab {
			return index
		}
	}
	return 0
}

func detailTabAtIndex(visibleTabs []DetailTab, index int) (DetailTab, bool) {
	if index < 0 || index >= len(visibleTabs) {
		return DescriptionDetailTab, false
	}
	return visibleTabs[index], true
}

func (state detailStateModel) withPendingPrefixCleared() detailStateModel {
	state.viewState.clearPendingPrefix()
	return state
}

func (state detailStateModel) pendingKeySequenceTarget() keySequenceTarget {
	return state.viewState.pendingKeySequence.pendingTarget
}

func (state detailStateModel) withPendingKeySequenceArmed(target keySequenceTarget) detailStateModel {
	state.viewState.pendingKeySequence.arm(target)
	return state
}

func (state detailStateModel) withPendingKeySequenceCleared() detailStateModel {
	state.viewState.pendingKeySequence.clear()
	return state
}

func (state detailStateModel) withVisualModeExited() detailStateModel {
	state.viewState.exitVisualMode()
	return state
}

func (state detailStateModel) withResetViewState() detailStateModel {
	state.viewState.reset()
	return state
}

func (state detailStateModel) withViewState(viewState detailViewState) detailStateModel {
	state.viewState = viewState
	return state
}

func (state detailStateModel) withSyncedViewport(detailDocument detailDocument, viewportHeight int) detailStateModel {
	state.viewState.sync(detailDocument, viewportHeight)
	return state
}

func (state detailStateModel) withSearchSynced(detailDocument detailDocument, searchQuery string) detailStateModel {
	state.viewState.syncSearch(detailDocument, searchQuery)
	return state
}

func (state detailStateModel) withViewportSyncPreserved() detailStateModel {
	state.viewState.preserveViewportSyncCount++
	return state
}

func (state detailStateModel) withViewportOperation(detailDocument detailDocument, viewportHeight int, operation detailViewportOperation) detailStateModel {
	switch operation {
	case detailViewportOperationScrollDown:
		state.viewState.scrollDown(detailDocument, viewportHeight)
	case detailViewportOperationScrollUp:
		state.viewState.scrollUp(detailDocument, viewportHeight)
	case detailViewportOperationRecenter:
		state.viewState.recenter(detailDocument, viewportHeight)
		state = state.withViewportSyncPreserved()
	case detailViewportOperationPlaceTop:
		state.viewState.placeCursorAtViewportTop(detailDocument, viewportHeight)
		state = state.withViewportSyncPreserved()
	case detailViewportOperationPlaceBottom:
		state.viewState.placeCursorAtViewportBottom(detailDocument, viewportHeight)
		state = state.withViewportSyncPreserved()
	}
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

func (state detailStateModel) withFocusedLineAndSearchSynced(detailDocument detailDocument, viewportHeight int, line int, searchQuery string) detailStateModel {
	state = state.withFocusedLine(detailDocument, viewportHeight, line)
	return state.withSearchSynced(detailDocument, searchQuery)
}

func (state detailStateModel) synced(detailIdentity string, detailDocument detailDocument, viewportHeight int, searchQuery string) detailStateModel {
	if detailIdentity != state.lastIdentity {
		state.lastIdentity = detailIdentity
		state = state.withResetViewState()
	}

	state = state.withSyncedViewport(detailDocument, viewportHeight)
	return state.withSearchSynced(detailDocument, searchQuery)
}

func (state detailStateModel) syncedForRender(detailIdentity string, detailDocument detailDocument, wrapWidth int, viewportHeight int, searchQuery string) detailStateModel {
	state = state.withWrapWidth(wrapWidth)
	return state.synced(detailIdentity, detailDocument, viewportHeight, searchQuery)
}

func (state detailStateModel) preparedClipboard(detailIdentity string, detailDocument detailDocument, viewportHeight int, searchQuery string) detailClipboardResult {
	state = state.synced(detailIdentity, detailDocument, viewportHeight, searchQuery)
	if !state.viewState.mode.isVisual() {
		return detailClipboardResult{state: state}
	}

	selection, _ := detailSelectionForCurrentMode(state.viewState, detailDocument)
	text := state.viewState.selectedText(detailDocument)
	state = state.withVisualModeExited()
	return detailClipboardResult{
		state:              state,
		selection:          selection,
		text:               text,
		hasVisualSelection: true,
	}
}
