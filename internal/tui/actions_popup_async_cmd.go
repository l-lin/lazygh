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
		success, err := command.request.run(deps)
		if command.request.asyncRequested() {
			program.dispatchAsyncMessage(MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success})
			return
		}
		program.executeCmds(capturedGUI, Update(program, MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success}))
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

	err := error(nil)
	if program.themePresetStore != nil {
		err = program.themePresetStore.SaveThemePreset(command.NormalizedName)
	}
	if program.themePresetStore == nil {
		err = errors.New("theme preset store is unavailable")
	}
	program.executeCmds(gui, Update(program, MsgThemePresetSaved{NormalizedName: command.NormalizedName, Label: command.Label, Err: err}))
}
