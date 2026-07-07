package tui

import "github.com/jesseduffield/gocui"

func (program *Program) modalEditorVisible() bool {
	return program != nil && program.overlayState.modalEditor.visible()
}

func (program *Program) openModalEditor(gui *gocui.Gui, title string, initialText string) error {
	return program.dispatch(gui, MsgModalEditorOpened{Descriptor: newModalEditorOpenDescriptor(title, initialText)})
}

func (program *Program) openModalEditorWithSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) error {
	return program.dispatch(gui, MsgModalEditorOpened{Descriptor: newModalEditorOpenDescriptorWithSubmitDescriptor(title, initialText, submitDescriptor)})
}

func (program *Program) openMultilineModalEditor(gui *gocui.Gui, title string, initialText string, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{Descriptor: newMultilineModalEditorOpenDescriptor(title, initialText, totalHeight)})
}

func (program *Program) openMultilineModalEditorWithSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{Descriptor: newMultilineModalEditorOpenDescriptorWithSubmitDescriptor(title, initialText, submitDescriptor, totalHeight)})
}

func (program *Program) openLineModalEditor(gui *gocui.Gui, title string, initialText string) error {
	return program.openLineModalEditorWithHeight(gui, title, initialText, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor) error {
	return program.openLineModalEditorWithHeightAndSubmitDescriptor(gui, title, initialText, submitDescriptor, lineModalEditorTotalHeight)
}

func (program *Program) openLineModalEditorWithHeight(gui *gocui.Gui, title string, initialText string, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{Descriptor: newLineModalEditorOpenDescriptorWithHeight(title, initialText, totalHeight)})
}

func (program *Program) openLineModalEditorWithHeightAndSubmitDescriptor(gui *gocui.Gui, title string, initialText string, submitDescriptor modalEditorSubmitDescriptor, totalHeight int) error {
	return program.dispatch(gui, MsgModalEditorOpened{Descriptor: newLineModalEditorOpenDescriptorWithHeightAndSubmitDescriptor(title, initialText, submitDescriptor, totalHeight)})
}

func (program *Program) closeModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorClosed{})
}

func (program *Program) submitModalEditor(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgModalEditorSubmitRequested{})
}

func (program *Program) moveModalEditorCursorDown(gui *gocui.Gui, _ *gocui.View) error {
	program.captureGUI(gui)
	return program.dispatchEditorRuntimeMessage(MsgModalEditorMultilineInputRequested{Intent: multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorDown}})
}

func (program *Program) moveModalEditorCursorUp(gui *gocui.Gui, _ *gocui.View) error {
	program.captureGUI(gui)
	return program.dispatchEditorRuntimeMessage(MsgModalEditorMultilineInputRequested{Intent: multilineEditorIntent{kind: multilineEditorIntentKindMoveCursorUp}})
}
