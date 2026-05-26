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
	title            string
	editor           *multilineEditor
	lineEditor       *lineEditor
	errorMessage     string
	submitDescriptor modalEditorSubmitDescriptor
	totalHeight      int
}

func newModalEditorState(title string, initialText string) *modalEditorState {
	return newMultilineModalEditorState(title, initialText, modalEditorTotalHeight)
}

func newModalEditorStateWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) *modalEditorState {
	state := newModalEditorState(title, initialText)
	state.submitDescriptor = submitDescriptor
	return state
}

func newLineModalEditorState(title string, initialText string) *modalEditorState {
	return newLineModalEditorStateWithHeight(title, initialText, lineModalEditorTotalHeight)
}

func newLineModalEditorStateWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) *modalEditorState {
	return newLineModalEditorStateWithHeightAndSubmitDescriptor(title, initialText, submitDescriptor, lineModalEditorTotalHeight)
}

func newLineModalEditorStateWithHeight(title string, initialText string, totalHeight int) *modalEditorState {
	if totalHeight < 1 {
		totalHeight = lineModalEditorTotalHeight
	}

	return &modalEditorState{
		title:       strings.TrimSpace(title),
		lineEditor:  newLineEditor(initialText),
		totalHeight: totalHeight,
	}
}

func newLineModalEditorStateWithHeightAndSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) *modalEditorState {
	state := newLineModalEditorStateWithHeight(title, initialText, totalHeight)
	state.submitDescriptor = submitDescriptor
	return state
}

func newMultilineModalEditorState(title string, initialText string, totalHeight int) *modalEditorState {
	return &modalEditorState{
		title:       strings.TrimSpace(title),
		editor:      newMultilineEditor(initialText),
		totalHeight: totalHeight,
	}
}

func newMultilineModalEditorStateWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) *modalEditorState {
	state := newMultilineModalEditorState(title, initialText, totalHeight)
	state.submitDescriptor = submitDescriptor
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
