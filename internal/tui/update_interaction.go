package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyActionsPopupActionErrorHandled(message MsgActionsPopupActionErrorHandled) {
	if message.Err == nil {
		return
	}
	if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
		program.actionsPopupWidget.errorMessage = ""
		program.reportError(program.gui, popupMessage)
		return
	}
	var feedbackErr actionsPopupStatusLineError
	if errors.As(message.Err, &feedbackErr) {
		program.actionsPopupWidget.errorMessage = ""
		program.setFeedback(feedbackErr.feedbackTarget, message.Err.Error())
		return
	}
	program.actionsPopupWidget.errorMessage = strings.TrimSpace(message.Err.Error())
	program.reportError(program.gui, program.actionsPopupWidget.errorMessage)
}

func (program *Program) applyActionsPopupClosedWithFeedback(message MsgActionsPopupClosedWithFeedback) {
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	program.setFeedback(message.Target, strings.TrimSpace(message.Message))
}

func (program *Program) applyModalEditorSubmitRequested() []Cmd {
	if program == nil || program.overlayState.modalEditor == nil {
		return nil
	}

	program.overlayState.modalEditor.errorMessage = ""
	var afterSubmit func(*Program)
	if callback := program.overlayState.modalEditor.afterSubmit; callback != nil {
		afterSubmit = func(program *Program) {
			callback(program.gui)
		}
	}
	return []Cmd{modalEditorSubmitCmd{Text: program.overlayState.modalEditor.Text(), Submit: program.overlayState.modalEditor.submit, AfterSubmit: afterSubmit}}
}

func (program *Program) applyModalEditorSubmitFinished(message MsgModalEditorSubmitFinished) {
	if program == nil || program.overlayState.modalEditor == nil {
		return
	}

	if message.Err != nil {
		if popupMessage, ok := transientErrorPopupActionMessage(message.Err); ok {
			program.overlayState.modalEditor.errorMessage = ""
			program.reportError(program.gui, popupMessage)
			return
		}
		var feedbackErr modalEditorStatusLineError
		if errors.As(message.Err, &feedbackErr) {
			program.setFeedback(feedbackErr.feedbackTarget, message.Err.Error())
			return
		}
		program.overlayState.modalEditor.errorMessage = strings.TrimSpace(message.Err.Error())
		return
	}

	if message.AfterSubmit != nil {
		message.AfterSubmit(program)
	}
	program.overlayState.modalEditor = nil
}

func (program *Program) applyModalEditorExternalEditRequested() []Cmd {
	if program == nil || program.overlayState.modalEditor == nil {
		return nil
	}
	return []Cmd{modalEditorExternalEditCmd{Text: program.overlayState.modalEditor.Text()}}
}

func (program *Program) applyModalEditorExternalEditFinished(message MsgModalEditorExternalEditFinished) {
	if program == nil || program.overlayState.modalEditor == nil {
		return
	}
	if message.Err != nil {
		program.overlayState.modalEditor.errorMessage = strings.TrimSpace(message.Err.Error())
		return
	}

	program.overlayState.modalEditor.errorMessage = ""
	program.setModalEditorTextFromExternalEditor(message.Text)
}

func (program *Program) openModalEditorState(state *modalEditorState) {
	if program == nil {
		return
	}
	program.overlayState.modalEditor = state
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
	return []Cmd{pullRequestBuildRunJobLogLoadCmd{Repository: repository, Check: message.Check}}
}

func (program *Program) applyPullRequestBuildRunPopupOpened(message MsgPullRequestBuildRunPopupOpened) {
	program.openPullRequestBuildRunPopupState(message.Content)
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
	url, ok := program.currentDetailCursorLink(program.resolveView(program.gui, message.View, viewDetailName))
	switch {
	case !ok:
		program.setFeedback(program.model.Focus(), openLinkUnavailableMessage)
		return nil
	case program.linkOpener == nil:
		program.setFeedback(program.model.Focus(), openLinkOpenerUnavailableMessage)
		return nil
	default:
		return program.applyOpenBrowserURLRequested(MsgOpenBrowserURLRequested{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: program.model.Focus()})
	}
}

