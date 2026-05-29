package tui

type pullRequestBuildRunPopupClipboardResult struct {
	state              pullRequestBuildRunPopupState
	selection          detailSelectionRange
	text               string
	hasVisualSelection bool
}

func (state pullRequestBuildRunPopupState) withViewState(viewState detailViewState) pullRequestBuildRunPopupState {
	state.viewState = viewState
	return state
}

func (state pullRequestBuildRunPopupState) withViewStateSynced(document detailDocument, viewportHeight int) pullRequestBuildRunPopupState {
	state.viewState.sync(document, viewportHeight)
	return state
}

func (state pullRequestBuildRunPopupState) withSearchSynced(document detailDocument) pullRequestBuildRunPopupState {
	state.viewState.syncSearch(document, state.searchQuery)
	return state
}

func (state pullRequestBuildRunPopupState) withRenderStateSynced(document detailDocument, viewportHeight int) pullRequestBuildRunPopupState {
	state = state.withViewStateSynced(document, viewportHeight)
	return state.withSearchSynced(document)
}

func (state pullRequestBuildRunPopupState) withPendingPrefixCleared() pullRequestBuildRunPopupState {
	state.viewState.clearPendingPrefix()
	return state
}

func (state pullRequestBuildRunPopupState) pendingKeySequenceTarget() keySequenceTarget {
	return state.viewState.pendingKeySequence.pendingTarget
}

func (state pullRequestBuildRunPopupState) withPendingKeySequenceArmed(target keySequenceTarget) pullRequestBuildRunPopupState {
	state.viewState.pendingKeySequence.arm(target)
	return state
}

func (state pullRequestBuildRunPopupState) withPendingKeySequenceCleared() pullRequestBuildRunPopupState {
	state.viewState.pendingKeySequence.clear()
	return state
}

func (state pullRequestBuildRunPopupState) withVisualModeExited() pullRequestBuildRunPopupState {
	state.viewState.exitVisualMode()
	return state
}

func (state pullRequestBuildRunPopupState) withSearchOpened() pullRequestBuildRunPopupState {
	state.searchActive = true
	return state.withPendingPrefixCleared()
}

func (state pullRequestBuildRunPopupState) withSearchSubmitted(query string) pullRequestBuildRunPopupState {
	state.searchActive = false
	state.searchQuery = query
	return state
}

func (state pullRequestBuildRunPopupState) withSearchCancelled() pullRequestBuildRunPopupState {
	state.searchActive = false
	return state
}

func (state pullRequestBuildRunPopupState) preparedClipboard(document detailDocument, viewportHeight int) pullRequestBuildRunPopupClipboardResult {
	state = state.withViewStateSynced(document, viewportHeight)
	state = state.withPendingPrefixCleared()
	if !state.viewState.mode.isVisual() {
		return pullRequestBuildRunPopupClipboardResult{state: state}
	}

	selection, _ := detailSelectionForCurrentMode(state.viewState, document)
	text := state.viewState.selectedText(document)
	state = state.withVisualModeExited()
	return pullRequestBuildRunPopupClipboardResult{
		state:              state,
		selection:          selection,
		text:               text,
		hasVisualSelection: true,
	}
}
