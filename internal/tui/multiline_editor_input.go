package tui

import (
	"unicode"

	"github.com/jesseduffield/gocui"
)

type multilineEditorIntentKind int

const (
	multilineEditorIntentKindNone multilineEditorIntentKind = iota
	multilineEditorIntentKindMoveCursorLeft
	multilineEditorIntentKindMoveCursorRight
	multilineEditorIntentKindMoveCursorUp
	multilineEditorIntentKindMoveCursorDown
	multilineEditorIntentKindMoveCursorToLineStart
	multilineEditorIntentKindMoveCursorToLineEnd
	multilineEditorIntentKindDeleteBackwardChar
	multilineEditorIntentKindDeleteForwardChar
	multilineEditorIntentKindDeleteBackwardWord
	multilineEditorIntentKindDeleteToLineStart
	multilineEditorIntentKindDeleteToLineEnd
	multilineEditorIntentKindDeleteForwardWord
	multilineEditorIntentKindMoveCursorWordLeft
	multilineEditorIntentKindMoveCursorWordRight
	multilineEditorIntentKindInsertCodeFence
	multilineEditorIntentKindInsertRune
)

type multilineEditorIntent struct {
	kind multilineEditorIntentKind
	ch   rune
}

func newMultilineEditorInsertRuneIntent(ch rune) multilineEditorIntent {
	return multilineEditorIntent{kind: multilineEditorIntentKindInsertRune, ch: ch}
}

func multilineEditorIntentFromKey(key gocui.Key, ch rune, mod gocui.Modifier) (multilineEditorIntent, bool) {
	switch {
	case key == gocui.KeyArrowLeft || key == gocui.KeyCtrlB:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorLeft}, true
	case key == gocui.KeyArrowRight || key == gocui.KeyCtrlF:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorRight}, true
	case key == gocui.KeyArrowUp:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorUp}, true
	case key == gocui.KeyArrowDown:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorDown}, true
	case key == gocui.KeyHome || key == gocui.KeyCtrlA:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorToLineStart}, true
	case key == gocui.KeyEnd || key == gocui.KeyCtrlE:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorToLineEnd}, true
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2 || key == gocui.KeyCtrlH:
		return multilineEditorIntent{kind: multilineEditorIntentKindDeleteBackwardChar}, true
	case key == gocui.KeyCtrlD:
		return multilineEditorIntent{kind: multilineEditorIntentKindDeleteForwardChar}, true
	case key == gocui.KeyCtrlW:
		return multilineEditorIntent{kind: multilineEditorIntentKindDeleteBackwardWord}, true
	case key == gocui.KeyCtrlU:
		return multilineEditorIntent{kind: multilineEditorIntentKindDeleteToLineStart}, true
	case key == gocui.KeyCtrlK:
		return multilineEditorIntent{kind: multilineEditorIntentKindDeleteToLineEnd}, true
	case (ch == 'd' || ch == 'D') && (mod&gocui.ModAlt) != 0:
		return multilineEditorIntent{kind: multilineEditorIntentKindDeleteForwardWord}, true
	case (ch == 'b' || ch == 'B') && (mod&gocui.ModAlt) != 0:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorWordLeft}, true
	case (ch == 'f' || ch == 'F') && (mod&gocui.ModAlt) != 0:
		return multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorWordRight}, true
	case (ch == 'c' || ch == 'C') && (mod&gocui.ModAlt) != 0:
		return multilineEditorIntent{kind: multilineEditorIntentKindInsertCodeFence}, true
	case key == gocui.KeyEnter || key == gocui.KeyCtrlJ:
		return newMultilineEditorInsertRuneIntent('\n'), true
	case key == gocui.KeySpace:
		return newMultilineEditorInsertRuneIntent(' '), true
	case ch != 0 && mod == gocui.ModNone:
		return newMultilineEditorInsertRuneIntent(ch), true
	default:
		return multilineEditorIntent{}, false
	}
}

func (editor *multilineEditor) ApplyIntent(intent multilineEditorIntent) bool {
	if editor == nil {
		return false
	}

	switch intent.kind {
	case multilineEditorIntentKindMoveCursorLeft:
		editor.MoveCursorLeft()
		return true
	case multilineEditorIntentKindMoveCursorRight:
		editor.MoveCursorRight()
		return true
	case multilineEditorIntentKindMoveCursorUp:
		editor.MoveCursorUp()
		return true
	case multilineEditorIntentKindMoveCursorDown:
		editor.MoveCursorDown()
		return true
	case multilineEditorIntentKindMoveCursorToLineStart:
		editor.MoveCursorToLineStart()
		return true
	case multilineEditorIntentKindMoveCursorToLineEnd:
		editor.MoveCursorToLineEnd()
		return true
	case multilineEditorIntentKindDeleteBackwardChar:
		editor.DeleteBackwardChar()
		return true
	case multilineEditorIntentKindDeleteForwardChar:
		editor.DeleteForwardChar()
		return true
	case multilineEditorIntentKindDeleteBackwardWord:
		editor.DeleteBackwardWord()
		return true
	case multilineEditorIntentKindDeleteToLineStart:
		editor.DeleteToLineStart()
		return true
	case multilineEditorIntentKindDeleteToLineEnd:
		editor.DeleteToLineEnd()
		return true
	case multilineEditorIntentKindDeleteForwardWord:
		editor.DeleteForwardWord()
		return true
	case multilineEditorIntentKindMoveCursorWordLeft:
		editor.MoveCursorWordLeft()
		return true
	case multilineEditorIntentKindMoveCursorWordRight:
		editor.MoveCursorWordRight()
		return true
	case multilineEditorIntentKindInsertCodeFence:
		editor.InsertCodeFence()
		return true
	case multilineEditorIntentKindInsertRune:
		editor.InsertRune(intent.ch)
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

func (editor *multilineEditor) DeleteForwardChar() {
	if editor == nil || editor.cursor >= len(editor.text) || len(editor.text) == 0 {
		return
	}

	editor.text = append(editor.text[:editor.cursor], editor.text[editor.cursor+1:]...)
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

func (editor *multilineEditor) DeleteForwardWord() {
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
