package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

type actionsPopupAsyncWorkCmd struct {
	Command string
	Async   bool
	Work    func(*Program) (actionsPopupAsyncSuccess, error)
}

type saveThemePresetCmd struct {
	NormalizedName string
	Label          string
}

func (command actionsPopupAsyncWorkCmd) execute(program *Program, gui *gocui.Gui) {
	if program == nil || command.Work == nil {
		return
	}

	program.actionsPopupWidget.errorMessage = ""
	if strings.TrimSpace(command.Command) != "" {
		program.startGHCommandLoading(command.Command)
	}
	run := func() {
		success, err := command.Work(program)
		if command.Async {
			program.dispatchAsync(gui, MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success})
			return
		}
		program.executeCmds(gui, Update(program, MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success}))
	}
	if command.Async {
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
