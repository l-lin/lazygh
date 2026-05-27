package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyActionsPopupActionErrorHandled(message MsgActionsPopupActionErrorHandled) []Cmd {
	if message.Err == nil {
		return nil
	}
	if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
		program.actionsPopupWidget.errorMessage = ""
		return []Cmd{reportErrorCmd{Message: popupMessage}}
	}
	var feedbackErr actionsPopupStatusLineError
	if errors.As(message.Err, &feedbackErr) {
		program.actionsPopupWidget.errorMessage = ""
		program.setFeedback(feedbackErr.feedbackTarget, message.Err.Error())
		return nil
	}
	program.actionsPopupWidget.errorMessage = strings.TrimSpace(message.Err.Error())
	return []Cmd{reportErrorCmd{Message: program.actionsPopupWidget.errorMessage}}
}

func (program *Program) applyActionsPopupClosedWithFeedback(message MsgActionsPopupClosedWithFeedback) {
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	program.setFeedback(message.Target, strings.TrimSpace(message.Message))
}

func (program *Program) applyActionsPopupSearchInputRequested(message MsgActionsPopupSearchInputRequested) []Cmd {
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	if !program.actionsPopupWidget.hasSearchEditor() {
		program.actionsPopupWidget.openSearchEditor(program.model.ActionsPopupSearchQuery())
	}
	if !program.actionsPopupWidget.searchEditor.ApplyIntent(message.Intent) {
		return nil
	}

	query := program.actionsPopupWidget.searchEditor.Text()
	program.clearActionsPopupPendingConfirmation()
	requestID := 0
	if program.assigneePickerVisible() {
		requestID = program.resetAssigneePickerSearch(query)
	}
	program.updateActionsPopupSearch(query)
	program.actionsPopupWidget.errorMessage = ""
	if program.assigneePickerVisible() && requestID > 0 && strings.TrimSpace(query) != "" {
		return []Cmd{assigneePickerSearchCmd{RequestID: requestID, Query: query, Delay: program.actionsPopupWidget.assigneePickerSearchDebounceDelay, DispatchLoading: true}}
	}
	return nil
}

func (program *Program) applyModalEditorSubmitRequested() []Cmd {
	if program == nil || !program.modalEditorVisible() {
		return nil
	}

	editorState := &program.overlayState.modalEditor
	editorState.errorMessage = ""
	requested := modalEditorSubmitRequestedMessage(editorState.submitDescriptor, editorState.Text())
	if requested == nil {
		return nil
	}
	return Update(program, requested)
}

func (program *Program) applyModalEditorSubmitFinished(message MsgModalEditorSubmitFinished) []Cmd {
	if program == nil || !program.modalEditorVisible() {
		return nil
	}

	if message.Err != nil {
		if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
			program.overlayState.modalEditor.errorMessage = ""
			return []Cmd{reportErrorCmd{Message: popupMessage}}
		}
		var feedbackErr modalEditorStatusLineError
		if errors.As(message.Err, &feedbackErr) {
			program.setFeedback(feedbackErr.feedbackTarget, message.Err.Error())
			return nil
		}
		program.overlayState.modalEditor.errorMessage = strings.TrimSpace(message.Err.Error())
		return nil
	}

	var commands []Cmd
	if message.Success != nil {
		commands = message.Success.apply(program)
	}
	program.overlayState.modalEditor = modalEditorState{}
	return commands
}

func (program *Program) applyModalEditorExternalEditRequested() []Cmd {
	if program == nil || !program.modalEditorVisible() {
		return nil
	}
	return []Cmd{modalEditorExternalEditCmd{Text: program.overlayState.modalEditor.Text()}}
}

func (program *Program) applyModalEditorExternalEditFinished(message MsgModalEditorExternalEditFinished) {
	if program == nil || !program.modalEditorVisible() {
		return
	}
	if message.Err != nil {
		program.overlayState.modalEditor.errorMessage = strings.TrimSpace(message.Err.Error())
		return
	}

	program.overlayState.modalEditor.errorMessage = ""
	program.setModalEditorTextFromExternalEditor(message.Text)
}

func (program *Program) applyModalEditorLineInputRequested(message MsgModalEditorLineInputRequested) {
	if program == nil || !program.overlayState.modalEditor.applyLineEditorIntent(message.Intent) {
		return
	}
	program.overlayState.modalEditor.errorMessage = ""
}

func (program *Program) applyModalEditorMultilineInputRequested(message MsgModalEditorMultilineInputRequested) {
	if program == nil || !program.overlayState.modalEditor.applyMultilineEditorIntent(message.Intent) {
		return
	}
	program.overlayState.modalEditor.errorMessage = ""
}

