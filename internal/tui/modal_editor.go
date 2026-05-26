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

type modalEditorKind int

const (
	modalEditorKindNone modalEditorKind = iota
	modalEditorKindMultiline
	modalEditorKindSingleLine
)

type modalEditorState struct {
	kind             modalEditorKind
	title            string
	editor           multilineEditor
	lineEditor       lineEditor
	errorMessage     string
	submitDescriptor modalEditorSubmitDescriptor
	totalHeight      int
}

func newModalEditorState(title string, initialText string) modalEditorState {
	return newMultilineModalEditorState(title, initialText, modalEditorTotalHeight)
}

func newModalEditorStateWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) modalEditorState {
	state := newModalEditorState(title, initialText)
	state.submitDescriptor = submitDescriptor
	return state
}

func newLineModalEditorState(title string, initialText string) modalEditorState {
	return newLineModalEditorStateWithHeight(title, initialText, lineModalEditorTotalHeight)
}

func newLineModalEditorStateWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) modalEditorState {
	return newLineModalEditorStateWithHeightAndSubmitDescriptor(title, initialText, submitDescriptor, lineModalEditorTotalHeight)
}

func newLineModalEditorStateWithHeight(title string, initialText string, totalHeight int) modalEditorState {
	if totalHeight < 1 {
		totalHeight = lineModalEditorTotalHeight
	}

	return modalEditorState{
		kind:        modalEditorKindSingleLine,
		title:       strings.TrimSpace(title),
		lineEditor:  newLineEditor(initialText),
		totalHeight: totalHeight,
	}
}

func newLineModalEditorStateWithHeightAndSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) modalEditorState {
	state := newLineModalEditorStateWithHeight(title, initialText, totalHeight)
	state.submitDescriptor = submitDescriptor
	return state
}

func newMultilineModalEditorState(title string, initialText string, totalHeight int) modalEditorState {
	return modalEditorState{
		kind:        modalEditorKindMultiline,
		title:       strings.TrimSpace(title),
		editor:      newMultilineEditor(initialText),
		totalHeight: totalHeight,
	}
}

func newMultilineModalEditorStateWithSubmitDescriptor(title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) modalEditorState {
	state := newMultilineModalEditorState(title, initialText, totalHeight)
	state.submitDescriptor = submitDescriptor
	return state
}

func (state modalEditorState) clone() modalEditorState {
	cloned := state
	switch state.kind {
	case modalEditorKindSingleLine:
		cloned.lineEditor = state.lineEditor.clone()
		cloned.editor = multilineEditor{}
	case modalEditorKindMultiline:
		cloned.editor = state.editor.clone()
		cloned.lineEditor = lineEditor{}
	}
	return cloned
}

func (state modalEditorState) visible() bool {
	return state.kind != modalEditorKindNone
}

func (state modalEditorState) isLineEditor() bool {
	return state.kind == modalEditorKindSingleLine
}

func (state *modalEditorState) Text() string {
	if state == nil || !state.visible() {
		return ""
	}
	if state.isLineEditor() {
		return state.lineEditor.Text()
	}
	return state.editor.Text()
}

func (state *modalEditorState) Cursor() int {
	if state == nil || !state.visible() {
		return 0
	}
	if state.isLineEditor() {
		return state.lineEditor.Cursor()
	}
	return state.editor.Cursor()
}

func (state *modalEditorState) CursorXY() (int, int) {
	if state == nil || !state.visible() {
		return 0, 0
	}
	if state.isLineEditor() {
		return state.lineEditor.Cursor(), 0
	}
	return state.editor.CursorXY()
}

func (state *modalEditorState) HandleKey(key gocui.Key, ch rune, mod gocui.Modifier) bool {
	if state == nil || !state.visible() {
		return false
	}
	if state.isLineEditor() {
		return state.lineEditor.HandleKey(key, ch, mod)
	}
	return state.editor.HandleKey(key, ch, mod)
}

func (state *modalEditorState) Height() int {
	if state == nil || state.totalHeight < 1 {
		return modalEditorTotalHeight
	}
	return state.totalHeight
}

func (state *modalEditorState) submitAction() string {
	if state != nil && state.isLineEditor() {
		return modalEditorSubmitLineAction
	}
	return modalEditorSubmitAction
}

func (state *modalEditorState) submitHintFallback() string {
	if state != nil && state.isLineEditor() {
		return "Enter"
	}
	return "Alt+Enter"
}
