package tui

import (
	"strings"
	"sync/atomic"
)

var detailDocumentSequence atomic.Uint64

type detailDocumentLine struct {
	prefix styledTextLine
	body   styledTextLine
}

func newDetailDocument(text string, width int) detailDocument {
	return newDetailDocumentWithWrap(text, width, true)
}

func newDetailDocumentWithWrap(text string, width int, wrap bool) detailDocument {
	styledLines := splitStyledTextLines(text)
	lines := make([]detailDocumentLine, 0, len(styledLines))
	for _, styledLine := range styledLines {
		lines = append(lines, detailDocumentLine{body: cloneStyledTextLine(styledLine)})
	}
	return newDetailDocumentFromLines(lines, width, wrap)
}

func newDetailDocumentFromLines(lines []detailDocumentLine, width int, wrap bool) detailDocument {
	if width < 1 {
		width = 1
	}

	visibleLines := make([]string, 0, len(lines))
	for _, line := range lines {
		visibleLines = append(visibleLines, string(line.body.runes))
	}

	document := detailDocument{
		id:                   detailDocumentSequence.Add(1),
		text:                 []rune(strings.Join(visibleLines, "\n")),
		prefixLines:          make([]styledTextLine, 0, len(lines)),
		lines:                make([][]rune, 0, len(lines)),
		lineStylePrefixes:    make([][]string, 0, len(lines)),
		lineHyperlinkTargets: make([][]string, 0, len(lines)),
		images:               make([]detailImagePlacement, 0),
		width:                width,
		wrap:                 wrap,
		lineStartOffsets:     make([]int, 0, len(lines)),
		lineStartRows:        make([]int, 0, len(lines)),
	}

	offset := 0
	rowIndex := 0
	for lineIndex, line := range lines {
		prefixLine := cloneStyledTextLine(line.prefix)
		bodyLine := cloneStyledTextLine(line.body)
		lineRunes := append([]rune(nil), bodyLine.runes...)
		lineStylePrefixes := append([]string(nil), bodyLine.stylePrefixes...)
		lineHyperlinkTargets := append([]string(nil), bodyLine.hyperlinkTargets...)
		document.prefixLines = append(document.prefixLines, prefixLine)
		document.lines = append(document.lines, lineRunes)
		document.lineStylePrefixes = append(document.lineStylePrefixes, lineStylePrefixes)
		document.lineHyperlinkTargets = append(document.lineHyperlinkTargets, lineHyperlinkTargets)
		document.lineStartOffsets = append(document.lineStartOffsets, offset)
		document.lineStartRows = append(document.lineStartRows, rowIndex)
		for _, control := range bodyLine.controls {
			if control.image == nil {
				continue
			}
			document.images = append(document.images, detailImagePlacement{line: lineIndex, column: control.column, imageID: control.image.imageID, columns: control.image.columns, rows: control.image.rows})
		}

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
		if lineIndex < len(lines)-1 {
			offset++
		}
	}

	if len(document.rows) == 0 {
		document.rows = append(document.rows, detailWrappedRow{line: 0, startColumn: 0, endColumn: 0, empty: true})
	}

	return document
}

func cloneStyledTextLine(line styledTextLine) styledTextLine {
	clonedLine := styledTextLine{
		runes:            append([]rune(nil), line.runes...),
		stylePrefixes:    append([]string(nil), line.stylePrefixes...),
		hyperlinkTargets: append([]string(nil), line.hyperlinkTargets...),
		controls:         make([]styledTextControl, 0, len(line.controls)),
	}
	for _, control := range line.controls {
		clonedLine.controls = append(clonedLine.controls, styledTextControl{column: control.column, image: control.image})
	}
	return clonedLine
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

func (document detailDocument) prefixWidthForLine(line int) int {
	if line < 0 || line >= len(document.prefixLines) {
		return 0
	}
	return len(document.prefixLines[line].runes)
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

func (document detailDocument) visualScreenColumnForPosition(position detailPosition) int {
	position = document.clampPosition(position)
	return document.prefixWidthForLine(position.line) + document.screenColumnForPosition(position)
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
	clampedColumn := max(desiredColumn, 0)
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
