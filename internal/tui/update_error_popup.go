package tui

import "strings"

func (program *Program) recordError(operationErr error) {
	if operationErr == nil {
		return
	}
	program.recordErrorMessage(operationErr.Error())
}

func (program *Program) recordStatusLineFailure(action string, operationErr error) {
	if message := formatStatusLineFailureContext(action, operationErr); message != "" {
		program.recordErrorMessage(message)
		return
	}
	program.recordError(operationErr)
}

func (program *Program) recordErrorMessage(message string) {
	if program == nil {
		return
	}
	now := program.currentTime()
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		return state.withRecordedError(message, now)
	})
}

func (program *Program) applyErrorReportedMessage(message string) []Cmd {
	return program.applyErrorReported(MsgErrorReported{Message: strings.TrimSpace(message)})
}

func (program *Program) applyErrorReported(message MsgErrorReported) []Cmd {
	if program == nil {
		return nil
	}

	trimmedMessage := strings.TrimSpace(message.Message)
	if trimmedMessage == "" {
		return nil
	}

	popup := transientErrorPopupState{}
	program.updateOverlayState(func(state overlayStateModel) overlayStateModel {
		updatedState, updatedPopup := state.withReportedError(trimmedMessage, program.currentTime(), program.timingState.transientErrorPopupDuration)
		popup = updatedPopup
		return updatedState
	})
	if popup.expiresAt.IsZero() {
		return nil
	}
	return []Cmd{transientErrorPopupExpiryCmd{Generation: popup.generation, Delay: program.timingState.transientErrorPopupDuration}}
}
