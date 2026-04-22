package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewModalEditorName        = "modal-editor"
	modalEditorTotalHeight     = 7
	lineModalEditorTotalHeight = 3
	modalEditorFallbackWidth   = 60
	modalEditorMinWidth        = 40
)

type modalEditorKeyHandler func(*Program, *gocui.View, gocui.Key, rune, gocui.Modifier) bool

type modalEditorState struct {
	title        string
	editor       *multilineEditor
	lineEditor   *lineEditor
	errorMessage string
	submit       func(string) error
	afterSubmit  func(*gocui.Gui)
	totalHeight  int
	handleKey    modalEditorKeyHandler
}

func newModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	return newModalEditorStateWithOptions(title, initialText, submit, modalEditorTotalHeight, false, nil)
}

func newModalEditorStateWithKeyHandler(title string, initialText string, submit func(string) error, handleKey modalEditorKeyHandler) *modalEditorState {
	return newModalEditorStateWithOptions(title, initialText, submit, modalEditorTotalHeight, false, handleKey)
}

func newLineModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	return newModalEditorStateWithOptions(title, initialText, submit, lineModalEditorTotalHeight, true, nil)
}

func newModalEditorStateWithOptions(title string, initialText string, submit func(string) error, totalHeight int, singleLine bool, handleKey modalEditorKeyHandler) *modalEditorState {
	if submit == nil {
		submit = func(string) error { return nil }
	}

	state := &modalEditorState{
		title:       strings.TrimSpace(title),
		submit:      submit,
		totalHeight: totalHeight,
		handleKey:   handleKey,
	}
	if singleLine {
		state.lineEditor = newLineEditor(initialText)
		return state
	}

	state.editor = newMultilineEditor(initialText)
	return state
}

func (state *modalEditorState) Text() string {
	if state == nil {
		return ""
	}
	if state.lineEditor != nil {
		return state.lineEditor.Text()
	}
	if state.editor != nil {
		return state.editor.Text()
	}
	return ""
}

func (state *modalEditorState) CursorXY() (int, int) {
	if state == nil {
		return 0, 0
	}
	if state.lineEditor != nil {
		return state.lineEditor.Cursor(), 0
	}
	if state.editor != nil {
		return state.editor.CursorXY()
	}
	return 0, 0
}

func (state *modalEditorState) HandleKey(key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if state == nil {
		return false
	}
	if state.lineEditor != nil {
		return state.lineEditor.HandleKey(key, ch, mod)
	}
	if state.editor != nil {
		return state.editor.HandleKey(key, ch, mod)
	}
	return false
}

func (state *modalEditorState) Height() int {
	if state == nil || state.totalHeight < 1 {
		return modalEditorTotalHeight
	}
	return state.totalHeight
}
