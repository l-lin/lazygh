package tui

import (
	"strings"
	"unicode"
)

func (document detailDocument) globalIndex(position detailPosition) int {
	position = document.clampPosition(position)
	if len(document.lineStartOffsets) == 0 {
		return 0
	}

	return document.lineStartOffsets[position.line] + position.column
}

func (document detailDocument) positionForGlobalIndex(index int) detailPosition {
	if len(document.text) == 0 {
		return detailPosition{}
	}
	if index <= 0 {
		return document.firstPosition()
	}
	if index >= len(document.text) {
		return document.lastPosition()
	}

	lineIndex := 0
	for candidateLine := 1; candidateLine < len(document.lineStartOffsets); candidateLine++ {
		if document.lineStartOffsets[candidateLine] > index {
			break
		}
		lineIndex = candidateLine
	}

	lineLength := document.lineLength(lineIndex)
	if lineLength == 0 {
		return detailPosition{line: lineIndex, column: 0}
	}

	column := index - document.lineStartOffsets[lineIndex]
	if column < 0 {
		column = 0
	}
	if column >= lineLength {
		column = lineLength - 1
	}

	return detailPosition{line: lineIndex, column: column}
}

func (document detailDocument) moveToNextWord(position detailPosition) detailPosition {
	if len(document.text) == 0 {
		return position
	}

	cursor := document.globalIndex(position)
	for cursor < len(document.text) && !unicode.IsSpace(document.text[cursor]) {
		cursor++
	}
	for cursor < len(document.text) && unicode.IsSpace(document.text[cursor]) {
		cursor++
	}
	if cursor >= len(document.text) {
		return document.lastPosition()
	}

	return document.positionForGlobalIndex(cursor)
}

func (document detailDocument) moveToWordEnd(position detailPosition) detailPosition {
	if len(document.text) == 0 {
		return position
	}

	cursor := document.globalIndex(position)
	if cursor >= len(document.text) {
		return document.lastPosition()
	}

	if unicode.IsSpace(document.text[cursor]) {
		for cursor < len(document.text) && unicode.IsSpace(document.text[cursor]) {
			cursor++
		}
		if cursor >= len(document.text) {
			return document.lastPosition()
		}
	} else if cursor+1 < len(document.text) && unicode.IsSpace(document.text[cursor+1]) {
		cursor++
		for cursor < len(document.text) && unicode.IsSpace(document.text[cursor]) {
			cursor++
		}
		if cursor >= len(document.text) {
			return document.lastPosition()
		}
	}

	for cursor+1 < len(document.text) && !unicode.IsSpace(document.text[cursor+1]) {
		cursor++
	}

	return document.positionForGlobalIndex(cursor)
}

func (document detailDocument) moveToPreviousWord(position detailPosition) detailPosition {
	if len(document.text) == 0 {
		return position
	}

	cursor := document.globalIndex(position)
	if cursor <= 0 {
		return document.firstPosition()
	}
	for cursor > 0 && unicode.IsSpace(document.text[cursor-1]) {
		cursor--
	}
	for cursor > 0 && !unicode.IsSpace(document.text[cursor-1]) {
		cursor--
	}

	return document.positionForGlobalIndex(cursor)
}

func (document detailDocument) comparePositions(left detailPosition, right detailPosition) int {
	left = document.clampPosition(left)
	right = document.clampPosition(right)
	if left.line < right.line {
		return -1
	}
	if left.line > right.line {
		return 1
	}
	if left.column < right.column {
		return -1
	}
	if left.column > right.column {
		return 1
	}

	return 0
}

func (document detailDocument) selectionText(start detailPosition, end detailPosition) string {
	if document.comparePositions(start, end) > 0 {
		start, end = end, start
	}
	start = document.clampPosition(start)
	end = document.clampPosition(end)

	if start.line == end.line {
		line := document.lines[start.line]
		if len(line) == 0 {
			return ""
		}
		return string(line[start.column : end.column+1])
	}

	var builder strings.Builder
	for lineIndex := start.line; lineIndex <= end.line; lineIndex++ {
		line := document.lines[lineIndex]
		switch {
		case lineIndex == start.line:
			if len(line) > 0 {
				builder.WriteString(string(line[start.column:]))
			}
		case lineIndex == end.line:
			if len(line) > 0 {
				builder.WriteString(string(line[:end.column+1]))
			}
		default:
			builder.WriteString(string(line))
		}
		if lineIndex < end.line {
			builder.WriteRune('\n')
		}
	}

	return builder.String()
}

func (document detailDocument) rowSelectionText(startRow int, endRow int) string {
	if len(document.rows) == 0 {
		return ""
	}
	if startRow > endRow {
		startRow, endRow = endRow, startRow
	}

	startRow = clampIndex(startRow, len(document.rows))
	endRow = clampIndex(endRow, len(document.rows))
	selectedRows := make([]string, 0, endRow-startRow+1)
	for rowIndex := startRow; rowIndex <= endRow; rowIndex++ {
		selectedRows = append(selectedRows, document.rows[rowIndex].text)
	}

	return strings.Join(selectedRows, "\n")
}
