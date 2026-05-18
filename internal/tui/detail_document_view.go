package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
)

func renderDetailDocumentView(view *gocui.View, document detailDocument, state detailViewState) {
	renderDetailDocumentRows(view, document, state, 0, document.rowCount(), state.originRow)
}

func renderVisibleDetailDocumentView(view *gocui.View, document detailDocument, state detailViewState) {
	startRow := clampInt(state.originRow, 0, maxInt(0, document.rowCount()-1))
	viewportHeight := maxInt(1, viewPageSize(view))
	endRow := minInt(document.rowCount(), startRow+viewportHeight)
	renderDetailDocumentRows(view, document, state, startRow, endRow, 0)
}

func renderDetailDocumentRows(view *gocui.View, document detailDocument, state detailViewState, startRow int, endRow int, verticalOrigin int) {
	if view == nil {
		return
	}

	rowCount := document.rowCount()
	if rowCount == 0 {
		rowCount = 1
	}
	startRow = clampInt(startRow, 0, rowCount-1)
	endRow = clampInt(endRow, startRow+1, rowCount)

	view.Clear()
	searchMatchRanges := detailSearchMatchRanges(state.searchMatches)
	for rowIndex := startRow; rowIndex < endRow; rowIndex++ {
		if rowIndex > startRow {
			fmt.Fprint(view, "\n")
		}
		fmt.Fprint(view, renderDetailRow(document, document.rows[rowIndex], searchMatchRanges, state))
	}

	cursorRow, cursorColumn := state.screenPosition(document)
	originX := 0
	innerWidth := view.InnerWidth()
	if innerWidth > 0 && cursorColumn >= innerWidth {
		originX = cursorColumn - innerWidth + 1
		cursorColumn = innerWidth - 1
	}
	view.SetOrigin(originX, verticalOrigin)
	view.SetCursor(cursorColumn, cursorRow-startRow-verticalOrigin)
}
