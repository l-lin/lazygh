package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"
)

type modalEditorCommandRuntime struct {
	externalEditor       ExternalEditor
	submitDeps           modalEditorSubmitCommandDeps
	executeMessage       func(*gocui.Gui, Msg) error
	dispatchAsyncMessage func(Msg)
	runAsync             func(func())
}

func newModalEditorCommandRuntime(program *Program) modalEditorCommandRuntime {
	if program == nil {
		return modalEditorCommandRuntime{}
	}
	dispatchAsync := asyncRuntimeMessageDispatcher(program)
	return modalEditorCommandRuntime{
		externalEditor:       program.externalEditor,
		submitDeps:           newModalEditorSubmitCommandDeps(program),
		executeMessage:       program.executeRuntimeMessage,
		dispatchAsyncMessage: dispatchAsync,
		runAsync:             program.runAsync,
	}
}

type modalEditorExternalEditCmd struct {
	Text string
}

func (command modalEditorExternalEditCmd) execute(program *Program, gui *gocui.Gui) {
	executeModalEditorExternalEditCommand(newModalEditorCommandRuntime(program), gui, command)
}

func executeModalEditorExternalEditCommand(runtime modalEditorCommandRuntime, gui *gocui.Gui, command modalEditorExternalEditCmd) {
	if runtime.executeMessage == nil {
		return
	}
	if runtime.externalEditor == nil {
		_ = runtime.executeMessage(gui, MsgModalEditorExternalEditFinished{Err: errors.New("external editor is unavailable")})
		return
	}

	editedText, err := runtime.externalEditor.Edit(gui, command.Text)
	_ = runtime.executeMessage(gui, MsgModalEditorExternalEditFinished{Text: editedText, Err: err})
}

type modalEditorSubmitCmd struct {
	request modalEditorSubmitRequest
}

func (command modalEditorSubmitCmd) execute(program *Program, gui *gocui.Gui) {
	executeModalEditorSubmitCommand(newModalEditorCommandRuntime(program), gui, command)
}

func executeModalEditorSubmitCommand(runtime modalEditorCommandRuntime, gui *gocui.Gui, command modalEditorSubmitCmd) {
	if command.request == nil {
		return
	}

	run := func() {
		completion, err := command.request.run(runtime.submitDeps)
		message := MsgModalEditorSubmitFinished{Err: err, Completion: completion}
		if command.request.asyncRequested() && runtime.dispatchAsyncMessage != nil {
			runtime.dispatchAsyncMessage(message)
			return
		}
		if runtime.executeMessage != nil {
			_ = runtime.executeMessage(gui, message)
		}
	}
	if command.request.asyncRequested() && runtime.runAsync != nil {
		runtime.runAsync(run)
		return
	}
	if runtime.executeMessage == nil && runtime.dispatchAsyncMessage == nil {
		return
	}
	run()
}
