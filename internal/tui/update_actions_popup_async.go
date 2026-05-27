package tui

import "strings"

func (program *Program) queueActionsPopupAsyncRequest(request actionsPopupAsyncRequest) []Cmd {
	if program == nil || request == nil {
		return nil
	}

	program.actionsPopupWidget.errorMessage = ""
	if statusCommand := strings.TrimSpace(request.statusCommand()); statusCommand != "" {
		program.startGHCommandLoading(statusCommand)
	}
	return []Cmd{actionsPopupAsyncCmd{request: request}}
}
