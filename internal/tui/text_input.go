package tui

import (
	"unicode"

	"github.com/jesseduffield/gocui"
)

type lineEditorIntentKind int

const (
	lineEditorIntentKindNone lineEditorIntentKind = iota
	lineEditorIntentKindMoveCursorLeft
	lineEditorIntentKindMoveCursorRight
	lineEditorIntentKindMoveCursorToStart
	lineEditorIntentKindMoveCursorToEnd
	lineEditorIntentKindDeleteBackwardChar
	lineEditorIntentKindDeleteForwardChar
	lineEditorIntentKindDeleteBackwardWord
	lineEditorIntentKindDeleteToStart
	lineEditorIntentKindDeleteToEnd
	lineEditorIntentKindDeleteForwardWord
	lineEditorIntentKindMoveCursorWordLeft
	lineEditorIntentKindMoveCursorWordRight
	lineEditorIntentKindInsertRune
)

type lineEditorIntent struct {
	kind lineEditorIntentKind
	ch   rune
}

type lineEditor struct {
	text   []rune
	cursor int
}

func newLineEditor(text string) lineEditor {
	editor := lineEditor{}
	editor.SetText(text)
	return editor
}

func newLineEditorInsertRuneIntent(ch rune) lineEditorIntent {
	return lineEditorIntent{kind: lineEditorIntentKindInsertRune, ch: ch}
}

func lineEditorIntentFromKey(key gocui.Key, ch rune, mod gocui.Modifier) (lineEditorIntent, bool) {
	switch {
	case key == gocui.KeyArrowLeft || key == gocui.KeyCtrlB:
		return lineEditorIntent{kind: lineEditorIntentKindMoveCursorLeft}, true
	case key == gocui.KeyArrowRight || key == gocui.KeyCtrlF:
		return lineEditorIntent{kind: lineEditorIntentKindMoveCursorRight}, true
	case key == gocui.KeyHome || key == gocui.KeyCtrlA:
		return lineEditorIntent{kind: lineEditorIntentKindMoveCursorToStart}, true
	case key == gocui.KeyEnd || key == gocui.KeyCtrlE:
		return lineEditorIntent{kind: lineEditorIntentKindMoveCursorToEnd}, true
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2 || key == gocui.KeyCtrlH:
		return lineEditorIntent{kind: lineEditorIntentKindDeleteBackwardChar}, true
	case key == gocui.KeyCtrlD:
		return lineEditorIntent{kind: lineEditorIntentKindDeleteForwardChar}, true
	case key == gocui.KeyCtrlW:
		return lineEditorIntent{kind: lineEditorIntentKindDeleteBackwardWord}, true
	case key == gocui.KeyCtrlU:
		return lineEditorIntent{kind: lineEditorIntentKindDeleteToStart}, true
	case key == gocui.KeyCtrlK:
		return lineEditorIntent{kind: lineEditorIntentKindDeleteToEnd}, true
	case (ch == 'd' || ch == 'D') && (mod&gocui.ModAlt) != 0:
		return lineEditorIntent{kind: lineEditorIntentKindDeleteForwardWord}, true
	case (ch == 'b' || ch == 'B') && (mod&gocui.ModAlt) != 0:
		return lineEditorIntent{kind: lineEditorIntentKindMoveCursorWordLeft}, true
	case (ch == 'f' || ch == 'F') && (mod&gocui.ModAlt) != 0:
		return lineEditorIntent{kind: lineEditorIntentKindMoveCursorWordRight}, true
	case key == gocui.KeySpace:
		return newLineEditorInsertRuneIntent(' '), true
	case ch != 0 && mod == gocui.ModNone:
		return newLineEditorInsertRuneIntent(ch), true
	default:
		return lineEditorIntent{}, false
	}
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

func (editor *lineEditor) ApplyIntent(intent lineEditorIntent) bool {
	if editor == nil {
		return false
	}

	switch intent.kind {
	case lineEditorIntentKindMoveCursorLeft:
		editor.MoveCursorLeft()
		return true
	case lineEditorIntentKindMoveCursorRight:
		editor.MoveCursorRight()
		return true
	case lineEditorIntentKindMoveCursorToStart:
		editor.MoveCursorToStart()
		return true
	case lineEditorIntentKindMoveCursorToEnd:
		editor.MoveCursorToEnd()
		return true
	case lineEditorIntentKindDeleteBackwardChar:
		editor.DeleteBackwardChar()
		return true
	case lineEditorIntentKindDeleteForwardChar:
		editor.DeleteForwardChar()
		return true
	case lineEditorIntentKindDeleteBackwardWord:
		editor.DeleteBackwardWord()
		return true
	case lineEditorIntentKindDeleteToStart:
		editor.DeleteToStart()
		return true
	case lineEditorIntentKindDeleteToEnd:
		editor.DeleteToEnd()
		return true
	case lineEditorIntentKindDeleteForwardWord:
		editor.DeleteForwardWord()
		return true
	case lineEditorIntentKindMoveCursorWordLeft:
		editor.MoveCursorWordLeft()
		return true
	case lineEditorIntentKindMoveCursorWordRight:
		editor.MoveCursorWordRight()
		return true
	case lineEditorIntentKindInsertRune:
		editor.InsertRune(intent.ch)
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
