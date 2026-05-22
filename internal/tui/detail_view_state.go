package tui

func newDetailViewState() detailViewState {
	return detailViewState{currentSearchMatch: -1}
}

func (state *detailViewState) sync(document detailDocument, viewportHeight int) {
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	state.cursor = document.clampPosition(state.cursor)
	if state.mode.isVisual() {
		state.visualAnchor = document.clampPosition(state.visualAnchor)
	} else {
		state.visualAnchor = state.cursor
	}

	state.originRow = clampInt(state.originRow, 0, maxInt(0, document.rowCount()-viewportHeight))
	if state.manualViewportScroll {
		return
	}
	if state.preserveViewportSyncCount > 0 {
		state.preserveViewportSyncCount--
		return
	}

	currentRow := document.rowIndexForPosition(state.cursor)
	state.originRow = visibleViewportOrigin(currentRow, state.originRow, viewportHeight, document.rowCount())
}

func (state *detailViewState) reset() {
	*state = detailViewState{currentSearchMatch: -1}
}

func (state *detailViewState) scrollDown(document detailDocument, viewportHeight int) {
	state.scroll(document, viewportHeight, 1)
}

func (state *detailViewState) scrollUp(document detailDocument, viewportHeight int) {
	state.scroll(document, viewportHeight, -1)
}

func (state *detailViewState) scroll(document detailDocument, viewportHeight int, delta int) {
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	currentRow := document.rowIndexForPosition(state.cursor)
	newOrigin := clampInt(state.originRow+delta, 0, maxInt(0, document.rowCount()-viewportHeight))
	newRow := clampInt(currentRow+delta, 0, maxInt(0, document.rowCount()-1))
	newRow = clampInt(newRow, newOrigin, minInt(document.rowCount()-1, newOrigin+viewportHeight-1))
	state.originRow = newOrigin
	state.cursor = document.positionForRow(newRow, state.preferredColumn)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.manualViewportScroll = true
	state.sync(document, viewportHeight)
}

func (state *detailViewState) clearPendingPrefix() {
	state.pendingKeySequence.clear()
	state.pendingCharacterMotion = detailPendingCharacterMotion{}
	state.pendingYank = false
}

func (state *detailViewState) enterVisualMode() {
	state.clearPendingPrefix()
	if state.mode == detailVisualMode {
		return
	}

	state.mode = detailVisualMode
	state.visualAnchor = state.cursor
}

func (state *detailViewState) enterLineVisualMode(document detailDocument) {
	state.clearPendingPrefix()
	if state.mode == detailLineVisualMode {
		return
	}

	state.mode = detailLineVisualMode
	state.visualAnchor = document.moveToRowStart(state.cursor)
}

func (state *detailViewState) exitVisualMode() {
	state.clearPendingPrefix()
	state.mode = detailNormalMode
	state.visualAnchor = state.cursor
}

func (state *detailViewState) moveLeft(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveLeft(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveRight(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveRight(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveDown(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, 1, state.preferredColumn)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveUp(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, -1, state.preferredColumn)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) pageDown(document detailDocument, viewportHeight int) {
	state.page(document, viewportHeight, pageDelta(viewportHeight))
}

func (state *detailViewState) pageUp(document detailDocument, viewportHeight int) {
	state.page(document, viewportHeight, -pageDelta(viewportHeight))
}

func (state *detailViewState) fullPageDown(document detailDocument, viewportHeight int) {
	state.page(document, viewportHeight, fullPageDelta(viewportHeight))
}

func (state *detailViewState) fullPageUp(document detailDocument, viewportHeight int) {
	state.page(document, viewportHeight, -fullPageDelta(viewportHeight))
}

func (state *detailViewState) page(document detailDocument, viewportHeight int, delta int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, delta, state.preferredColumn)
	state.recenter(document, viewportHeight)
}

func (state *detailViewState) recenter(document detailDocument, viewportHeight int) {
	state.placeCursorInViewport(document, viewportHeight, viewportPlacementCenter)
}

func (state *detailViewState) placeCursorAtViewportTop(document detailDocument, viewportHeight int) {
	state.placeCursorInViewport(document, viewportHeight, viewportPlacementTop)
}

func (state *detailViewState) placeCursorAtViewportBottom(document detailDocument, viewportHeight int) {
	state.placeCursorInViewport(document, viewportHeight, viewportPlacementBottom)
}

func (state *detailViewState) placeCursorInViewport(document detailDocument, viewportHeight int, placement viewportPlacement) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	currentRow := document.rowIndexForPosition(state.cursor)
	state.originRow = placedViewportOrigin(currentRow, viewportHeight, document.rowCount(), placement)
	state.preserveViewportSyncCount = 2
}

func (state *detailViewState) moveToRowStart(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToRowStart(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToRowEnd(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToRowEnd(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToTop(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToTop()
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToBottom(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToBottom()
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToNextWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToNextWord(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToWordEnd(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToWordEnd(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToNextBigWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToNextBigWord(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToBigWordEnd(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToBigWordEnd(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToPreviousWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToPreviousWord(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToPreviousBigWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToPreviousBigWord(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state detailViewState) visualSelection(document detailDocument) (detailPosition, detailPosition, bool) {
	if state.mode != detailVisualMode {
		return detailPosition{}, detailPosition{}, false
	}

	start := document.clampPosition(state.visualAnchor)
	end := document.clampPosition(state.cursor)
	if document.comparePositions(start, end) > 0 {
		start, end = end, start
	}

	return start, end, true
}

func (state detailViewState) visualRowSelection(document detailDocument) (int, int, bool) {
	if state.mode != detailLineVisualMode {
		return 0, 0, false
	}

	startRow := document.rowIndexForPosition(state.visualAnchor)
	endRow := document.rowIndexForPosition(state.cursor)
	if startRow > endRow {
		startRow, endRow = endRow, startRow
	}

	return startRow, endRow, true
}

func (state detailViewState) selectedText(document detailDocument) string {
	if startRow, endRow, ok := state.visualRowSelection(document); ok {
		return document.rowSelectionText(startRow, endRow)
	}

	start, end, ok := state.visualSelection(document)
	if !ok {
		return ""
	}

	return document.selectionText(start, end)
}

func (state detailViewState) isPositionSelected(document detailDocument, position detailPosition) bool {
	if startRow, endRow, ok := state.visualRowSelection(document); ok {
		rowIndex := document.rowIndexForPosition(position)
		return startRow <= rowIndex && rowIndex <= endRow
	}

	start, end, ok := state.visualSelection(document)
	if !ok {
		return false
	}

	return document.comparePositions(start, position) <= 0 && document.comparePositions(position, end) <= 0
}

func (state detailViewState) screenPosition(document detailDocument) (int, int) {
	return document.rowIndexForPosition(state.cursor), document.visualScreenColumnForPosition(state.cursor)
}
