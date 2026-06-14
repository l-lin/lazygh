package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/theme"
)

func renderVisibleDetailDocumentView(view *gocui.View, document detailDocument, state detailViewState) {
	viewportHeight := maxInt(1, viewPageSize(view))
	plan := planVisibleDetailViewport(document, state.originRow, viewportHeight)
	renderPlannedDetailDocumentRows(view, document, state, plan)
}

func renderPlannedDetailDocumentRows(view *gocui.View, document detailDocument, state detailViewState, plan detailViewportRenderPlan) {
	if view == nil {
		return
	}

	view.Clear()
	searchMatchRanges := detailSearchMatchRanges(state.searchMatches)
	wroteRow := false
	if plan.pinnedKnown {
		wroteRow = renderDetailDocumentRowRange(view, document, state, searchMatchRanges, plan.pinnedStartRow, plan.pinnedEndRow, detailRowRenderOptions{backgroundOverrideHex: theme.StickyFileHeaderBackgroundHex}, wroteRow)
	}
	wroteRow = renderDetailDocumentRowRange(view, document, state, searchMatchRanges, plan.bodyStartRow, plan.bodyEndRow, detailRowRenderOptions{}, wroteRow)

	cursorRow, cursorColumn := state.screenPosition(document)
	originX := 0
	innerWidth := view.InnerWidth()
	if innerWidth > 0 && cursorColumn >= innerWidth {
		originX = cursorColumn - innerWidth + 1
		cursorColumn = innerWidth - 1
	}
	cursorY := plan.bodyVerticalOrigin + (cursorRow - plan.bodyStartRow)
	if plan.pinnedKnown && cursorRow >= plan.pinnedStartRow && cursorRow < plan.pinnedEndRow {
		cursorY = cursorRow - plan.pinnedStartRow
	}
	view.SetOrigin(originX, 0)
	view.SetCursor(cursorColumn, maxInt(cursorY, 0))
}

func renderDetailDocumentRowRange(view *gocui.View, document detailDocument, state detailViewState, searchMatchRanges map[int][]detailColumnRange, startRow int, endRow int, options detailRowRenderOptions, wroteRow bool) bool {
	rowCount := document.rowCount()
	if rowCount == 0 {
		return wroteRow
	}

	startRow = clampInt(startRow, 0, rowCount-1)
	endRow = clampInt(endRow, startRow+1, rowCount)
	for rowIndex := startRow; rowIndex < endRow; rowIndex++ {
		if wroteRow {
			fmt.Fprint(view, "\n")
		}
		fmt.Fprint(view, renderDetailRowWithOptions(document, document.rows[rowIndex], searchMatchRanges, state, options))
		wroteRow = true
	}
	return wroteRow
}
