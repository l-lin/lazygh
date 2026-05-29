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
		program.clearActionsPopupErrorMessage()
		return program.applyErrorReportedMessage(popupMessage)
	}
	var feedbackErr actionsPopupStatusLineError
	if errors.As(message.Err, &feedbackErr) {
		program.clearActionsPopupErrorMessage()
		program.setFeedback(feedbackErr.feedbackTarget, message.Err.Error())
		return nil
	}
	program.setActionsPopupErrorMessage(message.Err.Error())
	return program.applyErrorReportedMessage(program.actionsPopupWidget.errorMessage)
}

func (program *Program) applyModalEditorSubmitRequestMessage(requested Msg) []Cmd {
	switch actual := requested.(type) {
	case nil:
		return nil
	case MsgOpenPullRequestByURLSubmitRequested:
		return program.applyOpenPullRequestByURLSubmitRequested(actual)
	case MsgPullRequestCustomSearchSubmitRequested:
		return program.applyPullRequestCustomSearchSubmitRequested(actual)
	case MsgPullRequestCommentSubmitRequested:
		return program.applyPullRequestCommentSubmitRequested(actual)
	case MsgPullRequestReviewCommentSubmitRequested:
		return program.applyPullRequestReviewCommentSubmitRequested(actual)
	case MsgPullRequestRequestChangesSubmitRequested:
		return program.applyPullRequestRequestChangesSubmitRequested(actual)
	case MsgPullRequestTitleEditRequested:
		return program.applyPullRequestTitleEditRequested(actual)
	case MsgPullRequestDescriptionEditRequested:
		return program.applyPullRequestDescriptionEditRequested(actual)
	case MsgPullRequestCommentUpdateRequested:
		return program.applyPullRequestCommentUpdateRequested(actual)
	case MsgInlineCommentUpdateRequested:
		return program.applyInlineCommentUpdateRequested(actual)
	case MsgInlineCommentReplySubmitRequested:
		return program.applyInlineCommentReplySubmitRequested(actual)
	case MsgReviewInlineCommentSubmitRequested:
		return program.applyReviewInlineCommentSubmitRequested(actual)
	case MsgPendingPullRequestReviewSubmitRequested:
		return program.applyPendingPullRequestReviewSubmitRequested(actual)
	default:
		return nil
	}
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
		program.openActionsPopupSearchEditor(program.model.ActionsPopupSearchQuery())
	}
	if !program.applyActionsPopupSearchEditorIntent(message.Intent) {
		return nil
	}

	query := program.actionsPopupWidget.searchEditor.Text()
	program.clearActionsPopupPendingConfirmation()
	requestID := 0
	if program.assigneePickerVisible() {
		requestID = program.resetAssigneePickerSearch(query)
	}
	program.updateActionsPopupSearch(query)
	program.clearActionsPopupErrorMessage()
	if program.assigneePickerVisible() && requestID > 0 && strings.TrimSpace(query) != "" {
		return []Cmd{assigneePickerSearchCmd{RequestID: requestID, Query: query, Delay: program.actionsPopupWidget.assigneePickerSearchDebounceDelay, DispatchLoading: true}}
	}
	return nil
}

func (program *Program) applyModalEditorSubmitRequested() []Cmd {
	if program == nil || !program.modalEditorVisible() {
		return nil
	}

	program.clearModalEditorErrorMessage()
	editorState := program.overlayState.modalEditor
	requested := modalEditorSubmitRequestedMessage(editorState.submitDescriptor, editorState.Text())
	if requested == nil {
		return nil
	}
	return program.applyModalEditorSubmitRequestMessage(requested)
}

func (program *Program) applyModalEditorSubmitFinished(message MsgModalEditorSubmitFinished) []Cmd {
	if program == nil || !program.modalEditorVisible() {
		return nil
	}

	if message.Err != nil {
		if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
			program.clearModalEditorErrorMessage()
			return program.applyErrorReportedMessage(popupMessage)
		}
		var feedbackErr modalEditorStatusLineError
		if errors.As(message.Err, &feedbackErr) {
			program.setFeedback(feedbackErr.feedbackTarget, message.Err.Error())
			return nil
		}
		program.setModalEditorErrorMessage(message.Err.Error())
		return nil
	}

	commands := program.applyModalEditorSubmitCompletion(message.Completion)
	if !modalEditorRemainsOpenAfterCompletion(message.Completion) {
		program.overlayState.modalEditor = modalEditorState{}
	}
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
		program.setModalEditorErrorMessage(message.Err.Error())
		return
	}

	program.setModalEditorTextFromExternalEditor(message.Text)
}

