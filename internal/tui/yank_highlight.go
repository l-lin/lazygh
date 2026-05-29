package tui

import "time"

type detailSelectionRange struct {
	linewise bool
	start    detailPosition
	end      detailPosition
	startRow int
	endRow   int
}

type detailYankHighlightState struct {
	active    bool
	linewise  bool
	start     detailPosition
	end       detailPosition
	startRow  int
	endRow    int
	expiresAt time.Time
}

func detailSelectionForCurrentMode(state detailViewState, document detailDocument) (detailSelectionRange, bool) {
	if startRow, endRow, ok := state.visualRowSelection(document); ok {
		return detailSelectionRange{linewise: true, startRow: startRow, endRow: endRow}, true
	}

	start, end, ok := state.visualSelection(document)
	if !ok {
		return detailSelectionRange{}, false
	}
	return detailSelectionRange{start: start, end: end}, true
}

func detailSelectionForYankMotion(document detailDocument, anchor detailPosition, target detailPosition, selectionKind detailYankMotionSelectionKind) (detailSelectionRange, bool) {
	anchor = document.clampPosition(anchor)
	target = document.clampPosition(target)

	switch selectionKind {
	case detailYankMotionLinewise:
		startRow := document.rowIndexForPosition(anchor)
		endRow := document.rowIndexForPosition(target)
		if startRow > endRow {
			startRow, endRow = endRow, startRow
		}
		return detailSelectionRange{linewise: true, startRow: startRow, endRow: endRow}, true
	case detailYankMotionCharacterExclusive:
		if document.comparePositions(anchor, target) == 0 {
			return detailSelectionRange{}, false
		}
		start, end := anchor, target
		if document.comparePositions(start, end) > 0 {
			start, end = end, start
		}
		end = document.moveLeft(end)
		if document.comparePositions(start, end) > 0 {
			return detailSelectionRange{}, false
		}
		return detailSelectionRange{start: start, end: end}, true
	default:
		if document.comparePositions(anchor, target) == 0 {
			return detailSelectionRange{}, false
		}
		start, end := anchor, target
		if document.comparePositions(start, end) > 0 {
			start, end = end, start
		}
		return detailSelectionRange{start: start, end: end}, true
	}
}

func (selection detailSelectionRange) text(document detailDocument) string {
	if selection.linewise {
		return document.rowSelectionText(selection.startRow, selection.endRow)
	}
	return document.selectionText(selection.start, selection.end)
}

func (state *detailViewState) setYankHighlight(selection detailSelectionRange, expiresAt time.Time) {
	state.yankHighlight = detailYankHighlightState{
		active:    true,
		linewise:  selection.linewise,
		start:     selection.start,
		end:       selection.end,
		startRow:  selection.startRow,
		endRow:    selection.endRow,
		expiresAt: expiresAt,
	}
}

func (state *detailViewState) clearYankHighlight() {
	state.yankHighlight = detailYankHighlightState{}
}

func (state *detailViewState) clearExpiredYankHighlight(now time.Time) bool {
	if !state.yankHighlight.active {
		return false
	}
	if state.yankHighlight.expiresAt.IsZero() || now.Before(state.yankHighlight.expiresAt) {
		return false
	}
	state.clearYankHighlight()
	return true
}

func (state detailViewState) hasYankHighlight() bool {
	return state.yankHighlight.active
}

func (state detailStateModel) hasYankHighlight() bool {
	return state.viewState.hasYankHighlight()
}

func (state detailStateModel) withYankHighlightActivated(selection detailSelectionRange, expiresAt time.Time) detailStateModel {
	state.viewState.setYankHighlight(selection, expiresAt)
	return state
}

func (state detailStateModel) withExpiredYankHighlightCleared(now time.Time) (detailStateModel, bool) {
	if !state.viewState.clearExpiredYankHighlight(now) {
		return state, false
	}
	return state, true
}

func (state pullRequestBuildRunPopupState) hasYankHighlight() bool {
	return state.viewState.hasYankHighlight()
}

func (state pullRequestBuildRunPopupState) withYankHighlightActivated(selection detailSelectionRange, expiresAt time.Time) pullRequestBuildRunPopupState {
	state.viewState.setYankHighlight(selection, expiresAt)
	return state
}

func (state pullRequestBuildRunPopupState) withExpiredYankHighlightCleared(now time.Time) (pullRequestBuildRunPopupState, bool) {
	if !state.viewState.clearExpiredYankHighlight(now) {
		return state, false
	}
	return state, true
}

func (state detailViewState) isPositionYankHighlighted(document detailDocument, position detailPosition) bool {
	highlight := state.yankHighlight
	if !highlight.active {
		return false
	}
	if highlight.linewise {
		rowIndex := document.rowIndexForPosition(position)
		return highlight.startRow <= rowIndex && rowIndex <= highlight.endRow
	}
	return document.comparePositions(highlight.start, position) <= 0 && document.comparePositions(position, highlight.end) <= 0
}

func (program *Program) yankHighlightExpiryTime() (time.Time, bool) {
	if program == nil || program.timingState.yankHighlightDuration <= 0 {
		return time.Time{}, false
	}
	return program.currentTime().Add(program.timingState.yankHighlightDuration), true
}

func (program *Program) activateDetailYankHighlight(selection detailSelectionRange) {
	expiresAt, ok := program.yankHighlightExpiryTime()
	if !ok {
		return
	}
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		return state.withYankHighlightActivated(selection, expiresAt)
	})
}

func (program *Program) activatePullRequestBuildRunPopupYankHighlight(selection detailSelectionRange) {
	expiresAt, ok := program.yankHighlightExpiryTime()
	if !ok {
		return
	}
	program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		return state.withYankHighlightActivated(selection, expiresAt)
	})
}

func (program *Program) hasYankHighlights() bool {
	if program == nil {
		return false
	}
	if program.detailState.hasYankHighlight() {
		return true
	}
	return program.pullRequestBuildRunPopup != nil && program.pullRequestBuildRunPopup.hasYankHighlight()
}

func (program *Program) clearExpiredYankHighlights() bool {
	if program == nil {
		return false
	}

	now := program.currentTime()
	cleared := false
	program.updateDetailState(func(state detailStateModel) detailStateModel {
		updatedState, changed := state.withExpiredYankHighlightCleared(now)
		cleared = cleared || changed
		return updatedState
	})
	if program.pullRequestBuildRunPopup != nil {
		program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
			updatedState, changed := state.withExpiredYankHighlightCleared(now)
			cleared = cleared || changed
			return updatedState
		})
	}
	return cleared
}

func (program *Program) currentTime() time.Time {
	if program == nil || program.timingState.now == nil {
		return time.Now()
	}
	return program.timingState.now()
}
