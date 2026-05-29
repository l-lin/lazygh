package tui

import "time"

func (state overlayStateModel) withHelpVisible(visible bool) overlayStateModel {
	state.helpVisible = visible
	return state
}

func (state overlayStateModel) withRecordedError(message string) overlayStateModel {
	state.errorMessages = recordedErrorMessagesWithAppended(state.errorMessages, message)
	return state
}

func (state overlayStateModel) withTransientErrorPopup(popup transientErrorPopupState) overlayStateModel {
	state.transientErrorPopup = popup
	return state
}

func (state overlayStateModel) withClearedTransientErrorPopup() overlayStateModel {
	state.transientErrorPopup = transientErrorPopupState{}
	return state
}

func (state overlayStateModel) withModalEditor(modalEditor modalEditorState) overlayStateModel {
	state.modalEditor = modalEditor
	return state
}

func (state overlayStateModel) withClearedModalEditor() overlayStateModel {
	state.modalEditor = modalEditorState{}
	return state
}

func (state overlayStateModel) withReportedError(message string, now time.Time, duration time.Duration) (overlayStateModel, transientErrorPopupState) {
	state = state.withRecordedError(message)
	popup := newTransientErrorPopupState(state.transientErrorPopup, message, now, duration)
	state = state.withTransientErrorPopup(popup)
	return state, popup
}
