package tui

import "github.com/jesseduffield/gocui"

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.overlayState.modalEditor != nil
}

func (program *Program) openModalEditorFromActionsPopup(gui *gocui.Gui, open func(*gocui.Gui) error) error {
	wasVisible := program.modalEditorVisible()
	if err := open(gui); err != nil {
		return err
	}
	if !wasVisible && program.modalEditorVisible() {
		return program.closeActionsPopupIfVisible(gui)
	}
	return errActionsPopupActionUnavailable
}

func (program *Program) openModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newModalEditorState(title, initialText, submit)})
}

func (program *Program) openModalEditorWithSubmitRequested(gui *gocui.Gui, title string, initialText string, submitRequested func(string) Msg) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newModalEditorStateWithSubmitRequested(title, initialText, submitRequested)})
}

func (program *Program) openMultilineModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newMultilineModalEditorState(title, initialText, submit, totalHeight)})
}

func (program *Program) openMultilineModalEditorWithSubmitRequested(gui *gocui.Gui, title string, initialText string, submitRequested func(string) Msg, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newMultilineModalEditorStateWithSubmitRequested(title, initialText, submitRequested, totalHeight)})
}

func (program *Program) openLineModalEditor(gui *gocui.Gui, title string, initialText string, submit func(string) error) error {
	return program.openLineModalEditorWithHeight(gui, title, initialText, submit, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithSubmitRequested(gui *gocui.Gui, title string, initialText string, submitRequested func(string) Msg) error {
	return program.openLineModalEditorWithHeightAndSubmitRequested(gui, title, initialText, submitRequested, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithHeight(gui *gocui.Gui, title string, initialText string, submit func(string) error, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newLineModalEditorStateWithHeight(title, initialText, submit, totalHeight)})
}

func (program *Program) openLineModalEditorWithHeightAndSubmitRequested(gui *gocui.Gui, title string, initialText string, submitRequested func(string) Msg, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newLineModalEditorStateWithHeightAndSubmitRequested(title, initialText, submitRequested, totalHeight)})
}

func (program *Program) closeModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorClosed{})
}

func (program *Program) submitModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorSubmitRequested{})
}
