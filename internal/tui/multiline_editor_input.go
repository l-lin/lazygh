package tui

import (
	"unicode"

	"github.com/jesseduffield/gocui"
)

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
