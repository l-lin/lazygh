package tui

import (
	"unicode"

	"github.com/jesseduffield/gocui"
)

const codeFenceSnippet = "```\n```"

type multilineEditor struct {
	text            []rune
	cursor          int
	preferredColumn int
}

func newMultilineEditor(text string) *multilineEditor {
	editor := &multilineEditor{preferredColumn: -1}
	editor.SetText(text)
	return editor
}

func (editor *multilineEditor) Text() string {
	if editor == nil {
		return ""
	}

	return string(editor.text)
}

func (editor *multilineEditor) Cursor() int {
	if editor == nil {
		return 0
	}

	return editor.cursor
}

func (editor *multilineEditor) CursorXY() (int, int) {
	if editor == nil {
		return 0, 0
	}

	column := 0
	row := 0
	for index := 0; index < editor.cursor && index < len(editor.text); index++ {
		if editor.text[index] == '\n' {
			row++
			column = 0
			continue
		}

		column++
	}

	return column, row
}

func (editor *multilineEditor) SetText(text string) {
	if editor == nil {
		return
	}

	editor.text = []rune(text)
	editor.cursor = len(editor.text)
	editor.preferredColumn = -1
}

func (editor *multilineEditor) HandleKey(key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if editor == nil {
		return false
	}

	switch {
	case key == gocui.KeyArrowLeft || key == gocui.KeyCtrlB:
		editor.MoveCursorLeft()
		return true
	case key == gocui.KeyArrowRight || key == gocui.KeyCtrlF:
		editor.MoveCursorRight()
		return true
	case key == gocui.KeyArrowUp:
		editor.MoveCursorUp()
		return true
	case key == gocui.KeyArrowDown:
		editor.MoveCursorDown()
		return true
	case key == gocui.KeyHome || key == gocui.KeyCtrlA:
		editor.MoveCursorToLineStart()
		return true
	case key == gocui.KeyEnd || key == gocui.KeyCtrlE:
		editor.MoveCursorToLineEnd()
		return true
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2 || key == gocui.KeyCtrlH:
		editor.DeleteBackwardChar()
		return true
	case key == gocui.KeyCtrlW:
		editor.DeleteBackwardWord()
		return true
	case key == gocui.KeyCtrlU:
		editor.DeleteToLineStart()
		return true
	case key == gocui.KeyCtrlK:
		editor.DeleteToLineEnd()
		return true
	case (ch == 'b' || ch == 'B') && (mod&gocui.ModAlt) != 0:
		editor.MoveCursorWordLeft()
		return true
	case (ch == 'f' || ch == 'F') && (mod&gocui.ModAlt) != 0:
		editor.MoveCursorWordRight()
		return true
	case (ch == 'c' || ch == 'C') && (mod&gocui.ModAlt) != 0:
		editor.InsertCodeFence()
		return true
	case key == gocui.KeyEnter || key == gocui.KeyCtrlJ:
		editor.InsertRune('\n')
		return true
	case key == gocui.KeySpace:
		editor.InsertRune(' ')
		return true
	case ch != 0 && mod == gocui.ModNone:
		editor.InsertRune(ch)
		return true
	default:
		return false
	}
}

func (editor *multilineEditor) InsertRune(ch rune) {
	if editor == nil {
		return
	}

	editor.normalizeCursor()
	editor.text = append(editor.text[:editor.cursor], append([]rune{ch}, editor.text[editor.cursor:]...)...)
	editor.cursor++
	editor.preferredColumn = -1
}

func (editor *multilineEditor) InsertCodeFence() {
	editor.InsertTextAndMoveCursor(codeFenceSnippet, len([]rune("```")))
}

func (editor *multilineEditor) InsertTextAndMoveCursor(text string, cursorOffset int) {
	if editor == nil {
		return
	}

	editor.normalizeCursor()
	inserted := []rune(text)
	insertAt := editor.cursor
	editor.text = append(editor.text[:insertAt], append(inserted, editor.text[insertAt:]...)...)
	editor.cursor = insertAt + cursorOffset
	editor.preferredColumn = -1
}

func (editor *multilineEditor) DeleteBackwardChar() {
	if editor == nil || editor.cursor == 0 || len(editor.text) == 0 {
		return
	}

	deleteIndex := editor.cursor - 1
	editor.text = append(editor.text[:deleteIndex], editor.text[editor.cursor:]...)
	editor.cursor = deleteIndex
	editor.preferredColumn = -1
}

