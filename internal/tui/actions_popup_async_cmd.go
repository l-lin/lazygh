package tui

import (
	"errors"
	"strings"

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

	program.actionsPopupWidget.errorMessage = ""
	if statusCommand := strings.TrimSpace(command.request.statusCommand()); statusCommand != "" {
		program.startGHCommandLoading(statusCommand)
	}
	run := func() {
		success, err := command.request.run(program)
		if command.request.asyncRequested() {
			program.dispatchAsync(gui, MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success})
			return
		}
		program.executeCmds(gui, Update(program, MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success}))
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
