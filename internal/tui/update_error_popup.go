package tui

import "strings"

func (program *Program) applyErrorReported(message MsgErrorReported) []Cmd {
	if program == nil {
		return nil
	}

	trimmedMessage := strings.TrimSpace(message.Message)
	if trimmedMessage == "" {
		return nil
	}

	program.overlayState.errorMessages = recordedErrorMessagesWithAppended(program.overlayState.errorMessages, trimmedMessage)
	popup := newTransientErrorPopupState(program.overlayState.transientErrorPopup, trimmedMessage, program.currentTime(), program.timingState.transientErrorPopupDuration)
	program.overlayState.transientErrorPopup = popup
	if popup.expiresAt.IsZero() {
		return nil
	}
	return []Cmd{transientErrorPopupExpiryCmd{Generation: popup.generation, Delay: program.timingState.transientErrorPopupDuration}}
}