func (program *Program) applyModalEditorLineInputRequested(message MsgModalEditorLineInputRequested) {
	if program == nil {
		return
	}
	updatedState, ok := program.overlayState.modalEditor.withLineEditorIntentApplied(message.Intent)
	if !ok {
		return
	}
	program.overlayState.modalEditor = updatedState
}

func (program *Program) applyModalEditorMultilineInputRequested(message MsgModalEditorMultilineInputRequested) {
	if program == nil {
		return
	}
	updatedState, ok := program.overlayState.modalEditor.withMultilineEditorIntentApplied(message.Intent)
	if !ok {
		return
	}
	program.overlayState.modalEditor = updatedState
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

	program.clearFeedbackMessage()
	program.startPullRequestBuildRunLoad(formatPullRequestBuildRunCommand(repository, target.check))
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

	program.clearFeedbackMessage()
	program.startPullRequestBuildRunJobLogLoad(formatPullRequestBuildRunJobsCommand(repository, message.Check))
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
	program.updateBuildStore(func(store buildStore) buildStore {
		return store.withPopupOpened(content)
	})
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
		program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
			return state.withVisualModeExited()
		})
		return
	}
	if popup := program.pullRequestBuildRunPopup; popup != nil && popup.viewState.hasPendingYank() {
		program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
			return state.withPendingPrefixCleared()
		})
		return
	}
	program.closePullRequestBuildRunPopupState()
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

func (program *Program) applySelectedDetailClipboardPrepared(message MsgSelectedDetailClipboardPrepared) []Cmd {
	searchQuery := program.model.DetailSearchQuery()
	detailState := program.detailState.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, searchQuery)
	selection, _ := detailSelectionForCurrentMode(detailState.viewState, message.Document)
	text := detailState.viewState.selectedText(message.Document)
	program.detailState = detailState.withVisualModeExited()
	return []Cmd{writeClipboardCmd{
		Text:            text,
		SuccessMessage:  detailYankSuccessMessage,
		FailureMessage:  detailYankFailureMessage,
		Target:          message.Target,
		Selection:       selection,
		SelectionTarget: clipboardWriteSelectionDetail,
	}}
}

func (program *Program) applyPullRequestBuildRunPopupClipboardPrepared(message MsgPullRequestBuildRunPopupClipboardPrepared) []Cmd {
	if program.pullRequestBuildRunPopup == nil {
		return nil
	}

	prepared := pullRequestBuildRunPopupClipboardResult{}
	program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		prepared = state.preparedClipboard(message.Document, message.ViewportHeight)
		return prepared.state
	})
	if prepared.hasVisualSelection {
		return []Cmd{writeClipboardCmd{
			Text:            prepared.text,
			SuccessMessage:  detailYankSuccessMessage,
			FailureMessage:  detailYankFailureMessage,
			Target:          message.Target,
			Selection:       prepared.selection,
			SelectionTarget: clipboardWriteSelectionBuildPopup,
		}}
	}

	runURL := strings.TrimSpace(prepared.state.runURL)
	if runURL == "" {
		program.setFeedback(message.Target, yankUnavailableMessage)
		return nil
	}
	return []Cmd{writeClipboardCmd{Text: runURL, SuccessMessage: yankSuccessMessage, FailureMessage: yankFailureMessage, Target: message.Target}}
}

func (program *Program) applyOpenPullRequestByURLSubmitRequested(message MsgOpenPullRequestByURLSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: openPullRequestByURLSubmitRequest{rawURL: message.URL}}}
}

