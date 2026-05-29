package tui

import "strings"

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
