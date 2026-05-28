package tui

import "strings"

func (program *Program) applyOpenPullRequestByURLPromptRequested() {
	if program == nil {
		return
	}

	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.mainPaneActionBlocked() || program.actionContext().IsReviewContext() {
		return
	}
	program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: openPullRequestByURLPromptDescriptor()})
}

func (program *Program) applyReadPullRequestURLFromClipboardRequested() []Cmd {
	if program == nil {
		return nil
	}

	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.mainPaneActionBlocked() || program.actionContext().IsReviewContext() {
		return nil
	}
	return []Cmd{readPullRequestURLFromClipboardCmd{}}
}

func (program *Program) applyOpenPullRequestCustomSearchEditorRequested() {
	if program == nil {
		return
	}

	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() || program.actionContext().IsReviewContext() {
		return
	}
	program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: program.pullRequestCustomSearchEditorDescriptor()})
}

func (program *Program) applyOpenPullRequestCommentComposerRequested() {
	if program == nil {
		return
	}

	program.clearPendingSelectionPrefix()
	if program.pullRequestCommentComposerBlocked() {
		return
	}

	target, ok := program.selectedPullRequestCommentTarget()
	if !ok {
		return
	}
	program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: newPullRequestCommentComposerOpenDescriptor(target, program.model.Focus())})
}

func (program *Program) applyOpenDetailPullRequestCommentRequested() {
	if program == nil {
		return
	}

	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.pullRequestCommentComposerBlocked() {
		return
	}

	switch program.inputContext().DetailInputMode {
	case DetailInputModeReviewInlineComment:
		selection, err := program.selectedReviewInlineCommentSelection()
		if err != nil {
			program.setFeedback(FocusDetailView, strings.TrimSpace(err.Error()))
			return
		}
		program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: newReviewInlineCommentOpenDescriptor(selection)})
	case DetailInputModeBrowserChangesInlineComment:
		selection, err := program.selectedBrowserChangesInlineCommentSelection()
		if err != nil {
			program.setFeedback(FocusDetailView, strings.TrimSpace(err.Error()))
			return
		}
		program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: newReviewInlineCommentOpenDescriptor(selection)})
	case DetailInputModePullRequestComment:
		target, ok := program.selectedPullRequestCommentTarget()
		if !ok {
			return
		}
		program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: newPullRequestCommentComposerOpenDescriptor(target, program.model.Focus())})
	}
}

func (program *Program) applyOpenInlineCommentReplyRequested() {
	if program == nil {
		return
	}

	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.overlayState.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
		return
	}
	if !program.inlineCommentReplyShortcutContextActive() {
		return
	}

	target, ok := program.selectedPullRequestReviewThreadReplyTarget()
	if !ok {
		program.setFeedback(FocusDetailView, inlineCommentReplyUnavailableMessage)
		return
	}
	program.applyModalEditorOpened(MsgModalEditorOpened{Descriptor: newInlineCommentReplyOpenDescriptor(target)})
}

func (program *Program) applyRefreshActiveViewRequested() []Cmd {
	if program == nil {
		return nil
	}

	program.clearPendingSelectionPrefix()
	state := program.screenState()
	switch state.ActiveView().Number {
	case sidePanelUserViewNumber:
		return nil
	case sidePanelPullRequestsViewNumber:
		if state.Mode != ScreenModeBrowser {
			return nil
		}
		return program.applyRefreshPullRequestListRequested()
	case sidePanelNotificationsViewNumber:
		return program.applyRefreshNotificationsRequested()
	case mainPanelViewNumber:
		if !program.actionContext().IsPullRequestContext() {
			return nil
		}
		target, ok := program.selectedPullRequestActionTarget()
		if !ok {
			return nil
		}
		summary, ok := program.currentPullRequestSummary()
		if !ok {
			return nil
		}
		return program.applyRefreshPullRequestRequested(MsgRefreshPullRequestRequested{Target: target, Summary: summary})
	default:
		return nil
	}
}

func (program *Program) applyExecuteSelectedActionsPopupActionRequested() []Cmd {
	if program == nil {
		return nil
	}

	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.ghCommandLoading() {
		return nil
	}
	program.syncVisibleActionsPopupSearchSelection()

	action, ok := program.selectedActionsPopupAction()
	if !ok {
		return nil
	}
	return program.applyActionsPopupActionRequested(MsgActionsPopupActionRequested{Action: action})
}

func (program *Program) applySubmitSelectedActionsPopupActionRequested() []Cmd {
	if program == nil {
		return nil
	}

	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() {
		return nil
	}
	program.syncVisibleActionsPopupSearchSelection()
	if !program.assigneePickerVisible() {
		return program.applyExecuteSelectedActionsPopupActionRequested()
	}

	repository := program.actionsPopupWidget.assigneePicker.target.repository
	number := program.actionsPopupWidget.assigneePicker.target.number
	addLogins, removeLogins := program.actionsPopupWidget.assigneePicker.selectedDiff()
	return program.applySubmitAssigneePickerRequested(MsgSubmitAssigneePickerRequested{Repository: repository, Number: number, AddLogins: addLogins, RemoveLogins: removeLogins})
}
