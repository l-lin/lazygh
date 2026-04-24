package tui

const viewportScrolloff = 4

func effectiveViewportScrolloff(visibleHeight int) int {
	visibleHeight = maxInt(1, visibleHeight)
	return minInt(viewportScrolloff, maxInt(0, (visibleHeight-1)/2))
}

func visibleViewportOrigin(selectedRow int, currentOriginY int, visibleHeight int, rowCount int) int {
	visibleHeight = maxInt(1, visibleHeight)
	rowCount = maxInt(1, rowCount)
	selectedRow = clampIndex(selectedRow, rowCount)
	maxOriginY := maxInt(0, rowCount-visibleHeight)
	originY := clampInt(currentOriginY, 0, maxOriginY)

	scrolloff := effectiveViewportScrolloff(visibleHeight)
	minimumCursorY := scrolloff
	maximumCursorY := visibleHeight - 1 - scrolloff
	cursorY := selectedRow - originY
	if cursorY < minimumCursorY {
		originY = selectedRow - minimumCursorY
	}
	if cursorY > maximumCursorY {
		originY = selectedRow - maximumCursorY
	}

	return clampInt(originY, 0, maxOriginY)
}
