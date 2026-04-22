package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func handleMultilineModalEditorExternalEditKey(program *Program, view *gocui.View, key gocui.Key, _ rune, _ gocui.Modifier) bool {
	if key != gocui.KeyCtrlG || program == nil || program.modalEditor == nil || program.modalEditor.editor == nil {
		return false
	}
	if program.externalEditor == nil {
		program.modalEditor.errorMessage = "external editor is unavailable"
		program.configureModalEditorView(view)
		program.renderModalEditorView(view)
		return true
	}

	editedText, err := program.externalEditor.Edit(program.gui, program.modalEditor.Text())
	if err != nil {
		program.modalEditor.errorMessage = strings.TrimSpace(err.Error())
		program.configureModalEditorView(view)
		program.renderModalEditorView(view)
		return true
	}

	program.modalEditor.errorMessage = ""
	program.modalEditor.editor.SetText(editedText)
	program.configureModalEditorView(view)
	program.renderModalEditorView(view)
	return true
}
