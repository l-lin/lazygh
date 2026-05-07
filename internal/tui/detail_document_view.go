package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
)

func renderDetailDocumentView(view *gocui.View, document detailDocument, state detailViewState) {
	if view == nil {
		return
	}

	view.Clear()
	searchMatchRanges := detailSearchMatchRanges(state.searchMatches)
	for rowIndex, row := range document.rows {
		if rowIndex > 0 {
			fmt.Fprint(view, "\n")
		}
		fmt.Fprint(view, renderDetailRow(document, row, searchMatchRanges, state))
	}

	cursorRow, cursorColumn := state.screenPosition(document)
	originX := 0
	innerWidth := view.InnerWidth()
	if innerWidth > 0 && cursorColumn >= innerWidth {
		originX = cursorColumn - innerWidth + 1
		cursorColumn = innerWidth - 1
	}
	view.SetOrigin(originX, state.originRow)
	view.SetCursor(cursorColumn, cursorRow-state.originRow)
}