func (program *Program) applyPullRequestURLReadFromClipboard(message MsgPullRequestURLReadFromClipboard) {
	if message.Err != nil {
		program.setFeedback(program.model.Focus(), openPullRequestByClipboardFeedbackMessage(message.Err))
		return
	}
	if err := program.openPullRequestInPastedTabByURL(message.URL); err != nil {
		if errors.Is(err, githubdomain.ErrInvalidPullRequestURL) {
			program.setFeedback(program.model.Focus(), openPullRequestByClipboardInvalidMessage)
			return
		}
		program.setFeedback(program.model.Focus(), strings.TrimSpace(err.Error()))
	}
}

func (program *Program) applyOpenLinkUnderCursorRequested(message MsgOpenLinkUnderCursorRequested) []Cmd {
	program.clearDetailPendingPrefix()
	program.closeActionsPopupForAcceptedRequest()
	return []Cmd{openLinkUnderCursorCmd{Target: program.model.Focus()}}
}

func (program *Program) applyOpenPullRequestBuildRunPopupLinkRequested(message MsgOpenPullRequestBuildRunPopupLinkRequested) []Cmd {
	if program.pullRequestBuildRunPopup == nil {
		return nil
	}
	program.updatePullRequestBuildRunPopup(func(state pullRequestBuildRunPopupState) pullRequestBuildRunPopupState {
		return state.withPendingPrefixCleared()
	})
	return []Cmd{openPullRequestBuildRunPopupLinkCmd{Target: program.model.Focus()}}
}

func (program *Program) applyOpenLinkUnderCursorResolved(message MsgOpenLinkUnderCursorResolved) []Cmd {
	return program.applyResolvedLinkOpen(message.Target, message.URL, message.LinkAvailable, message.OpenerAvailable)
}

func (program *Program) applyOpenPullRequestBuildRunPopupLinkResolved(message MsgOpenPullRequestBuildRunPopupLinkResolved) []Cmd {
	return program.applyResolvedLinkOpen(message.Target, message.URL, message.LinkAvailable, message.OpenerAvailable)
}

func (program *Program) applyResolvedLinkOpen(target Focus, url string, linkAvailable bool, openerAvailable bool) []Cmd {
	switch {
	case !linkAvailable:
		program.setFeedback(target, openLinkUnavailableMessage)
		return nil
	case !openerAvailable:
		program.setFeedback(target, openLinkOpenerUnavailableMessage)
		return nil
	default:
		return program.applyOpenBrowserURLRequested(MsgOpenBrowserURLRequested{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: target})
	}
}

func (program *Program) applyCopyPullRequestURLRequested(message MsgCopyPullRequestURLRequested) []Cmd {
	if program == nil {
		return nil
	}

	program.clearPendingSelectionPrefix()
	if program.overlayState.helpVisible || program.model.SearchActive() {
		return nil
	}
	if program.model.Focus() == FocusDetailView && program.detailState.viewState.mode.isVisual() {
		return []Cmd{prepareSelectedDetailClipboardWriteCmd{Target: program.model.Focus()}}
	}

	program.clearDetailPendingPrefix()
	url, ok := program.selectedPullRequestURL()
	if !ok {
		if program.model != nil && program.model.ActionsPopupVisible() {
			program.setActionsPopupErrorMessage(yankUnavailableMessage)
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
			program.setActionsPopupErrorMessage(openLinkOpenerUnavailableMessage)
			return nil
		}
		return program.applyErrorReportedMessage(openLinkOpenerUnavailableMessage)
	}
	browserURL, ok := program.selectedNotificationBrowserURL()
	if !ok {
		if program.model != nil && program.model.ActionsPopupVisible() {
			program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
			return nil
		}
		return program.applyErrorReportedMessage(errActionsPopupActionUnavailable.Error())
	}
	program.closeActionsPopupForAcceptedRequest()
	return program.applyOpenBrowserURLRequested(MsgOpenBrowserURLRequested{URL: browserURL, SuccessMessage: notificationOpenBrowserSuccessMessage, FailureMessage: openLinkFailureMessage, Target: program.model.Focus()})
}

func (program *Program) applyRefreshNotificationsRequested() []Cmd {
	program.beginManualNotificationsRefresh()
	return []Cmd{refreshNotificationsCmd{}}
}
