package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) openModalEditorInExternalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorExternalEditRequested{})
}

func (program *Program) setModalEditorTextFromExternalEditor(text string) {
	if program == nil || !program.modalEditorVisible() {
		return
	}
	program.updateModalEditorState(func(state modalEditorState) modalEditorState {
		return state.withTextFromExternalEditor(text)
	})
}

func normalizeSingleLineExternalEditorText(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimRight(normalized, "\n")
	return strings.ReplaceAll(normalized, "\n", " ")
}
