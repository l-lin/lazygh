package tui

import "strings"

func (program *Program) queueModalEditorSubmitRequest(request modalEditorSubmitRequest) []Cmd {
	if program == nil || request == nil {
		return nil
	}

	program.clearModalEditorErrorMessage()
	statusLineOperationID := program.startStatusLineOperation(statusLineOperationForModalEditorRequest(program, request))
	if statusCommand := strings.TrimSpace(request.statusCommand()); statusCommand != "" {
		program.startGHCommandLoading(statusCommand)
	}
	return []Cmd{modalEditorSubmitCmd{request: request, statusLineOperationID: statusLineOperationID}}
}
