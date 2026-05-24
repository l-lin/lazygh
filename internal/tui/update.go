package tui

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
			program.searchEditor = newLineEditor(actual.Query)
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
		program.searchEditor = newLineEditor(actual.Query)
	case MsgSearchDraftChanged:
		program.updateActiveSearchDraft(actual.Query)
	case MsgSubmitSearch:
		if program.pullRequestBuildRunPopupSearchActive() {
			_ = program.submitPullRequestBuildRunPopupSearch(nil)
			return nil
		}
		if program.activeSearchIsReviewFileTreeSearch() {
			program.submitReviewFileTreeSearch()
			program.searchEditor = nil
			return nil
		}

		target := program.model.SearchTarget()
		targetPullRequestTab := program.model.SearchTargetPullRequestTab()
		targetPullRequestIndex := program.model.SelectedPullRequestIndex(targetPullRequestTab)
		program.model.SubmitSearch()
		if target == FocusDetailView {
			program.detailSearchReversed = false
		}
		program.searchEditor = nil

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
			program.searchEditor = nil
			return nil
		}
		program.model.CancelSearch()
		program.searchEditor = nil
	case MsgCloseSearch:
		program.searchEditor = nil
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
		program.reactionPicker = nil
		program.themePicker = nil
		program.assigneePicker = nil
		program.assigneePickerLoad = nil
		program.model.OpenActionsPopup(actual.ActionCount)
		program.actionsPopupSearchEditor = nil
		program.actionsPopupErrorMessage = ""
	case MsgCloseActionsPopup:
		program.clearPendingSelectionPrefix()
		program.model.CloseActionsPopup()
		program.actionsPopupSearchEditor = nil
		program.clearActionsPopupPendingConfirmation()
		program.actionsPopupErrorMessage = ""
		program.reactionPicker = nil
		program.themePicker = nil
		program.assigneePicker = nil
		program.assigneePickerLoad = nil
	case MsgFocusActionsPopupSearch:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.model.ClearPaneSearchQueries()
		program.clearActionsPopupPendingConfirmation()
		program.actionsPopupSearchEditor = newLineEditor("")
		program.updateActionsPopupSearch("")
		program.model.FocusActionsPopupSearch()
		program.actionsPopupErrorMessage = ""
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
		program.actionsPopupErrorMessage = ""
	case MsgMoveActionsPopupSelectionToTop:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		program.model.MoveActionsPopupSelectionToTop()
		program.actionsPopupErrorMessage = ""
	case MsgMoveActionsPopupSelectionToBottom:
		program.clearPendingSelectionPrefix()
		if !program.model.ActionsPopupVisible() {
			return nil
		}
		program.clearActionsPopupPendingConfirmation()
		program.model.MoveActionsPopupSelectionToBottom()
		program.actionsPopupErrorMessage = ""
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
