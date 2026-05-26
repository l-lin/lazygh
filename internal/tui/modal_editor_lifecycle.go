package tui

import "github.com/jesseduffield/gocui"

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.overlayState.modalEditor != nil
}

func (program *Program) openModalEditor(gui *gocui.Gui, title string, initialText string) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newModalEditorState(title, initialText)})
}

func (program *Program) openModalEditorWithSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newModalEditorStateWithSubmitDescriptor(title, initialText, submitDescriptor)})
}

func (program *Program) openMultilineModalEditor(gui *gocui.Gui, title string, initialText string, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newMultilineModalEditorState(title, initialText, totalHeight)})
}

func (program *Program) openMultilineModalEditorWithSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newMultilineModalEditorStateWithSubmitDescriptor(title, initialText, submitDescriptor, totalHeight)})
}

func (program *Program) openLineModalEditor(gui *gocui.Gui, title string, initialText string) error {
	return program.openLineModalEditorWithHeight(gui, title, initialText, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) error {
	return program.openLineModalEditorWithHeightAndSubmitDescriptor(gui, title, initialText, submitDescriptor, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithHeight(gui *gocui.Gui, title string, initialText string, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newLineModalEditorStateWithHeight(title, initialText, totalHeight)})
}

func (program *Program) openLineModalEditorWithHeightAndSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{State: newLineModalEditorStateWithHeightAndSubmitDescriptor(title, initialText, submitDescriptor, totalHeight)})
}

func (program *Program) closeModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorClosed{})
}

func (program *Program) submitModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorSubmitRequested{})
}
