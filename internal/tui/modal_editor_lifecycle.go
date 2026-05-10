package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.modalEditor != nil
}

func (program *Program) openModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	program.modalEditor = newModalEditorState(title, initialText, submit)
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) openMultilineModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error, totalHeight int) error {
	program.modalEditor = newMultilineModalEditorState(title, initialText, submit, totalHeight)
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) openLineModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	program.modalEditor = newLineModalEditorState(title, initialText, submit)
	if gui == nil {
		return nil
	}

	return program.layout(gui)
}

func (program *Program) closeModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	program.modalEditor = nil
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) submitModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	if program.modalEditor == nil {
		return nil
	}

	program.modalEditor.errorMessage = ""
	if err := program.modalEditor.submit(program.modalEditor.Text()); err != nil {
		program.modalEditor.errorMessage = strings.TrimSpace(err.Error())
		if gui == nil {
			return nil
		}
		return program.refreshViews(gui)
	}
	if program.modalEditor.afterSubmit != nil {
		program.modalEditor.afterSubmit(gui)
	}

	return program.closeModalEditor(gui, nil)
}
