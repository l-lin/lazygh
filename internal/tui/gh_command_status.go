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

func (program *Program) startActionsPopupAsyncGHCommand(gui *gocui.Gui, command string, work func() error, onSuccess func()) actionsPopupActionResult {
	if gui == nil {
		if err := work(); err != nil {
			return actionsPopupActionResult{err: err}
		}
		if onSuccess != nil {
			onSuccess()
		}
		return actionsPopupActionResult{closePopup: true}
	}

	program.actionsPopupErrorMessage = ""
	program.startGHCommandLoading(command)
	program.asyncRunner.Go(func() {
		err := work()
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			return program.finishActionsPopupAsyncGHCommand(gui, err, onSuccess)
		})
	})
	return actionsPopupActionResult{}
}

func (program *Program) finishActionsPopupAsyncGHCommand(gui *gocui.Gui, err error, onSuccess func()) error {
	program.clearGHCommandLoading()
	if err != nil {
		program.reportError(gui, strings.TrimSpace(err.Error()))
		return program.refreshViewsIfGUI(gui)
	}

	if onSuccess != nil {
		onSuccess()
	}
	if program == nil || program.model == nil || !program.model.ActionsPopupVisible() {
		return program.refreshViewsIfGUI(gui)
	}
	return program.closeActionsPopup(gui, nil)
}
