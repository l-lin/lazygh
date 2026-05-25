package tui

import "strings"

func (state *reviewSessionState) setSelectedFileTreeRow(row int) {
	if state == nil {
		return
	}
	state.selectedFileTreeRow = row
}

func (state *reviewSessionState) clampSelection(selectableRows []int, fileRows []int) {
	if state == nil {
		return
	}
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return
	}
	if state.selectedFileTreeRow < 0 {
		if state.mode != reviewSessionModeStory && len(fileRows) > 0 {
			state.selectedFileTreeRow = fileRows[0]
			return
		}
		state.selectedFileTreeRow = selectableRows[0]
		return
	}
	state.selectedFileTreeRow = adjustVisibleSelection(state.selectedFileTreeRow, selectableRows, 0)
}

func (state *reviewSessionState) adjustSelection(selectableRows []int, change int) {
	if state == nil {
		return
	}
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return
	}
	state.selectedFileTreeRow = adjustVisibleSelection(state.selectedFileTreeRow, selectableRows, change)
}

func (state *reviewSessionState) moveSelectionToTop(selectableRows []int) {
	if state == nil {
		return
	}
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return
	}
	state.selectedFileTreeRow = selectableRows[0]
}

func (state *reviewSessionState) moveSelectionToBottom(selectableRows []int) {
	if state == nil {
		return
	}
	if len(selectableRows) == 0 {
		state.selectedFileTreeRow = 0
		return
	}
	state.selectedFileTreeRow = selectableRows[len(selectableRows)-1]
}

func (state *reviewSessionState) setTreeRowCollapsed(rowID string, collapsed bool) {
	trimmedRowID := strings.TrimSpace(rowID)
	if state == nil || trimmedRowID == "" {
		return
	}
	if state.collapsedTreeRowIDs == nil {
		state.collapsedTreeRowIDs = map[string]bool{}
	}
	state.collapsedTreeRowIDs[trimmedRowID] = collapsed
}

func (state *reviewSessionState) setAllTreeRowsCollapsed(tree reviewDiffTree, collapsed bool) bool {
	if state == nil || len(tree.Rows) == 0 {
		return false
	}
	if state.collapsedTreeRowIDs == nil {
		state.collapsedTreeRowIDs = map[string]bool{}
	}

	changed := false
	for _, row := range tree.Rows {
		if !row.Foldable {
			continue
		}
		trimmedRowID := strings.TrimSpace(row.ID)
		if trimmedRowID == "" {
			continue
		}
		if actualCollapsed, ok := state.collapsedTreeRowIDs[trimmedRowID]; !ok || actualCollapsed != collapsed {
			changed = true
		}
		state.collapsedTreeRowIDs[trimmedRowID] = collapsed
	}
	return changed
}

func (state *reviewSessionState) setThreadCollapsed(threadID string, collapsed bool) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if state == nil || trimmedThreadID == "" {
		return
	}
	if state.collapsedThreadIDs == nil {
		state.collapsedThreadIDs = map[string]bool{}
	}
	state.collapsedThreadIDs[trimmedThreadID] = collapsed
}

func (state *reviewSessionState) setAllThreadsCollapsed(threads []reviewDiffThread, collapsed bool) bool {
	if state == nil || len(threads) == 0 {
		return false
	}
	if state.collapsedThreadIDs == nil {
		state.collapsedThreadIDs = map[string]bool{}
	}

	changed := false
	for _, thread := range threads {
		trimmedThreadID := strings.TrimSpace(thread.ID)
		if trimmedThreadID == "" {
			continue
		}
		if reviewDiffThreadCollapsed(thread, state.collapsedThreadIDs) != collapsed {
			changed = true
		}
		state.collapsedThreadIDs[trimmedThreadID] = collapsed
	}
	return changed
}
