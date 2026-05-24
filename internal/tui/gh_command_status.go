package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

func formatStatusLineCommand(arguments ...string) string {
	normalizedArguments := make([]string, 0, len(arguments))
	for _, argument := range arguments {
		trimmedArgument := strings.TrimSpace(argument)
		if trimmedArgument == "" {
			continue
		}
		normalizedArguments = append(normalizedArguments, trimmedArgument)
	}
	return strings.Join(normalizedArguments, " ")
}

func formatRunningCommandStatus(command string) string {
	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		return ""
	}
	return "Running `" + trimmedCommand + "`."
}

func (program *Program) startGHCommandLoading(command string) {
	if program == nil {
		return
	}

	program.feedbackMessage = ""
	program.ghCommandLoadingMessage = formatRunningCommandStatus(command)
}

func (program *Program) clearGHCommandLoading() {
	if program == nil {
		return
	}

	program.ghCommandLoadingMessage = ""
}

func (program *Program) startActionsPopupAsyncGHCommand(gui *gocui.Gui, command string, work func() error, success actionsPopupAsyncSuccess) actionsPopupActionResult {
	if gui == nil {
		if err := work(); err != nil {
			return actionsPopupActionResult{err: err}
		}
		if success != nil {
			success.apply(program)
		}
		return actionsPopupActionResult{closePopup: true}
	}

	program.actionsPopupWidget.errorMessage = ""
	program.startGHCommandLoading(command)
	program.runAsync(func() {
		err := work()
		program.dispatchAsync(gui, MsgActionsPopupAsyncGHCommandFinished{Err: err, Success: success})
	})
	return actionsPopupActionResult{}
}
