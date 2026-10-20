package tui

type detailViewportRenderPlan struct {
	pinnedStartRow     int
	pinnedEndRow       int
	pinnedKnown        bool
	bodyStartRow       int
	bodyEndRow         int
	bodyVerticalOrigin int
}

func maximumDetailViewportOrigin(document detailDocument, viewportHeight int) int {
	origin := maxInt(0, document.rowCount()-maxInt(1, viewportHeight))
	for {
		plan := planVisibleDetailViewport(document, origin, viewportHeight)
		if plan.bodyEndRow >= document.rowCount() {
			return origin
		}
		// Pinned headers consume screen rows without advancing the document body.
		origin++
	}
}

func planVisibleDetailViewport(document detailDocument, originRow int, viewportHeight int) detailViewportRenderPlan {
	rowCount := maxInt(document.rowCount(), 1)
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	bodyStartRow := clampInt(originRow, 0, maxInt(0, rowCount-1))
	plan := detailViewportRenderPlan{
		bodyStartRow: bodyStartRow,
		bodyEndRow:   minInt(document.rowCount(), bodyStartRow+viewportHeight),
	}
	if document.rowCount() == 0 || bodyStartRow >= document.rowCount() {
		return plan
	}

	topRow := document.rows[bodyStartRow]
	if topRow.owningHeaderLine < 0 {
		return plan
	}

	headerStartRow, headerEndRow, ok := document.rowRangeForLine(topRow.owningHeaderLine)
	if !ok || bodyStartRow < headerEndRow {
		return plan
	}

	maximumPinnedHeight := viewportHeight - 1
	if maximumPinnedHeight < 1 {
		return plan
	}

	pinnedHeight := minInt(headerEndRow-headerStartRow, maximumPinnedHeight)
	if pinnedHeight < 1 {
		return plan
	}

	plan.pinnedStartRow = headerStartRow
	plan.pinnedEndRow = headerStartRow + pinnedHeight
	plan.pinnedKnown = true
	plan.bodyVerticalOrigin = pinnedHeight
	plan.bodyEndRow = minInt(document.rowCount(), bodyStartRow+viewportHeight-pinnedHeight)
	return plan
}
