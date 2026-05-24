package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) openModalEditorInExternalEditor(gui *gocui.Gui, view *gocui.View) error {
	if program == nil || program.modalEditor == nil {
		return nil
	}
	if program.externalEditor == nil {
		program.modalEditor.errorMessage = "external editor is unavailable"
		return program.refreshModalEditorAfterExternalEdit(gui, view)
	}

	editedText, err := program.externalEditor.Edit(gui, program.modalEditor.Text())
	if err != nil {
		program.modalEditor.errorMessage = strings.TrimSpace(err.Error())
		return program.refreshModalEditorAfterExternalEdit(gui, view)
	}

	program.modalEditor.errorMessage = ""
	program.setModalEditorTextFromExternalEditor(editedText)
	return program.refreshModalEditorAfterExternalEdit(gui, view)
}

func (program *Program) refreshModalEditorAfterExternalEdit(gui *gocui.Gui, view *gocui.View) error {
	if gui != nil {
		return program.afterStateChange(gui)
	}
	if view != nil {
		program.configureModalEditorView(view)
		program.renderModalEditorView(view)
	}
	return nil
}

func (program *Program) setModalEditorTextFromExternalEditor(text string) {
	if program == nil || program.modalEditor == nil {
		return
	}
	if program.modalEditor.lineEditor != nil {
		program.modalEditor.lineEditor.SetText(normalizeSingleLineExternalEditorText(text))
		return
	}
	if program.modalEditor.editor != nil {
		program.modalEditor.editor.SetText(text)
	}
}

func normalizeSingleLineExternalEditorText(text string) string {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimRight(normalized, "\n")
	return strings.ReplaceAll(normalized, "\n", " ")
}
