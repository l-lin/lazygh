package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/rivo/uniseg"
)

type wrappedInputCursorPosition struct {
	column int
	row    int
}

type wrappedInputSegment struct {
	start        int
	end          int
	displayWidth int
}

func wrappedInputCursorPositionForText(text string, cursorIndex int, width int) wrappedInputCursorPosition {
	if width < 1 {
		width = 1
	}

	textRunes := []rune(text)
	if cursorIndex < 0 {
		cursorIndex = 0
	}
	if cursorIndex > len(textRunes) {
		cursorIndex = len(textRunes)
	}

	lines := strings.Split(text, "\n")
	accumulatedRows := 0
	accumulatedCursorIndex := 0
	for lineIndex, line := range lines {
		lineRunes := []rune(line)
		lineLength := len(lineRunes)
		if cursorIndex <= accumulatedCursorIndex+lineLength || lineIndex == len(lines)-1 {
			actual := wrappedInputCursorPositionForLine(lineRunes, cursorIndex-accumulatedCursorIndex, width)
			actual.row += accumulatedRows
			return actual
		}

		accumulatedRows += wrappedInputRowCountForLine(lineRunes, width)
		accumulatedCursorIndex += lineLength + 1
	}

	return wrappedInputCursorPosition{}
}

func wrappedInputCursorPositionForLine(line []rune, cursorPosition int, width int) wrappedInputCursorPosition {
	if cursorPosition < 0 {
		cursorPosition = 0
	}
	if cursorPosition > len(line) {
		cursorPosition = len(line)
	}

	segments := wrappedInputSegmentsForLine(line, width)
	for segmentIndex, segment := range segments {
		if cursorPosition < segment.end {
			return wrappedInputCursorPosition{
				column: wrappedInputDisplayWidth(line[segment.start:cursorPosition]),
				row:    segmentIndex,
			}
		}

		isLastSegment := segmentIndex == len(segments)-1
		if cursorPosition != segment.end {
			continue
		}
		if isLastSegment {
			return wrappedInputCursorPosition{column: segment.displayWidth, row: segmentIndex}
		}
		nextSegment := segments[segmentIndex+1]
		if nextSegment.start > segment.end {
			return wrappedInputCursorPosition{column: segment.displayWidth, row: segmentIndex}
		}
	}

	lastSegment := segments[len(segments)-1]
	return wrappedInputCursorPosition{column: lastSegment.displayWidth, row: len(segments) - 1}
}

func wrappedInputRowCountForLine(line []rune, width int) int {
	return len(wrappedInputSegmentsForLine(line, width))
}

func wrappedInputSegmentsForLine(line []rune, columns int) []wrappedInputSegment {
	if columns < 1 {
		columns = 1
	}
	if len(line) == 0 {
		return []wrappedInputSegment{{}}
	}

	runeWidths := make([]int, 0, len(line))
	for _, character := range line {
		runeWidths = append(runeWidths, wrappedInputRuneWidth(character))
	}

	segments := make([]wrappedInputSegment, 0, 1)
	currentWidth := 0
	offset := 0
	lastWhitespaceIndex := -1
	for index, character := range line {
		runeWidth := runeWidths[index]
		currentWidth += runeWidth
		if currentWidth > columns {
			switch {
			case character == ' ':
				segments = append(segments, newWrappedInputSegment(offset, index, line, runeWidths))
				offset = index + 1
				currentWidth = 0
			case character == '-':
				segments = append(segments, newWrappedInputSegment(offset, index, line, runeWidths))
				offset = index
				currentWidth = runeWidth
			case lastWhitespaceIndex >= 0:
				end := lastWhitespaceIndex
				if line[lastWhitespaceIndex] == '-' {
					end = lastWhitespaceIndex + 1
				}
				segments = append(segments, newWrappedInputSegment(offset, end, line, runeWidths))
				offset = lastWhitespaceIndex + 1
				currentWidth = wrappedInputDisplayWidthByIndex(runeWidths, offset, index+1)
			default:
				segments = append(segments, newWrappedInputSegment(offset, index, line, runeWidths))
				offset = index
				currentWidth = runeWidth
			}
			lastWhitespaceIndex = -1
			continue
		}
		if character == ' ' || character == '-' {
			lastWhitespaceIndex = index
		}
	}

	segments = append(segments, newWrappedInputSegment(offset, len(line), line, runeWidths))
	return segments
}

func newWrappedInputSegment(start int, end int, line []rune, runeWidths []int) wrappedInputSegment {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if end > len(line) {
		end = len(line)
	}
	return wrappedInputSegment{
		start:        start,
		end:          end,
		displayWidth: wrappedInputDisplayWidthByIndex(runeWidths, start, end),
	}
}

func wrappedInputDisplayWidth(text []rune) int {
	actual := 0
	for _, character := range text {
		actual += wrappedInputRuneWidth(character)
	}
	return actual
}

func wrappedInputDisplayWidthByIndex(runeWidths []int, start int, end int) int {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if end > len(runeWidths) {
		end = len(runeWidths)
	}

	actual := 0
	for _, runeWidth := range runeWidths[start:end] {
		actual += runeWidth
	}
	return actual
}

func wrappedInputRuneWidth(character rune) int {
	actual := uniseg.StringWidth(string(character))
	if actual < 1 {
		return 1
	}
	return actual
}

func (program *Program) setWrappedMultilineInputCursor(view *gocui.View, text string, cursorIndex int) {
	if view == nil {
		return
	}

	innerWidth := max(view.InnerWidth(), 1)
	innerHeight := max(view.InnerHeight(), 1)

	position := wrappedInputCursorPositionForText(text, cursorIndex, innerWidth)
	originY := 0
	if position.row >= innerHeight {
		originY = position.row - innerHeight + 1
	}
	cursorY := max(position.row-originY, 0)
	if cursorY >= innerHeight {
		cursorY = innerHeight - 1
	}
	cursorX := max(position.column, 0)
	if cursorX >= innerWidth {
		cursorX = innerWidth - 1
	}

	view.SetOrigin(0, originY)
	view.SetCursor(cursorX, cursorY)
}
