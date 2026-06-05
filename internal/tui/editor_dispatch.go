package tui

import "github.com/jesseduffield/gocui"

func (program *Program) dispatchEditorMessage(msg Msg) bool {
	if program == nil || msg == nil {
		return false
	}
	return program.dispatchEditorRuntimeMessage(msg) == nil
}

func (program *Program) dispatchEditorRuntimeMessage(msg Msg) error {
	if program == nil || msg == nil {
		return nil
	}

	switch msg.(type) {
	case MsgModalEditorLineInputRequested, MsgModalEditorMultilineInputRequested:
		if actualErr := program.applyRuntimeMessage(nil, msg, false); actualErr != nil {
			return actualErr
		}
		return program.refreshModalEditorRuntimeViews(program.captureGUI(nil))
	case MsgActionsPopupSearchInputRequested:
		if actualErr := program.applyRuntimeMessage(nil, msg, false); actualErr != nil {
			return actualErr
		}
		return program.refreshActionsPopupRuntimeViews(program.captureGUI(nil))
	default:
		return program.dispatchRuntimeMessage(msg)
	}
}

func (program *Program) refreshModalEditorRuntimeViews(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	gui = program.captureGUI(gui)
	var actualErr error
	program.withRefreshReadCache(func() {
		actualErr = program.refreshOverlayView(gui, program.modalEditorVisible(), viewModalEditorName, program.configureModalEditorView, program.renderModalEditorView)
		if actualErr != nil {
			return
		}
		actualErr = program.refreshCurrentViewFocus(gui)
	})
	return actualErr
}

func (program *Program) refreshActionsPopupRuntimeViews(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	gui = program.captureGUI(gui)
	var actualErr error
	program.withRefreshReadCache(func() {
		actualErr = program.refreshActionsPopupViews(gui)
		if actualErr != nil {
			return
		}
		actualErr = program.refreshCurrentViewFocus(gui)
	})
	return actualErr
}