func (program *Program) applyModalEditorOpened(message MsgModalEditorOpened) {
	program.openModalEditorState(message.Descriptor.state())
	if program != nil && program.model != nil && program.model.ActionsPopupVisible() {
		program.closeActionsPopupForAcceptedRequest()
	}
}

func (program *Program) openModalEditorState(state modalEditorState) {
	if program == nil {
		return
	}
	program.overlayState.modalEditor = state.clone()
}

func (program *Program) applyPullRequestBuildRunLoadRequested(message MsgPullRequestBuildRunLoadRequested) []Cmd {
	if program == nil || !program.hasBuildQueries() || program.pullRequestBuildRunLoad != nil {
		return nil
	}

	target := message.Target
	repository := strings.TrimSpace(target.popupContent.repository)
	if repository == "" {
		repository = pullRequestRepositoryName(target.summary.Repository)
	}
	if repository == "" || repository == "-" {
		return nil
	}

	program.feedbackMessage = ""
	program.pullRequestBuildRunPopup = nil
	program.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: formatPullRequestBuildRunCommand(repository, target.check)}
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{pullRequestBuildRunLoadCmd{Repository: repository, Target: target}}
}

func (program *Program) applyPullRequestBuildRunJobLogLoadRequested(message MsgPullRequestBuildRunJobLogLoadRequested) []Cmd {
	if program == nil || !program.hasBuildQueries() || program.pullRequestBuildRunLoad != nil {
		return nil
	}

	repository := pullRequestRepositoryName(message.Summary.Repository)
	if repository == "" || repository == "-" {
		return nil
	}

	program.feedbackMessage = ""
	program.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: formatPullRequestBuildRunJobsCommand(repository, message.Check)}
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{pullRequestBuildRunJobLogLoadCmd{Repository: repository, Check: message.Check}}
}

func (program *Program) applyPullRequestBuildRunPopupOpened(message MsgPullRequestBuildRunPopupOpened) {
	program.openPullRequestBuildRunPopupState(message.Content)
	program.closeActionsPopupForAcceptedRequest()
}

func (program *Program) openPullRequestBuildRunPopupState(content pullRequestBuildRunPopupContent) {
	if program == nil {
		return
	}
	program.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(content)
}

func newPullRequestBuildRunPopupState(content pullRequestBuildRunPopupContent) *pullRequestBuildRunPopupState {
	title := strings.TrimSpace(content.title)
	if title == "" {
		title = pullRequestBuildRunPopupTitle(content.checkTitle)
	}

	copiedJobs := append([]githubdomain.PullRequestBuildRunJob(nil), content.jobs...)
	return &pullRequestBuildRunPopupState{
		title:         title,
		runURL:        strings.TrimSpace(content.runURL),
		repository:    strings.TrimSpace(content.repository),
		body:          renderPullRequestBuildRunPopupContent(content),
		jobs:          copiedJobs,
		previousPopup: content.previousPopup,
		widthPercent:  content.widthPercent,
		heightPercent: content.heightPercent,
		viewState:     newDetailViewState(),
		documents:     map[int]detailDocument{},
	}
}

func (program *Program) applyPullRequestBuildRunPopupClosed() {
	if popup := program.pullRequestBuildRunPopup; popup != nil && popup.viewState.mode.isVisual() {
		popup.viewState.exitVisualMode()
		return
	}
	if popup := program.pullRequestBuildRunPopup; popup != nil && popup.viewState.hasPendingYank() {
		popup.viewState.clearPendingPrefix()
		return
	}
	if popup := program.pullRequestBuildRunPopup; popup != nil && popup.previousPopup != nil {
		program.pullRequestBuildRunPopup = popup.previousPopup
		return
	}
	program.pullRequestBuildRunPopup = nil
}

func (program *Program) applyOpenBrowserURLRequested(message MsgOpenBrowserURLRequested) []Cmd {
	if strings.TrimSpace(message.URL) == "" {
		program.setFeedback(message.Target, message.FailureMessage)
		return nil
	}
	return []Cmd{openBrowserURLCmd{URL: message.URL, SuccessMessage: message.SuccessMessage, FailureMessage: message.FailureMessage, Target: message.Target}}
}

func (program *Program) applyOpenBrowserURLFinished(message MsgOpenBrowserURLFinished) {
	if message.Err == nil {
		program.setFeedback(message.Target, message.SuccessMessage)
		return
	}
	program.setFeedback(message.Target, message.FailureMessage)
}

