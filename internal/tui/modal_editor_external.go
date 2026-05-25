package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) openModalEditorInExternalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorExternalEditRequested{})
}

func (program *Program) setModalEditorTextFromExternalEditor(text string) {
	if program == nil || program.overlayState.modalEditor == nil {
		return
	}
	if program.overlayState.modalEditor.lineEditor != nil {
		program.overlayState.modalEditor.lineEditor.SetText(normalizeSingleLineExternalEditorText(text))
		return
	}
	if program.overlayState.modalEditor.editor != nil {
		program.overlayState.modalEditor.editor.SetText(text)
	}
}

func normalizeSingleLineExternalEditorText(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimRight(normalized, "\n")
	return strings.ReplaceAll(normalized, "\n", " ")
}
