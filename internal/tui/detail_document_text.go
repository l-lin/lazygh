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

type wordMotionKind int

const (
	wordMotionSmall wordMotionKind = iota
	wordMotionBig
)

const (
	wordMotionWhitespaceClass = iota
	wordMotionKeywordClass
	wordMotionPunctuationClass
)

func (document detailDocument) moveToNextWord(position detailPosition) detailPosition {
	return document.moveToNextWordWithKind(position, wordMotionSmall)
}

func (document detailDocument) moveToNextBigWord(position detailPosition) detailPosition {
	return document.moveToNextWordWithKind(position, wordMotionBig)
}

func (document detailDocument) moveToNextWordWithKind(position detailPosition, kind wordMotionKind) detailPosition {
	if len(document.text) == 0 {
		return position
	}

	cursor := document.globalIndex(position)
	if cursor >= len(document.text) {
		return document.lastPosition()
	}

	currentClass := wordMotionClass(document.text[cursor], kind)
	if currentClass == wordMotionWhitespaceClass {
		for cursor < len(document.text) && wordMotionClass(document.text[cursor], kind) == wordMotionWhitespaceClass {
			cursor++
		}
	} else {
		for cursor < len(document.text) && wordMotionClass(document.text[cursor], kind) == currentClass {
			cursor++
		}
		for cursor < len(document.text) && wordMotionClass(document.text[cursor], kind) == wordMotionWhitespaceClass {
			cursor++
		}
	}
	if cursor >= len(document.text) {
		return document.lastPosition()
	}

	return document.positionForGlobalIndex(cursor)
}

func (document detailDocument) moveToWordEnd(position detailPosition) detailPosition {
	return document.moveToWordEndWithKind(position, wordMotionSmall)
}

func (document detailDocument) moveToBigWordEnd(position detailPosition) detailPosition {
	return document.moveToWordEndWithKind(position, wordMotionBig)
}

func (document detailDocument) moveToWordEndWithKind(position detailPosition, kind wordMotionKind) detailPosition {
	if len(document.text) == 0 {
		return position
	}

	cursor := document.globalIndex(position)
	if cursor >= len(document.text) {
		return document.lastPosition()
	}

	currentClass := wordMotionClass(document.text[cursor], kind)
	if currentClass == wordMotionWhitespaceClass {
		for cursor < len(document.text) && wordMotionClass(document.text[cursor], kind) == wordMotionWhitespaceClass {
			cursor++
		}
		if cursor >= len(document.text) {
			return document.lastPosition()
		}
	} else if cursor+1 >= len(document.text) || wordMotionClass(document.text[cursor+1], kind) != currentClass {
		cursor++
		for cursor < len(document.text) && wordMotionClass(document.text[cursor], kind) == wordMotionWhitespaceClass {
			cursor++
		}
		if cursor >= len(document.text) {
			return document.lastPosition()
		}
	}

	currentClass = wordMotionClass(document.text[cursor], kind)
	for cursor+1 < len(document.text) && wordMotionClass(document.text[cursor+1], kind) == currentClass {
		cursor++
	}

	return document.positionForGlobalIndex(cursor)
}

func (document detailDocument) moveToPreviousWord(position detailPosition) detailPosition {
	return document.moveToPreviousWordWithKind(position, wordMotionSmall)
}

func (document detailDocument) moveToPreviousBigWord(position detailPosition) detailPosition {
	return document.moveToPreviousWordWithKind(position, wordMotionBig)
}

func (document detailDocument) moveToPreviousWordWithKind(position detailPosition, kind wordMotionKind) detailPosition {
	if len(document.text) == 0 {
		return position
	}

	cursor := document.globalIndex(position)
	if cursor <= 0 {
		return document.firstPosition()
	}

	cursor--
	for cursor >= 0 && wordMotionClass(document.text[cursor], kind) == wordMotionWhitespaceClass {
		cursor--
	}
	if cursor < 0 {
		return document.firstPosition()
	}

	currentClass := wordMotionClass(document.text[cursor], kind)
	for cursor > 0 && wordMotionClass(document.text[cursor-1], kind) == currentClass {
		cursor--
	}

	return document.positionForGlobalIndex(cursor)
}

func wordMotionClass(value rune, kind wordMotionKind) int {
	if unicode.IsSpace(value) {
		return wordMotionWhitespaceClass
	}
	if kind == wordMotionBig {
		return wordMotionKeywordClass
	}
	if unicode.IsLetter(value) || unicode.IsNumber(value) || value == '_' {
		return wordMotionKeywordClass
	}
	return wordMotionPunctuationClass
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
