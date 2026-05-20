package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.modalEditor != nil
}

func (program *Program) openModalEditorFromActionsPopup(gui *gocui.Gui, open func(*gocui.Gui) error) actionsPopupActionResult {
	wasVisible := program.modalEditorVisible()
	if err := open(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	if !wasVisible && program.modalEditorVisible() {
		return actionsPopupActionResult{closePopup: true}
	}
	return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
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
	return program.openLineModalEditorWithHeight(gui, title, initialText, submit, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithHeight(gui *gocui.Gui, title string, initialText string, submit func(string) error, totalHeight int) error {
	program.modalEditor = newLineModalEditorStateWithHeight(title, initialText, submit, totalHeight)
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
		if message, ok := transientErrorPopupActionMessage(err); ok {
			program.modalEditor.errorMessage = ""
			program.reportError(gui, message)
			if gui == nil {
				return nil
			}
			return program.refreshViews(gui)
		}
		var feedbackErr modalEditorStatusLineError
		if errors.As(err, &feedbackErr) {
			program.setFeedback(feedbackErr.feedbackTarget, err.Error())
			if gui == nil {
				return nil
			}
			return program.refreshViews(gui)
		}
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