func (program *Program) applyOpenPullRequestBuildRunPopupLinkRequested(message MsgOpenPullRequestBuildRunPopupLinkRequested) []Cmd {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}
	popup.viewState.clearPendingPrefix()
	if program.linkOpener == nil {
		program.setFeedback(program.model.Focus(), openLinkOpenerUnavailableMessage)
		return nil
	}

	actualView := program.resolveView(program.gui, message.View, viewPullRequestBuildInfoName)
	url, ok := program.currentPullRequestBuildRunPopupLink(actualView)
	if !ok {
		program.setFeedback(program.model.Focus(), openLinkUnavailableMessage)
		return nil
	}
	return program.applyOpenBrowserURLRequested(MsgOpenBrowserURLRequested{URL: url, SuccessMessage: openLinkSuccessMessage, FailureMessage: openLinkFailureMessage, Target: program.model.Focus()})
}

func (program *Program) applyCopyPullRequestURLRequested(message MsgCopyPullRequestURLRequested) []Cmd {
	if program.model.Focus() == FocusDetailView && program.detailState.viewState.mode.isVisual() {
		return program.selectedDetailClipboardWriteCmd(program.resolveView(program.gui, message.View, viewDetailName))
	}

	program.detailState.viewState.clearPendingPrefix()
	url, ok := program.selectedPullRequestURL()
	if !ok {
		program.setFeedback(program.model.Focus(), yankUnavailableMessage)
		return nil
	}
	return []Cmd{writeClipboardCmd{Text: url, SuccessMessage: yankSuccessMessage, FailureMessage: yankFailureMessage, Target: program.model.Focus()}}
}

func (program *Program) applyCopyPullRequestBuildRunPopupContentRequested(message MsgCopyPullRequestBuildRunPopupContentRequested) []Cmd {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(program.gui, message.View, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)
	popup.viewState.clearPendingPrefix()
	if popup.viewState.mode.isVisual() {
		selection, _ := detailSelectionForCurrentMode(popup.viewState, document)
		text := popup.viewState.selectedText(document)
		popup.viewState.exitVisualMode()
		return []Cmd{writeClipboardCmd{Text: text, SuccessMessage: detailYankSuccessMessage, FailureMessage: detailYankFailureMessage, Target: program.model.Focus(), Selection: selection, SelectionTarget: clipboardWriteSelectionBuildPopup}}
	}

	trimmedRunURL := strings.TrimSpace(popup.runURL)
	if trimmedRunURL == "" {
		program.setFeedback(program.model.Focus(), yankUnavailableMessage)
		return nil
	}
	return []Cmd{writeClipboardCmd{Text: trimmedRunURL, SuccessMessage: yankSuccessMessage, FailureMessage: yankFailureMessage, Target: program.model.Focus()}}
}

func (program *Program) selectedDetailClipboardWriteCmd(view *gocui.View) []Cmd {
	detailDocument := program.currentDetailDocument(view)
	program.syncDetailViewState(detailDocument, viewPageSize(view))
	selection, _ := detailSelectionForCurrentMode(program.detailState.viewState, detailDocument)
	text := program.detailState.viewState.selectedText(detailDocument)
	program.detailState.viewState.exitVisualMode()
	return []Cmd{writeClipboardCmd{Text: text, SuccessMessage: detailYankSuccessMessage, FailureMessage: detailYankFailureMessage, Target: program.model.Focus(), Selection: selection, SelectionTarget: clipboardWriteSelectionDetail}}
}

func (program *Program) applyOpenNotificationInBrowserRequested() []Cmd {
	if program.linkOpener == nil {
		program.reportError(program.gui, openLinkOpenerUnavailableMessage)
		return nil
	}
	browserURL, ok := program.selectedNotificationBrowserURL()
	if !ok {
		program.reportError(program.gui, errActionsPopupActionUnavailable.Error())
		return nil
	}
	return program.applyOpenBrowserURLRequested(MsgOpenBrowserURLRequested{URL: browserURL, SuccessMessage: notificationOpenBrowserSuccessMessage, FailureMessage: openLinkFailureMessage, Target: program.model.Focus()})
}