func (editor *multilineEditor) DeleteBackwardWord() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	start := editor.cursor
	for start > 0 && unicode.IsSpace(editor.text[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(editor.text[start-1]) {
		start--
	}

	editor.text = append(editor.text[:start], editor.text[editor.cursor:]...)
	editor.cursor = start
	editor.preferredColumn = -1
}

func (editor *multilineEditor) DeleteToLineStart() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	start := editor.lineStart(editor.cursor)
	editor.text = append(editor.text[:start], editor.text[editor.cursor:]...)
	editor.cursor = start
	editor.preferredColumn = -1
}

func (editor *multilineEditor) DeleteToLineEnd() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	end := editor.lineEnd(editor.cursor)
	editor.text = append(editor.text[:editor.cursor], editor.text[end:]...)
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorLeft() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	editor.cursor--
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorRight() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	editor.cursor++
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorUp() {
	if editor == nil {
		return
	}

	currentLineStart := editor.lineStart(editor.cursor)
	if currentLineStart == 0 {
		editor.cursor = 0
		return
	}

	desiredColumn := editor.verticalDesiredColumn()
	previousLineEnd := currentLineStart - 1
	previousLineStart := editor.lineStart(previousLineEnd)
	previousLineLength := previousLineEnd - previousLineStart
	editor.cursor = previousLineStart + min(desiredColumn, previousLineLength)
}

func (editor *multilineEditor) MoveCursorDown() {
	if editor == nil {
		return
	}

	currentLineEnd := editor.lineEnd(editor.cursor)
	if currentLineEnd >= len(editor.text) {
		editor.cursor = len(editor.text)
		return
	}

	desiredColumn := editor.verticalDesiredColumn()
	nextLineStart := currentLineEnd + 1
	nextLineEnd := editor.lineEnd(nextLineStart)
	nextLineLength := nextLineEnd - nextLineStart
	editor.cursor = nextLineStart + min(desiredColumn, nextLineLength)
}

func (editor *multilineEditor) MoveCursorToLineStart() {
	if editor == nil {
		return
	}

	editor.cursor = editor.lineStart(editor.cursor)
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorToLineEnd() {
	if editor == nil {
		return
	}

	editor.cursor = editor.lineEnd(editor.cursor)
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorWordLeft() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	cursor := editor.cursor
	for cursor > 0 && unicode.IsSpace(editor.text[cursor-1]) {
		cursor--
	}
	for cursor > 0 && !unicode.IsSpace(editor.text[cursor-1]) {
		cursor--
	}
	editor.cursor = cursor
	editor.preferredColumn = -1
}

func (editor *multilineEditor) MoveCursorWordRight() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	cursor := editor.cursor
	for cursor < len(editor.text) && unicode.IsSpace(editor.text[cursor]) {
		cursor++
	}
	for cursor < len(editor.text) && !unicode.IsSpace(editor.text[cursor]) {
		cursor++
	}
	editor.cursor = cursor
	editor.preferredColumn = -1
}

func (editor *multilineEditor) verticalDesiredColumn() int {
	if editor.preferredColumn >= 0 {
		return editor.preferredColumn
	}

	column, _ := editor.CursorXY()
	editor.preferredColumn = column
	return editor.preferredColumn
}

func (editor *multilineEditor) lineStart(index int) int {
	if editor == nil {
		return 0
	}

	if index < 0 {
		index = 0
	}
	if index > len(editor.text) {
		index = len(editor.text)
	}

	start := index
	for start > 0 && editor.text[start-1] != '\n' {
		start--
	}

	return start
}

func (editor *multilineEditor) lineEnd(index int) int {
	if editor == nil {
		return 0
	}

	if index < 0 {
		index = 0
	}
	if index > len(editor.text) {
		index = len(editor.text)
	}

	end := index
	for end < len(editor.text) && editor.text[end] != '\n' {
		end++
	}

	return end
}

func (editor *multilineEditor) normalizeCursor() {
	if editor.cursor < 0 {
		editor.cursor = 0
	}
	if editor.cursor > len(editor.text) {
		editor.cursor = len(editor.text)
	}
}
