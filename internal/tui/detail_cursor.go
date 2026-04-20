package tui

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

const (
	detailYankSuccessMessage = "󰆏 Selection copied"
	detailYankFailureMessage = "󰅚 Copy failed"
)

type detailMode int

const (
	detailNormalMode detailMode = iota
	detailVisualMode
)

type detailPosition struct {
	line   int
	column int
}

type detailColumnRange struct {
	start int
	end   int
}

type detailWrappedRow struct {
	line        int
	startColumn int
	endColumn   int
	empty       bool
	text        string
}

type detailDocument struct {
	text             []rune
	lines            [][]rune
	width            int
	lineStartOffsets []int
	lineStartRows    []int
	rows             []detailWrappedRow
}

type detailViewState struct {
	cursor          detailPosition
	originRow       int
	preferredColumn int
	mode            detailMode
	visualAnchor    detailPosition
	pendingGoToTop  bool
}

type detailCellStyle struct {
	selected bool
	search   bool
}

func newDetailDocument(text string, width int) detailDocument {
	if width < 1 {
		width = 1
	}

	lineTexts := strings.Split(text, "\n")
	if len(lineTexts) == 0 {
		lineTexts = []string{""}
	}

	document := detailDocument{
		text:             []rune(strings.Join(lineTexts, "\n")),
		lines:            make([][]rune, 0, len(lineTexts)),
		width:            width,
		lineStartOffsets: make([]int, 0, len(lineTexts)),
		lineStartRows:    make([]int, 0, len(lineTexts)),
	}

	offset := 0
	rowIndex := 0
	for lineIndex, lineText := range lineTexts {
		lineRunes := []rune(lineText)
		document.lines = append(document.lines, lineRunes)
		document.lineStartOffsets = append(document.lineStartOffsets, offset)
		document.lineStartRows = append(document.lineStartRows, rowIndex)

		if len(lineRunes) == 0 {
			document.rows = append(document.rows, detailWrappedRow{line: lineIndex, startColumn: 0, endColumn: 0, empty: true})
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
		if lineIndex < len(lineTexts)-1 {
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
	if lineLength == 0 {
		return document.lineStartRows[position.line]
	}

	return document.lineStartRows[position.line] + (position.column / document.width)
}

func (document detailDocument) screenColumnForPosition(position detailPosition) int {
	position = document.clampPosition(position)
	if document.lineLength(position.line) == 0 {
		return 0
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

func (document detailDocument) searchMatchRanges(query string) map[int][]detailColumnRange {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil
	}

	pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(trimmedQuery))
	matchRanges := map[int][]detailColumnRange{}
	for lineIndex, line := range document.lines {
		lineText := string(line)
		matches := pattern.FindAllStringIndex(lineText, -1)
		if len(matches) == 0 {
			continue
		}

		lineRanges := make([]detailColumnRange, 0, len(matches))
		for _, match := range matches {
			start := utf8.RuneCountInString(lineText[:match[0]])
			end := start + utf8.RuneCountInString(lineText[match[0]:match[1]])
			lineRanges = append(lineRanges, detailColumnRange{start: start, end: end})
		}
		matchRanges[lineIndex] = lineRanges
	}

	return matchRanges
}

func newDetailViewState() detailViewState {
	return detailViewState{}
}

func (state *detailViewState) sync(document detailDocument, viewportHeight int) {
	if viewportHeight < 1 {
		viewportHeight = 1
	}

	state.cursor = document.clampPosition(state.cursor)
	if state.mode == detailVisualMode {
		state.visualAnchor = document.clampPosition(state.visualAnchor)
	} else {
		state.visualAnchor = state.cursor
	}

	currentRow := document.rowIndexForPosition(state.cursor)
	maxOriginRow := maxInt(0, document.rowCount()-viewportHeight)
	if state.originRow > maxOriginRow {
		state.originRow = maxOriginRow
	}
	if state.originRow < 0 {
		state.originRow = 0
	}
	if currentRow < state.originRow {
		state.originRow = currentRow
	}
	if currentRow >= state.originRow+viewportHeight {
		state.originRow = currentRow - viewportHeight + 1
	}
	if state.originRow < 0 {
		state.originRow = 0
	}
	if state.originRow > maxOriginRow {
		state.originRow = maxOriginRow
	}
}

func (state *detailViewState) reset() {
	*state = detailViewState{}
}

func (state *detailViewState) clearPendingPrefix() {
	state.pendingGoToTop = false
}

func (state *detailViewState) enterVisualMode() {
	state.clearPendingPrefix()
	if state.mode == detailVisualMode {
		return
	}

	state.mode = detailVisualMode
	state.visualAnchor = state.cursor
}

func (state *detailViewState) exitVisualMode() {
	state.clearPendingPrefix()
	state.mode = detailNormalMode
	state.visualAnchor = state.cursor
}

func (state *detailViewState) moveLeft(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveLeft(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveRight(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveRight(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveDown(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, 1, state.preferredColumn)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveUp(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, -1, state.preferredColumn)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) pageDown(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, pageDelta(viewportHeight), state.preferredColumn)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) pageUp(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.sync(document, viewportHeight)
	state.cursor = document.moveVertical(state.cursor, -pageDelta(viewportHeight), state.preferredColumn)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToRowStart(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToRowStart(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToRowEnd(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToRowEnd(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) handleGoToTopPrefix(document detailDocument, viewportHeight int) {
	if state.pendingGoToTop {
		state.moveToTop(document, viewportHeight)
		return
	}

	state.pendingGoToTop = true
}

func (state *detailViewState) moveToTop(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToTop()
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToBottom(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToBottom()
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToNextWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToNextWord(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state *detailViewState) moveToPreviousWord(document detailDocument, viewportHeight int) {
	state.clearPendingPrefix()
	state.cursor = document.moveToPreviousWord(state.cursor)
	state.preferredColumn = document.screenColumnForPosition(state.cursor)
	state.sync(document, viewportHeight)
}

func (state detailViewState) visualSelection(document detailDocument) (detailPosition, detailPosition, bool) {
	if state.mode != detailVisualMode {
		return detailPosition{}, detailPosition{}, false
	}

	start := document.clampPosition(state.visualAnchor)
	end := document.clampPosition(state.cursor)
	if document.comparePositions(start, end) > 0 {
		start, end = end, start
	}

	return start, end, true
}

func (state detailViewState) selectedText(document detailDocument) string {
	start, end, ok := state.visualSelection(document)
	if !ok {
		return ""
	}

	return document.selectionText(start, end)
}

func (state detailViewState) isPositionSelected(document detailDocument, position detailPosition) bool {
	start, end, ok := state.visualSelection(document)
	if !ok {
		return false
	}

	return document.comparePositions(start, position) <= 0 && document.comparePositions(position, end) <= 0
}

func (state detailViewState) screenPosition(document detailDocument) (int, int) {
	return document.rowIndexForPosition(state.cursor), document.screenColumnForPosition(state.cursor)
}

func renderDetailRow(document detailDocument, row detailWrappedRow, searchMatchRanges map[int][]detailColumnRange, state detailViewState) string {
	if row.empty {
		return ""
	}

	line := document.lines[row.line]
	rowRunes := line[row.startColumn : row.endColumn+1]
	lineMatchRanges := searchMatchRanges[row.line]
	if len(lineMatchRanges) == 0 && state.mode != detailVisualMode {
		return row.text
	}

	var builder strings.Builder
	currentPrefix := ""
	for offset, character := range rowRunes {
		column := row.startColumn + offset
		prefix := detailCellStylePrefix(detailCellStyle{
			selected: state.isPositionSelected(document, detailPosition{line: row.line, column: column}),
			search:   detailColumnInRanges(column, lineMatchRanges),
		})
		if prefix != currentPrefix {
			if currentPrefix != "" {
				builder.WriteString(ansiReset)
			}
			if prefix != "" {
				builder.WriteString(prefix)
			}
			currentPrefix = prefix
		}
		builder.WriteRune(character)
	}
	if currentPrefix != "" {
		builder.WriteString(ansiReset)
	}

	return builder.String()
}

func detailCellStylePrefix(style detailCellStyle) string {
	if style.selected {
		return ansiBold + backgroundColorEscape(theme.SelectedLineBackgroundHex)
	}
	if style.search {
		return backgroundColorEscape(theme.SearchHighlightHex)
	}

	return ""
}

func detailColumnInRanges(column int, ranges []detailColumnRange) bool {
	for _, matchRange := range ranges {
		if column >= matchRange.start && column < matchRange.end {
			return true
		}
	}

	return false
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