func (program *Program) applyClipboardWriteFinished(message MsgClipboardWriteFinished) {
	if message.Err == nil {
		switch message.SelectionTarget {
		case clipboardWriteSelectionDetail:
			program.activateYankHighlight(&program.detailState.viewState, message.Selection)
		case clipboardWriteSelectionBuildPopup:
			if program.pullRequestBuildRunPopup != nil {
				program.activateYankHighlight(&program.pullRequestBuildRunPopup.viewState, message.Selection)
			}
		}
		program.setFeedback(message.Target, message.SuccessMessage)
		return
	}
	program.setFeedback(message.Target, message.FailureMessage)
}

func (program *Program) applyOpenPullRequestByURLSubmitRequested(message MsgOpenPullRequestByURLSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: openPullRequestByURLSubmitRequest{rawURL: message.URL}}}
}

func (program *Program) applyPullRequestURLReadFromClipboard(message MsgPullRequestURLReadFromClipboard) {
	if message.Err != nil {
		program.setFeedback(program.model.Focus(), openPullRequestByClipboardFeedbackMessage(message.Err))
		return
	}
	if err := program.OpenPullRequestByURL(message.URL); err != nil {
		if errors.Is(err, githubdomain.ErrInvalidPullRequestURL) {
			program.setFeedback(program.model.Focus(), openPullRequestByClipboardInvalidMessage)
			return
		}
		program.setFeedback(program.model.Focus(), strings.TrimSpace(err.Error()))
	}
}

func (program *Program) applyOpenLinkUnderCursorRequested(message MsgOpenLinkUnderCursorRequested) []Cmd {
	program.detailState.viewState.clearPendingPrefix()
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{openLinkUnderCursorCmd{Target: program.model.Focus()}}
}

func (program *Program) applyOpenPullRequestBuildRunPopupLinkRequested(message MsgOpenPullRequestBuildRunPopupLinkRequested) []Cmd {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}
	popup.viewState.clearPendingPrefix()
	return []Cmd{openPullRequestBuildRunPopupLinkCmd{Target: program.model.Focus()}}
}

func (program *Program) applyCopyPullRequestURLRequested(message MsgCopyPullRequestURLRequested) []Cmd {
	if program.model.Focus() == FocusDetailView && program.detailState.viewState.mode.isVisual() {
		return []Cmd{prepareSelectedDetailClipboardWriteCmd{Target: program.model.Focus()}}
	}

	program.detailState.viewState.clearPendingPrefix()
	url, ok := program.selectedPullRequestURL()
	if !ok {
		if program.model != nil && program.model.ActionsPopupVisible() {
			program.actionsPopupWidget.errorMessage = yankUnavailableMessage
			return nil
		}
		program.setFeedback(program.model.Focus(), yankUnavailableMessage)
		return nil
	}
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{writeClipboardCmd{Text: url, SuccessMessage: yankSuccessMessage, FailureMessage: yankFailureMessage, Target: program.model.Focus()}}
}

func (program *Program) applyCopyPullRequestBuildRunPopupContentRequested(message MsgCopyPullRequestBuildRunPopupContentRequested) []Cmd {
	if program.pullRequestBuildRunPopup == nil {
		return nil
	}
	return []Cmd{preparePullRequestBuildRunPopupClipboardWriteCmd{Target: program.model.Focus()}}
}

func (program *Program) applyOpenNotificationInBrowserRequested() []Cmd {
	if program.linkOpener == nil {
		if program.model != nil && program.model.ActionsPopupVisible() {
			program.actionsPopupWidget.errorMessage = openLinkOpenerUnavailableMessage
			return nil
		}
		return []Cmd{reportErrorCmd{Message: openLinkOpenerUnavailableMessage}}
	}
	browserURL, ok := program.selectedNotificationBrowserURL()
	if !ok {
		if program.model != nil && program.model.ActionsPopupVisible() {
			program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
			return nil
		}
		return []Cmd{reportErrorCmd{Message: errActionsPopupActionUnavailable.Error()}}
	}
	program.closeActionsPopupForAcceptedRequest()
	return program.applyOpenBrowserURLRequested(MsgOpenBrowserURLRequested{URL: browserURL, SuccessMessage: notificationOpenBrowserSuccessMessage, FailureMessage: openLinkFailureMessage, Target: program.model.Focus()})
}

func (program *Program) applyRefreshNotificationsRequested() []Cmd {
	program.beginManualNotificationsRefresh()
	return []Cmd{refreshNotificationsCmd{}}
}
