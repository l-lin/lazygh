package tui

func boundedHalfWidth(maxWidth int, minWidth int, fallbackWidth int) int {
	return boundedFractionWidth(maxWidth, minWidth, fallbackWidth, 2)
}

func boundedQuarterWidth(maxWidth int, minWidth int, fallbackWidth int) int {
	return boundedFractionWidth(maxWidth, minWidth, fallbackWidth, 4)
}

func boundedFractionWidth(maxWidth int, minWidth int, fallbackWidth int, divisor int) int {
	if divisor < 1 {
		divisor = 1
	}

	totalWidth := maxWidth / divisor
	if totalWidth < minWidth {
		totalWidth = fallbackWidth
	}
	if totalWidth > maxWidth-4 {
		totalWidth = maxInt(10, maxWidth-4)
	}
	if totalWidth > maxWidth {
		totalWidth = maxWidth
	}
	if totalWidth < 1 {
		totalWidth = 1
	}

	return totalWidth
}

func centeredOverlayFrame(maxX int, maxY int, totalWidth int, totalHeight int) paneFrame {
	if maxX < 1 {
		maxX = 1
	}
	if maxY < 1 {
		maxY = 1
	}

	totalWidth = clampOverlayDimension(totalWidth, maxX)
	totalHeight = clampOverlayDimension(totalHeight, maxY)

	x0 := clampCoordinate((maxX-totalWidth)/2, maxX)
	y0 := clampCoordinate((maxY-totalHeight)/2, maxY)
	x1 := x0 + totalWidth - 1
	y1 := y0 + totalHeight - 1
	if x1 >= maxX {
		x1 = maxX - 1
		x0 = clampCoordinate(x1-totalWidth+1, maxX)
	}
	if y1 >= maxY {
		y1 = maxY - 1
		y0 = clampCoordinate(y1-totalHeight+1, maxY)
	}

	return paneFrame{x0: x0, y0: y0, x1: x1, y1: y1}
}

func clampOverlayDimension(size int, maxSize int) int {
	if size > maxSize {
		size = maxSize
	}
	if size < 1 {
		return 1
	}

	return size
}

func clampCoordinate(value int, maxValue int) int {
	if maxValue <= 0 {
		return 0
	}
	if value < 0 {
		return 0
	}
	if value >= maxValue {
		return maxValue - 1
	}
	return value
}
