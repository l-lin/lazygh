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

	currentRow := document.rowIndexForPosition(state.cursor)
	maxOriginRow := maxInt(0, document.rowCount()-viewportHeight)
	if state.originRow > maxOriginRow {
		state.originRow = maxOriginRow
	}
	if state.originRow < 0 {
		state.originRow = 0
	}
	if currentRow < state.originRow {
		state.originRow = currentRow
	}
	if currentRow >= state.originRow+viewportHeight {
		state.originRow = currentRow - viewportHeight + 1
	}
	if state.originRow < 0 {
		state.originRow = 0
	}
	if state.originRow > maxOriginRow {
		state.originRow = maxOriginRow
	}
}

func (state *detailViewState) reset() {
	*state = detailViewState{currentSearchMatch: -1}
}

func (state *detailViewState) clearPendingPrefix() {
	state.pendingKeySequence.clear()
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
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, pageDelta(viewportHeight), state.preferredColumn)
	state.recenter(document, viewportHeight)
}

func (state *detailViewState) pageUp(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, -pageDelta(viewportHeight), state.preferredColumn)
	state.recenter(document, viewportHeight)
}

func (state *detailViewState) recenter(document detailDocument, viewportHeight int) {
	state.sync(document, viewportHeight)
	currentRow := document.rowIndexForPosition(state.cursor)
	state.originRow = centeredViewportOrigin(currentRow, viewportHeight, document.rowCount())
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

func (state *detailViewState) handleGoToTopPrefix(document detailDocument, viewportHeight int) {
	target := keySequenceTargetFor(viewDetailName, keymapScopeDetail, "move_cursor_to_top")
	if !state.pendingKeySequence.armOrConsume(target) {
		return
	}

	state.moveToTop(document, viewportHeight)
}

func (state *detailViewState) armInlineConversationTogglePrefix() {
	state.pendingKeySequence.arm(keySequenceTargetFor(viewDetailName, keymapScopeDetail, "open_actions_popup"))
}

func (state *detailViewState) consumeInlineConversationTogglePrefix() bool {
	return state.pendingKeySequence.consume(keySequenceTargetFor(viewDetailName, keymapScopeDetail, "open_actions_popup"))
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

func (state *detailViewState) moveToPreviousWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToPreviousWord(state.cursor)
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
	return document.rowIndexForPosition(state.cursor), document.screenColumnForPosition(state.cursor)
}
