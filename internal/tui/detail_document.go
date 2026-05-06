package tui

import "strings"

func newDetailDocument(text string, width int) detailDocument {
	return newDetailDocumentWithWrap(text, width, true)
}

func newDetailDocumentWithWrap(text string, width int, wrap bool) detailDocument {
	if width < 1 {
		width = 1
	}

	styledLines := splitStyledTextLines(text)
	visibleLines := make([]string, 0, len(styledLines))
	for _, styledLine := range styledLines {
		visibleLines = append(visibleLines, string(styledLine.runes))
	}

	document := detailDocument{
		text:                 []rune(strings.Join(visibleLines, "\n")),
		lines:                make([][]rune, 0, len(styledLines)),
		lineStylePrefixes:    make([][]string, 0, len(styledLines)),
		lineHyperlinkTargets: make([][]string, 0, len(styledLines)),
		width:                width,
		wrap:                 wrap,
		lineStartOffsets:     make([]int, 0, len(styledLines)),
		lineStartRows:        make([]int, 0, len(styledLines)),
	}

	offset := 0
	rowIndex := 0
	for lineIndex, styledLine := range styledLines {
		lineRunes := append([]rune(nil), styledLine.runes...)
		lineStylePrefixes := append([]string(nil), styledLine.stylePrefixes...)
		lineHyperlinkTargets := append([]string(nil), styledLine.hyperlinkTargets...)
		document.lines = append(document.lines, lineRunes)
		document.lineStylePrefixes = append(document.lineStylePrefixes, lineStylePrefixes)
		document.lineHyperlinkTargets = append(document.lineHyperlinkTargets, lineHyperlinkTargets)
		document.lineStartOffsets = append(document.lineStartOffsets, offset)
		document.lineStartRows = append(document.lineStartRows, rowIndex)

		if len(lineRunes) == 0 {
			document.rows = append(document.rows, detailWrappedRow{line: lineIndex, startColumn: 0, endColumn: 0, empty: true})
			rowIndex++
		} else if !wrap {
			document.rows = append(document.rows, detailWrappedRow{line: lineIndex, startColumn: 0, endColumn: len(lineRunes) - 1, text: string(lineRunes)})
			rowIndex++
		} else {
			for startColumn := 0; startColumn < len(lineRunes); startColumn += width {
				endColumnExclusive := minInt(startColumn+width, len(lineRunes))
				document.rows = append(document.rows, detailWrappedRow{
					line:        lineIndex,
					startColumn: startColumn,
					endColumn:   endColumnExclusive - 1,
					text:        string(lineRunes[startColumn:endColumnExclusive]),
				})
				rowIndex++
			}
		}

		offset += len(lineRunes)
		if lineIndex < len(styledLines)-1 {
			offset++
		}
	}

	if len(document.rows) == 0 {
		document.rows = append(document.rows, detailWrappedRow{line: 0, startColumn: 0, endColumn: 0, empty: true})
	}

	return document
}

func (document detailDocument) lineCount() int {
	return len(document.lines)
}

func (document detailDocument) lineLength(line int) int {
	if line < 0 || line >= len(document.lines) {
		return 0
	}

	return len(document.lines[line])
}

func (document detailDocument) rowCount() int {
	return len(document.rows)
}

func (document detailDocument) firstPosition() detailPosition {
	return detailPosition{}
}

func (document detailDocument) lastPosition() detailPosition {
	if len(document.lines) == 0 {
		return detailPosition{}
	}

	lastLine := len(document.lines) - 1
	lastLineLength := len(document.lines[lastLine])
	if lastLineLength == 0 {
		return detailPosition{line: lastLine, column: 0}
	}

	return detailPosition{line: lastLine, column: lastLineLength - 1}
}

