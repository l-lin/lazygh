package tui

import "strings"

func Update(program *Program, msg Msg) []Cmd {
	if program == nil || msg == nil {
		return nil
	}

	switch actual := msg.(type) {
	case MsgAppStarted:
		program.appStarted = true
	case MsgNextSideView:
		program.clearPendingSelectionPrefix()
		program.detailViewState.clearPendingPrefix()
		if program.sideViewCyclingBlocked() {
			return nil
		}
		program.applyProjectedScreenState(program.screenState().NextSideView())
	case MsgPreviousSideView:
		program.clearPendingSelectionPrefix()
		program.detailViewState.clearPendingPrefix()
		if program.sideViewCyclingBlocked() {
			return nil
		}
		program.applyProjectedScreenState(program.screenState().PreviousSideView())
	case MsgFocusPanelView:
		program.clearPendingSelectionPrefix()
		if program.mainPaneActionBlocked() {
			return nil
		}
		program.detailViewState.clearPendingPrefix()
		state := program.screenState()
		targetView, ok := state.ViewByNumber(actual.Number)
		if !ok {
			return nil
		}
		if program.model.PaneLayoutSize() == PaneLayoutFullscreen && program.model.FullscreenPane() != targetView.Focus {
			return nil
		}
		program.applyProjectedScreenState(state.FocusViewNumber(actual.Number))
	case MsgMoveSideSelection:
		program.clearPendingSelectionPrefix()
		if program.selectionChangeBlocked() {
			return nil
		}
		program.model.adjustSelectionBy(actual.Delta)
	case MsgMoveSideSelectionToTop:
		program.clearPendingSelectionPrefix()
		if program.selectionChangeBlocked() {
			return nil
		}
		program.model.MoveSelectionToTop()
	case MsgMoveSideSelectionToBottom:
		program.clearPendingSelectionPrefix()
		if program.selectionChangeBlocked() {
			return nil
		}
		program.model.MoveSelectionToBottom()
	case MsgOpenSearch:
		program.clearPendingSelectionPrefix()
		if program.pullRequestBuildRunPopupVisible() {
			program.startPullRequestBuildRunPopupSearch()
			program.searchWidget.editor = newLineEditor(actual.Query)
			return nil
		}
		inputContext := program.inputContext()
		if program.mainPaneActionBlocked() || (inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusUserView) {
			return nil
		}
		program.detailViewState.clearPendingPrefix()
		if inputContext.SearchUsesReviewTree {
			program.startReviewFileTreeSearch()
		} else {
			if inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusDetailView {
				program.reviewSession.fileTreeSearchQuery = ""
			}
			program.model.StartSearch()
		}
		program.updateActiveSearchDraft(actual.Query)
		program.searchWidget.editor = newLineEditor(actual.Query)
	case MsgSearchDraftChanged:
		program.updateActiveSearchDraft(actual.Query)
	case MsgFeedbackSet:
		program.applyFeedbackSet(actual)
	case MsgActionsPopupActionResultHandled:
		program.applyActionsPopupActionResultHandled(actual)
	case MsgModalEditorOpened:
		program.openModalEditorState(actual.State)
	case MsgModalEditorClosed:
		program.modalEditor = nil
	case MsgModalEditorSubmitRequested:
		return program.applyModalEditorSubmitRequested()
	case MsgModalEditorSubmitFinished:
		program.applyModalEditorSubmitFinished(actual)
	case MsgModalEditorExternalEditRequested:
		return program.applyModalEditorExternalEditRequested()
	case MsgModalEditorExternalEditFinished:
		program.applyModalEditorExternalEditFinished(actual)
	case MsgPullRequestBuildRunLoadRequested:
		return program.applyPullRequestBuildRunLoadRequested(actual)
	case MsgPullRequestBuildRunJobLogLoadRequested:
		return program.applyPullRequestBuildRunJobLogLoadRequested(actual)
	case MsgPullRequestBuildRunPopupOpened:
		program.applyPullRequestBuildRunPopupOpened(actual)
	case MsgPullRequestBuildRunPopupClosed:
		program.applyPullRequestBuildRunPopupClosed()
	case MsgAdvanceDetailTab:
		program.applyAdvanceDetailTab(actual)
	case MsgExitReviewMode:
		program.applyExitReviewMode()
	case MsgToggleHelp:
		program.applyToggleHelp()
	case MsgCloseHelp:
		program.applyCloseHelp()
	case MsgAdjustFocusedPane:
		program.applyAdjustFocusedPane(actual)
	case MsgOpenBrowserURLRequested:
		return program.applyOpenBrowserURLRequested(actual)
	case MsgOpenBrowserURLFinished:
		program.applyOpenBrowserURLFinished(actual)
	case MsgClipboardWriteFinished:
		program.applyClipboardWriteFinished(actual)
	case MsgReadPullRequestURLFromClipboardRequested:
		return []Cmd{readPullRequestURLFromClipboardCmd{}}
	case MsgPullRequestURLReadFromClipboard:
		program.applyPullRequestURLReadFromClipboard(actual)
	case MsgOpenLinkUnderCursorRequested:
		return program.applyOpenLinkUnderCursorRequested(actual)
	case MsgOpenPullRequestBuildRunPopupLinkRequested:
		return program.applyOpenPullRequestBuildRunPopupLinkRequested(actual)
	case MsgCopySelectedDetailTextRequested:
		return program.selectedDetailClipboardWriteCmd(program.resolveView(program.gui, actual.View, viewDetailName))
	case MsgCopyPullRequestURLRequested:
		return program.applyCopyPullRequestURLRequested(actual)
	case MsgCopyPullRequestBuildRunPopupContentRequested:
		return program.applyCopyPullRequestBuildRunPopupContentRequested(actual)
	case MsgOpenNotificationInBrowserRequested:
		return program.applyOpenNotificationInBrowserRequested()
	case MsgRepeatActionsPopupSearch:
		program.applyRepeatActionsPopupSearch(actual)
	case MsgRepeatSideSearch:
		program.applyRepeatSideSearch(actual)
	case MsgRepeatPullRequestSearch:
		program.applyRepeatPullRequestSearch(actual)
	case MsgRepeatReviewFileTreeSearch:
		program.applyRepeatReviewFileTreeSearch(actual)
	case MsgMoveReviewSelection:
		program.applyMoveReviewSelection(actual)
	case MsgMoveReviewSelectionToTop:
		program.applyMoveReviewSelectionToTop()
	case MsgMoveReviewSelectionToBottom:
		program.applyMoveReviewSelectionToBottom()
	case MsgMoveReviewFile:
		program.applyMoveReviewFile(actual)
	case MsgMoveReviewComment:
		program.applyMoveReviewComment(actual)
	case MsgToggleReviewTreeRowVisibility:
		program.applyToggleReviewTreeRowVisibility()
	case MsgSetAllReviewTreeFolds:
		program.applySetAllReviewTreeFolds(actual)
	case MsgSearchWordUnderCursor:
		program.applySearchWordUnderCursor(actual)
	case MsgToggleInlineConversationVisibility:
		program.applyToggleInlineConversationVisibility(actual)
	case MsgSetAllDetailFolds:
		program.applySetAllDetailFolds(actual)
	case MsgSubmitSearch:
		if program.pullRequestBuildRunPopupSearchActive() {
			_ = program.submitPullRequestBuildRunPopupSearch(nil)
			return nil
		}
		if program.activeSearchIsReviewFileTreeSearch() {
			program.submitReviewFileTreeSearch()
			program.searchWidget.editor = nil
			return nil
		}

		target := program.model.SearchTarget()
		targetPullRequestTab := program.model.SearchTargetPullRequestTab()
		targetPullRequestIndex := program.model.SelectedPullRequestIndex(targetPullRequestTab)
		program.model.SubmitSearch()
		if target == FocusDetailView {
			program.searchWidget.detailReversed = false
		}
		program.searchWidget.editor = nil

		if target == FocusDetailView {
			_ = program.followSubmittedDetailSearch(nil)
		}
		if target == FocusPullRequestsView {
			program.followSubmittedPullRequestSearch(targetPullRequestTab, targetPullRequestIndex)
		}
	case MsgCancelSearch:
		if program.pullRequestBuildRunPopupSearchActive() {
			_ = program.cancelPullRequestBuildRunPopupSearch(nil)
			return nil
		}
		if program.activeSearchIsReviewFileTreeSearch() {
			program.cancelReviewFileTreeSearch()
			program.searchWidget.editor = nil
			return nil
		}
		program.model.CancelSearch()
		program.searchWidget.editor = nil
	case MsgCloseSearch:
		program.searchWidget.editor = nil
	case MsgActionsPopupSearchEdited:
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		requestID := 0
		if program.assigneePickerVisible() {
			requestID = program.resetAssigneePickerSearch(actual.Query)
		}
		program.updateActionsPopupSearch(actual.Query)
		program.actionsPopupWidget.errorMessage = ""
		if program.assigneePickerVisible() && requestID > 0 && strings.TrimSpace(actual.Query) != "" {
			return []Cmd{assigneePickerSearchCmd{RequestID: requestID, Query: actual.Query, Delay: program.actionsPopupWidget.assigneePickerSearchDebounceDelay, DispatchLoading: true}}
		}
	case MsgConnectedUserLoaded:
		program.applyConnectedUserLoaded(actual)
	case MsgPullRequestsLoaded:
		program.applyPullRequestsLoaded(actual)
	case MsgNotificationsLoaded:
		program.applyNotificationsLoaded(actual)
	case MsgPullRequestDetailLoaded:
		program.applyPullRequestDetailLoaded(actual)
	case MsgPullRequestDiffLoaded:
		program.applyPullRequestDiffLoaded(actual)
	case MsgIssueDetailLoaded:
		program.applyIssueDetailLoaded(actual)
	case MsgReleaseDetailLoaded:
		program.applyReleaseDetailLoaded(actual)
	case MsgCurrentDetailImageHTMLLoaded:
		program.applyCurrentDetailImageHTMLLoaded(actual)
	case MsgCurrentDetailImageLoaded:
		program.applyCurrentDetailImageLoaded(actual)
	case MsgLoadingSpinnerTick:
		program.applyLoadingSpinnerTick()
	case MsgTransientErrorPopupExpired:
		program.applyTransientErrorPopupExpired(actual)
	case MsgActionsPopupAsyncGHCommandFinished:
		program.applyActionsPopupAsyncGHCommandFinished(actual)
	case MsgNotificationMutationStarted:
		program.applyNotificationMutationStarted(actual)
	case MsgNotificationMutationFinished:
		program.applyNotificationMutationFinished(actual)
	case MsgStoryReviewPrepared:
		program.applyStoryReviewPrepared(actual)
	case MsgAssigneePickerSearchLoadingStarted:
		program.applyAssigneePickerSearchLoadingStarted(actual)
	case MsgAssigneePickerSearchLoaded:
		program.applyAssigneePickerSearchLoaded(actual)
	case MsgPullRequestBuildRunLoaded:
		program.applyPullRequestBuildRunLoaded(actual)
	case MsgPullRequestBuildRunJobLogLoaded:
		program.applyPullRequestBuildRunJobLogLoaded(actual)
	case MsgOpenActionsPopup:
		program.clearPendingSelectionPrefix()
		program.detailViewState.clearPendingPrefix()
		program.clearActionsPopupPendingConfirmation()
		if program.helpVisible || program.model.SearchActive() || program.modalEditorVisible() {
			return nil
		}
		if actual.ActionCount <= 0 {
			return nil
		}
		program.actionsPopupWidget.reactionPicker = nil
		program.actionsPopupWidget.themePicker = nil
		program.actionsPopupWidget.assigneePicker = nil
		program.actionsPopupWidget.assigneePickerLoad = nil
		program.model.OpenActionsPopup(actual.ActionCount)
		program.actionsPopupWidget.searchEditor = nil
		program.actionsPopupWidget.errorMessage = ""
	case MsgCloseActionsPopup:
		program.clearPendingSelectionPrefix()
		program.closeActionsPopupState()
	case MsgFocusActionsPopupSearch:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.model.ClearPaneSearchQueries()
		program.clearActionsPopupPendingConfirmation()
		program.actionsPopupWidget.searchEditor = newLineEditor("")
		program.updateActionsPopupSearch("")
		program.model.FocusActionsPopupSearch()
		program.actionsPopupWidget.errorMessage = ""
	case MsgFocusActionsPopupList:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		program.model.BlurActionsPopupSearch()
	case MsgMoveActionsPopupSelection:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		program.moveActionsPopupSelection(actual.Delta)
		program.actionsPopupWidget.errorMessage = ""
	case MsgMoveActionsPopupSelectionToTop:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		program.model.MoveActionsPopupSelectionToTop()
		program.actionsPopupWidget.errorMessage = ""
	case MsgMoveActionsPopupSelectionToBottom:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		program.model.MoveActionsPopupSelectionToBottom()
		program.actionsPopupWidget.errorMessage = ""
	case MsgModalEditorEdited:
		if program.modalEditor == nil {
			return nil
		}
		program.modalEditor.errorMessage = ""
	}

	return nil
}

func (program *Program) moveActionsPopupSelection(delta int) {
	if program == nil || delta == 0 {
		return
	}
	if delta > 0 {
		for range delta {
			program.model.MoveActionsPopupSelectionDown()
		}
		return
	}
	for range -delta {
		program.model.MoveActionsPopupSelectionUp()
	}
}
