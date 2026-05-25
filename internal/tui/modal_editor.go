package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	viewModalEditorName         = "modal-editor"
	modalEditorTotalHeight      = 7
	lineModalEditorTotalHeight  = 3
	modalEditorFallbackWidth    = 60
	modalEditorMinWidth         = 40
	modalEditorSubmitAction     = "submit"
	modalEditorSubmitLineAction = "submit_line"
)

type modalEditorState struct {
	title           string
	editor          *multilineEditor
	lineEditor      *lineEditor
	errorMessage    string
	submit          func(string) error
	submitRequested func(string) Msg
	afterSubmit     func(*gocui.Gui)
	totalHeight     int
}

func newModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	return newMultilineModalEditorState(title, initialText, submit, modalEditorTotalHeight)
}

func newModalEditorStateWithSubmitRequested(title string, initialText string, submitRequested func(string) Msg) *modalEditorState {
	state := newModalEditorState(title, initialText, nil)
	state.submitRequested = submitRequested
	return state
}

func newLineModalEditorState(title string, initialText string, submit func(string) error) *modalEditorState {
	return newLineModalEditorStateWithHeight(title, initialText, submit, lineModalEditorTotalHeight)
}

func newLineModalEditorStateWithSubmitRequested(title string, initialText string, submitRequested func(string) Msg) *modalEditorState {
	return newLineModalEditorStateWithHeightAndSubmitRequested(title, initialText, submitRequested, lineModalEditorTotalHeight)
}

func newLineModalEditorStateWithHeight(title string, initialText string, submit func(string) error, totalHeight int) *modalEditorState {
	if submit == nil {
		submit = func(string) error { return nil }
	}
	if totalHeight < 1 {
		totalHeight = lineModalEditorTotalHeight
	}

	return &modalEditorState{
		title:       strings.TrimSpace(title),
		lineEditor:  newLineEditor(initialText),
		submit:      submit,
		totalHeight: totalHeight,
	}
}

func newLineModalEditorStateWithHeightAndSubmitRequested(title string, initialText string, submitRequested func(string) Msg, totalHeight int) *modalEditorState {
	state := newLineModalEditorStateWithHeight(title, initialText, nil, totalHeight)
	state.submitRequested = submitRequested
	return state
}

func newMultilineModalEditorState(title string, initialText string, submit func(string) error, totalHeight int) *modalEditorState {
	if submit == nil {
		submit = func(string) error { return nil }
	}

	return &modalEditorState{
		title:       strings.TrimSpace(title),
		editor:      newMultilineEditor(initialText),
		submit:      submit,
		totalHeight: totalHeight,
	}
}

func newMultilineModalEditorStateWithSubmitRequested(title string, initialText string, submitRequested func(string) Msg, totalHeight int) *modalEditorState {
	state := newMultilineModalEditorState(title, initialText, nil, totalHeight)
	state.submitRequested = submitRequested
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

func (state *modalEditorState) Cursor() int {
	if state == nil {
		return 0
	}
	if state.lineEditor != nil {
		return state.lineEditor.Cursor()
	}
	if state.editor != nil {
		return state.editor.Cursor()
	}
	return 0
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

func (state *modalEditorState) submitAction() string {
	if state != nil && state.lineEditor != nil {
		return modalEditorSubmitLineAction
	}
	return modalEditorSubmitAction
}

func (state *modalEditorState) submitHintFallback() string {
	if state != nil && state.lineEditor != nil {
		return "Enter"
	}
	return "Alt+Enter"
}
