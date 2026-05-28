package tui

func (program *Program) routeBootstrapFocusAndSidePaneSelection(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgAppStarted:
		program.startupState.appStarted = true
		return handledUpdate(nil)
	case MsgNextSideView:
		program.clearPendingSelectionPrefix()
		program.detailState.viewState.clearPendingPrefix()
		if program.sideViewCyclingBlocked() {
			return handledUpdate(nil)
		}
		program.applyProjectedScreenState(program.screenState().NextSideView())
		return handledUpdate(nil)
	case MsgPreviousSideView:
		program.clearPendingSelectionPrefix()
		program.detailState.viewState.clearPendingPrefix()
		if program.sideViewCyclingBlocked() {
			return handledUpdate(nil)
		}
		program.applyProjectedScreenState(program.screenState().PreviousSideView())
		return handledUpdate(nil)
	case MsgFocusPanelView:
		program.clearPendingSelectionPrefix()
		if program.mainPaneActionBlocked() {
			return handledUpdate(nil)
		}
		program.detailState.viewState.clearPendingPrefix()
		state := program.screenState()
		targetView, ok := state.ViewByNumber(actual.Number)
		if !ok {
			return handledUpdate(nil)
		}
		if program.model.PaneLayoutSize() == PaneLayoutFullscreen && program.model.FullscreenPane() != targetView.Focus {
			return handledUpdate(nil)
		}
		program.applyProjectedScreenState(state.FocusViewNumber(actual.Number))
		return handledUpdate(nil)
	case MsgMoveSideSelection:
		program.applyMoveSideSelection(actual)
		return handledUpdate(nil)
	case MsgMoveSideSelectionToTop:
		program.applyMoveSideSelectionToTop()
		return handledUpdate(nil)
	case MsgMoveSideSelectionToBottom:
		program.applyMoveSideSelectionToBottom()
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeSearchPromptAndDraftUpdate(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgOpenSearch:
		program.clearPendingSelectionPrefix()
		if program.pullRequestBuildRunPopupVisible() {
			program.startPullRequestBuildRunPopupSearch()
			program.searchWidget.openEditor(actual.Query)
			return handledUpdate(nil)
		}
		inputContext := program.inputContext()
		if inputContext.SearchUsesReviewTree {
			program.applyStartReviewFileTreeSearch(MsgStartReviewFileTreeSearch{Query: actual.Query})
			return handledUpdate(nil)
		}
		if program.mainPaneActionBlocked() || (inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusUserView) {
			return handledUpdate(nil)
		}
		program.detailState.viewState.clearPendingPrefix()
		if inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusDetailView {
			program.model.ClearReviewTreeSearchQuery()
		}
		program.model.StartSearch()
		program.applySearchDraftChanged(MsgSearchDraftChanged{Query: actual.Query})
		program.searchWidget.openEditor(actual.Query)
		return handledUpdate(nil)
	case MsgSearchDraftChanged:
		program.applySearchDraftChanged(actual)
		return handledUpdate(nil)
	case MsgSearchEditorInputRequested:
		program.applySearchEditorInputRequested(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeFeedbackErrorAndModalEditorLifecycle(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgFeedbackSet:
		program.applyFeedbackSet(actual)
		return handledUpdate(nil)
	case MsgErrorReported:
		return handledUpdate(program.applyErrorReported(actual))
	case MsgActionsPopupClosedWithFeedback:
		program.applyActionsPopupClosedWithFeedback(actual)
		return handledUpdate(nil)
	case MsgActionsPopupActionErrorHandled:
		return handledUpdate(program.applyActionsPopupActionErrorHandled(actual))
	case MsgActionsPopupActionRequested:
		return handledUpdate(program.applyActionsPopupActionRequested(actual))
	case MsgModalEditorOpened:
		program.applyModalEditorOpened(actual)
		return handledUpdate(nil)
	case MsgModalEditorLineInputRequested:
		program.applyModalEditorLineInputRequested(actual)
		return handledUpdate(nil)
	case MsgModalEditorMultilineInputRequested:
		program.applyModalEditorMultilineInputRequested(actual)
		return handledUpdate(nil)
	case MsgModalEditorClosed:
		program.overlayState.modalEditor = modalEditorState{}
		return handledUpdate(nil)
	case MsgModalEditorSubmitRequested:
		return handledUpdate(program.applyModalEditorSubmitRequested())
	case MsgModalEditorSubmitFinished:
		return handledUpdate(program.applyModalEditorSubmitFinished(actual))
	case MsgModalEditorExternalEditRequested:
		return handledUpdate(program.applyModalEditorExternalEditRequested())
	case MsgModalEditorExternalEditFinished:
		program.applyModalEditorExternalEditFinished(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeBuildRunPopupLifecycle(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgPullRequestBuildRunLoadRequested:
		return handledUpdate(program.applyPullRequestBuildRunLoadRequested(actual))
	case MsgPullRequestBuildRunJobLogLoadRequested:
		return handledUpdate(program.applyPullRequestBuildRunJobLogLoadRequested(actual))
	case MsgPullRequestBuildRunPopupOpened:
		program.applyPullRequestBuildRunPopupOpened(actual)
		return handledUpdate(nil)
	case MsgPullRequestBuildRunPopupClosed:
		program.applyPullRequestBuildRunPopupClosed()
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}
