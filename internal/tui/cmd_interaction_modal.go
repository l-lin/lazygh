package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"
)

type modalEditorCommandRuntime struct {
	externalEditor ExternalEditor
	submitDeps     modalEditorSubmitCommandDeps
	executeMessage func(*gocui.Gui, Msg) error
}

func newModalEditorCommandRuntime(program *Program) modalEditorCommandRuntime {
	if program == nil {
		return modalEditorCommandRuntime{}
	}
	return modalEditorCommandRuntime{
		externalEditor: program.externalEditor,
		submitDeps:     newModalEditorSubmitCommandDeps(program),
		executeMessage: program.executeRuntimeMessage,
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
	if command.request == nil || runtime.executeMessage == nil {
		return
	}

	completion, err := command.request.run(runtime.submitDeps)
	_ = runtime.executeMessage(gui, MsgModalEditorSubmitFinished{Err: err, Completion: completion})
}