func (document detailDocument) clampPosition(position detailPosition) detailPosition {
	if len(document.lines) == 0 {
		return detailPosition{}
	}

	clampedLine := clampIndex(position.line, len(document.lines))
	lineLength := document.lineLength(clampedLine)
	if lineLength == 0 {
		return detailPosition{line: clampedLine, column: 0}
	}
	if position.column < 0 {
		return detailPosition{line: clampedLine, column: 0}
	}
	if position.column >= lineLength {
		return detailPosition{line: clampedLine, column: lineLength - 1}
	}

	return detailPosition{line: clampedLine, column: position.column}
}

func (document detailDocument) rowIndexForPosition(position detailPosition) int {
	position = document.clampPosition(position)
	if len(document.lineStartRows) == 0 {
		return 0
	}

	lineLength := document.lineLength(position.line)
	if lineLength == 0 || !document.wrap {
		return document.lineStartRows[position.line]
	}

	return document.lineStartRows[position.line] + (position.column / document.width)
}

func (document detailDocument) screenColumnForPosition(position detailPosition) int {
	position = document.clampPosition(position)
	if document.lineLength(position.line) == 0 {
		return 0
	}
	if !document.wrap {
		return position.column
	}

	return position.column % document.width
}

func (document detailDocument) positionForRow(rowIndex int, desiredColumn int) detailPosition {
	if len(document.rows) == 0 {
		return detailPosition{}
	}

	clampedRowIndex := clampIndex(rowIndex, len(document.rows))
	row := document.rows[clampedRowIndex]
	if row.empty {
		return detailPosition{line: row.line, column: 0}
	}

	rowLength := row.endColumn - row.startColumn + 1
	clampedColumn := desiredColumn
	if clampedColumn < 0 {
		clampedColumn = 0
	}
	if clampedColumn >= rowLength {
		clampedColumn = rowLength - 1
	}

	return detailPosition{line: row.line, column: row.startColumn + clampedColumn}
}

func (document detailDocument) moveVertical(position detailPosition, delta int, desiredColumn int) detailPosition {
	rowIndex := document.rowIndexForPosition(position)
	return document.positionForRow(rowIndex+delta, desiredColumn)
}

func (document detailDocument) moveLeft(position detailPosition) detailPosition {
	position = document.clampPosition(position)
	if position.line == 0 && position.column == 0 {
		return position
	}

	if document.lineLength(position.line) > 0 && position.column > 0 {
		return detailPosition{line: position.line, column: position.column - 1}
	}
	if position.line == 0 {
		return detailPosition{}
	}

	previousLine := position.line - 1
	previousLineLength := document.lineLength(previousLine)
	if previousLineLength == 0 {
		return detailPosition{line: previousLine, column: 0}
	}

	return detailPosition{line: previousLine, column: previousLineLength - 1}
}

func (document detailDocument) moveRight(position detailPosition) detailPosition {
	position = document.clampPosition(position)
	currentLineLength := document.lineLength(position.line)
	if currentLineLength == 0 {
		if position.line >= document.lineCount()-1 {
			return position
		}
		return detailPosition{line: position.line + 1, column: 0}
	}

	if position.column < currentLineLength-1 {
		return detailPosition{line: position.line, column: position.column + 1}
	}
	if position.line >= document.lineCount()-1 {
		return position
	}

	return detailPosition{line: position.line + 1, column: 0}
}

func (document detailDocument) moveToRowStart(position detailPosition) detailPosition {
	row := document.rows[document.rowIndexForPosition(position)]
	if row.empty {
		return detailPosition{line: row.line, column: 0}
	}

	return detailPosition{line: row.line, column: row.startColumn}
}

func (document detailDocument) moveToRowEnd(position detailPosition) detailPosition {
	row := document.rows[document.rowIndexForPosition(position)]
	if row.empty {
		return detailPosition{line: row.line, column: 0}
	}

	return detailPosition{line: row.line, column: row.endColumn}
}

func (document detailDocument) moveToTop() detailPosition {
	return document.positionForRow(0, 0)
}

func (document detailDocument) moveToBottom() detailPosition {
	return document.positionForRow(document.rowCount()-1, 0)
}
