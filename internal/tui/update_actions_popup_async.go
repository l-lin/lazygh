package tui

import "strings"

func (program *Program) queueActionsPopupAsyncRequest(request actionsPopupAsyncRequest) []Cmd {
	if program == nil || request == nil {
		return nil
	}

	program.clearActionsPopupErrorMessage()
	statusLineOperationID := program.startStatusLineOperation(statusLineOperationForActionsPopupRequest(program, request))
	if statusCommand := strings.TrimSpace(request.statusCommand()); statusCommand != "" {
		program.startGHCommandLoading(statusCommand)
	}
	return []Cmd{actionsPopupAsyncCmd{request: request, statusLineOperationID: statusLineOperationID}}
}
