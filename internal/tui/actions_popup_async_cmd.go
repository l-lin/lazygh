package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"
)

type actionsPopupAsyncCmd struct {
	request actionsPopupAsyncRequest
}

type saveThemePresetCmd struct {
	NormalizedName string
	Label          string
}

func (command actionsPopupAsyncCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.request == nil {
		return
	}

	capturedGUI := program.captureGUI(gui)
	deps := newActionsPopupAsyncCommandDeps(program)
	run := func() {
		completion, err := command.request.run(deps)
		if command.request.asyncRequested() {
			program.dispatchAsyncMessage(MsgActionsPopupAsyncGHCommandFinished{Err: err, Completion: completion})
			return
		}
		_ = program.executeRuntimeMessage(capturedGUI, MsgActionsPopupAsyncGHCommandFinished{Err: err, Completion: completion})
	}
	if command.request.asyncRequested() {
		program.runAsync(run)
		return
	}
	run()
}

func (command saveThemePresetCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil {
		return
	}

	runtime := newThemePresetRuntime(program)
	err := errors.New(themePresetStoreUnavailableMessage)
	if runtime.saveThemePreset != nil {
		err = runtime.saveThemePreset(command.NormalizedName)
	}
	_ = program.executeRuntimeMessage(gui, MsgThemePresetSaved{NormalizedName: command.NormalizedName, Label: command.Label, Err: err})
}
