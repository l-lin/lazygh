package tui

import (
	"errors"
	"strings"
)

func (program *Program) applyOpenCommandMode() {
	if program == nil || program.model == nil || program.commandModeBlocked() {
		return
	}

	program.clearPendingSelectionPrefix()
	program.clearDetailPendingPrefix()
	program.updateSearchWidget(func(state searchWidgetState) searchWidgetState {
		return state.withCommandModeOpened()
	})
}

func (program *Program) applyCommandInputRequested(message MsgCommandInputRequested) {
	if program == nil || !program.commandModeActive() {
		return
	}
	program.applySearchWidgetEditorIntent(message.Intent)
}

func (program *Program) applySubmitCommand() {
	if program == nil || !program.commandModeActive() {
		return
	}

	command := strings.TrimSpace(program.searchWidget.editor.Text())
	if command == "" {
		program.clearSearchWidgetEditor()
		return
	}

	switch command {
	case commandMessages:
		program.openPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{
			title:         command,
			body:          program.renderRecordedErrorsPopupBody(),
			widthPercent:  pullRequestBuildRunPopupDefaultWidthPercent,
			heightPercent: pullRequestBuildRunPopupDefaultHeightPercent,
		})
	case commandHistory:
		program.openPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{
			title:         command,
			body:          "",
			widthPercent:  pullRequestBuildRunPopupDefaultWidthPercent,
			heightPercent: pullRequestBuildRunPopupDefaultHeightPercent,
		})
	default:
		program.clearSearchWidgetEditor()
		program.setFailureFeedback(program.model.Focus(), "Unknown command", errors.New(command))
	}
}

func (program *Program) applyCancelCommand() {
	if program == nil || !program.commandModeActive() {
		return
	}
	program.clearSearchWidgetEditor()
}
