package tui

import (
	"unicode"

	"github.com/jesseduffield/gocui"
)

type lineEditor struct {
	text   []rune
	cursor int
}

func newLineEditor(text string) lineEditor {
	editor := lineEditor{}
	editor.SetText(text)
	return editor
}

func (editor *lineEditor) Text() string {
	if editor == nil {
		return ""
	}

	return string(editor.text)
}

func (editor *lineEditor) Cursor() int {
	if editor == nil {
		return 0
	}

	return editor.cursor
}

func (editor *lineEditor) SetText(text string) {
	if editor == nil {
		return
	}

	editor.text = []rune(text)
	editor.cursor = len(editor.text)
}

func (editor *lineEditor) HandleKey(key gocui.Key, ch rune, mod gocui.Modifier) bool {
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
	case key == gocui.KeyHome || key == gocui.KeyCtrlA:
		editor.MoveCursorToStart()
		return true
	case key == gocui.KeyEnd || key == gocui.KeyCtrlE:
		editor.MoveCursorToEnd()
		return true
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2 || key == gocui.KeyCtrlH:
		editor.DeleteBackwardChar()
		return true
	case key == gocui.KeyCtrlD:
		editor.DeleteForwardChar()
		return true
	case key == gocui.KeyCtrlW:
		editor.DeleteBackwardWord()
		return true
	case key == gocui.KeyCtrlU:
		editor.DeleteToStart()
		return true
	case key == gocui.KeyCtrlK:
		editor.DeleteToEnd()
		return true
	case (ch == 'd' || ch == 'D') && (mod&gocui.ModAlt) != 0:
		editor.DeleteForwardWord()
		return true
	case (ch == 'b' || ch == 'B') && (mod&gocui.ModAlt) != 0:
		editor.MoveCursorWordLeft()
		return true
	case (ch == 'f' || ch == 'F') && (mod&gocui.ModAlt) != 0:
		editor.MoveCursorWordRight()
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

func (editor *lineEditor) InsertRune(ch rune) {
	if editor == nil {
		return
	}

	if editor.cursor < 0 {
		editor.cursor = 0
	}
	if editor.cursor > len(editor.text) {
		editor.cursor = len(editor.text)
	}

	editor.text = append(editor.text[:editor.cursor], append([]rune{ch}, editor.text[editor.cursor:]...)...)
	editor.cursor++
}

func (editor *lineEditor) DeleteBackwardChar() {
	if editor == nil || editor.cursor == 0 || len(editor.text) == 0 {
		return
	}

	deleteIndex := editor.cursor - 1
	editor.text = append(editor.text[:deleteIndex], editor.text[editor.cursor:]...)
	editor.cursor = deleteIndex
}

func (editor *lineEditor) DeleteForwardChar() {
	if editor == nil || editor.cursor >= len(editor.text) || len(editor.text) == 0 {
		return
	}

	editor.text = append(editor.text[:editor.cursor], editor.text[editor.cursor+1:]...)
}

func (editor *lineEditor) DeleteBackwardWord() {
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
}

func (editor *lineEditor) DeleteForwardWord() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	end := editor.cursor
	for end < len(editor.text) && unicode.IsSpace(editor.text[end]) {
		end++
	}
	for end < len(editor.text) && !unicode.IsSpace(editor.text[end]) {
		end++
	}

	editor.text = append(editor.text[:editor.cursor], editor.text[end:]...)
}

func (editor *lineEditor) DeleteToStart() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	editor.text = append([]rune{}, editor.text[editor.cursor:]...)
	editor.cursor = 0
}

func (editor *lineEditor) DeleteToEnd() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	editor.text = append([]rune{}, editor.text[:editor.cursor]...)
}

func (editor *lineEditor) MoveCursorLeft() {
	if editor == nil || editor.cursor == 0 {
		return
	}

	editor.cursor--
}

func (editor *lineEditor) MoveCursorRight() {
	if editor == nil || editor.cursor >= len(editor.text) {
		return
	}

	editor.cursor++
}

func (editor *lineEditor) MoveCursorToStart() {
	if editor == nil {
		return
	}

	editor.cursor = 0
}

func (editor *lineEditor) MoveCursorToEnd() {
	if editor == nil {
		return
	}

	editor.cursor = len(editor.text)
}

func (editor *lineEditor) MoveCursorWordLeft() {
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
}

func (editor *lineEditor) MoveCursorWordRight() {
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
}

func (editor lineEditor) clone() lineEditor {
	editor.text = append([]rune(nil), editor.text...)
	return editor
}
