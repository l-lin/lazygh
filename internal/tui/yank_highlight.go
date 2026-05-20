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

func (program *Program) activateYankHighlight(state *detailViewState, selection detailSelectionRange) {
	if program == nil || state == nil || program.yankHighlightDuration <= 0 {
		return
	}
	state.setYankHighlight(selection, program.currentTime().Add(program.yankHighlightDuration))
}

func (program *Program) hasYankHighlights() bool {
	if program == nil {
		return false
	}
	if program.detailViewState.hasYankHighlight() {
		return true
	}
	return program.pullRequestBuildRunPopup != nil && program.pullRequestBuildRunPopup.viewState.hasYankHighlight()
}

func (program *Program) clearExpiredYankHighlights() bool {
	if program == nil {
		return false
	}

	now := program.currentTime()
	cleared := program.detailViewState.clearExpiredYankHighlight(now)
	if popup := program.pullRequestBuildRunPopup; popup != nil {
		if popup.viewState.clearExpiredYankHighlight(now) {
			cleared = true
		}
	}
	return cleared
}

func (program *Program) currentTime() time.Time {
	if program == nil || program.now == nil {
		return time.Now()
	}
	return program.now()
}
