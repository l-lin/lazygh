package tui

import "strings"

func (state reviewSessionState) withSelectedFileTreeRow(row int) reviewSessionState {
	state.selectedFileTreeRow = row
	return state
}

func (state reviewSessionState) clampedSelection(selectableRows []int, fileRows []int) reviewSessionState {
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return state
	}
	if state.selectedFileTreeRow < 0 {
		if state.mode != reviewSessionModeStory && len(fileRows) > 0 {
			state.selectedFileTreeRow = fileRows[0]
			return state
		}
		state.selectedFileTreeRow = selectableRows[0]
		return state
	}
	state.selectedFileTreeRow = adjustVisibleSelection(state.selectedFileTreeRow, selectableRows, 0)
	return state
}

func (state reviewSessionState) adjustedSelection(selectableRows []int, change int) reviewSessionState {
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return state
	}
	state.selectedFileTreeRow = adjustVisibleSelection(state.selectedFileTreeRow, selectableRows, change)
	return state
}

func (state reviewSessionState) selectionAtTop(selectableRows []int) reviewSessionState {
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return state
	}
	state.selectedFileTreeRow = selectableRows[0]
	return state
}

func (state reviewSessionState) selectionAtBottom(selectableRows []int) reviewSessionState {
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return state
	}
	state.selectedFileTreeRow = selectableRows[len(selectableRows)-1]
	return state
}

func (state reviewSessionState) withTreeRowCollapsed(rowID string, collapsed bool) reviewSessionState {
	trimmedRowID := strings.TrimSpace(rowID)
	if trimmedRowID == "" {
		return state
	}
	state.collapsedTreeRowIDs = copyReviewSessionCollapsedIDs(state.collapsedTreeRowIDs)
	state.collapsedTreeRowIDs[trimmedRowID] = collapsed
	return state
}

func (state reviewSessionState) withAllTreeRowsCollapsed(tree reviewDiffTree, collapsed bool) (reviewSessionState, bool) {
	if len(tree.Rows) == 0 {
		return state, false
	}

	collapsedRowIDs := copyReviewSessionCollapsedIDs(state.collapsedTreeRowIDs)
	changed := false
	for _, row := range tree.Rows {
		if !row.Foldable {
			continue
		}
		trimmedRowID := strings.TrimSpace(row.ID)
		if trimmedRowID == "" {
			continue
		}
		if actualCollapsed, ok := collapsedRowIDs[trimmedRowID]; !ok || actualCollapsed != collapsed {
			changed = true
		}
		collapsedRowIDs[trimmedRowID] = collapsed
	}
	state.collapsedTreeRowIDs = collapsedRowIDs
	return state, changed
}

func (state reviewSessionState) withThreadCollapsed(threadID string, collapsed bool) reviewSessionState {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return state
	}
	state.collapsedThreadIDs = copyReviewSessionCollapsedIDs(state.collapsedThreadIDs)
	state.collapsedThreadIDs[trimmedThreadID] = collapsed
	return state
}

func (state reviewSessionState) withAllThreadsCollapsed(threads []reviewDiffThread, collapsed bool) (reviewSessionState, bool) {
	if len(threads) == 0 {
		return state, false
	}

	collapsedThreadIDs := copyReviewSessionCollapsedIDs(state.collapsedThreadIDs)
	changed := false
	for _, thread := range threads {
		trimmedThreadID := strings.TrimSpace(thread.ID)
		if trimmedThreadID == "" {
			continue
		}
		if reviewDiffThreadCollapsed(thread, collapsedThreadIDs) != collapsed {
			changed = true
		}
		collapsedThreadIDs[trimmedThreadID] = collapsed
	}
	state.collapsedThreadIDs = collapsedThreadIDs
	return state, changed
}

func copyReviewSessionCollapsedIDs(source map[string]bool) map[string]bool {
	copied := make(map[string]bool, len(source))
	for id, collapsed := range source {
		copied[id] = collapsed
	}
	return copied
}
