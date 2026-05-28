package tui

func (program *Program) routeBrowserAndReviewNavigation(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgAdvanceDetailTab:
		program.applyAdvanceDetailTab(actual)
		return handledUpdate(nil)
	case MsgAdvancePullRequestTab:
		return handledUpdate(program.applyAdvancePullRequestTab(actual))
	case MsgOpenDetailRequested:
		program.applyOpenDetailRequested()
		return handledUpdate(nil)
	case MsgCloseDetailRequested:
		program.applyCloseDetailRequested()
		return handledUpdate(nil)
	case MsgStartReviewFileTreeSearch:
		program.applyStartReviewFileTreeSearch(actual)
		return handledUpdate(nil)
	case MsgSubmitReviewFileTreeSearch:
		program.applySubmitReviewFileTreeSearch()
		return handledUpdate(nil)
	case MsgCancelReviewFileTreeSearch:
		program.applyCancelReviewFileTreeSearch()
		return handledUpdate(nil)
	case MsgOpenPullRequestInBrowserView:
		program.applyOpenPullRequestInBrowserView(actual)
		return handledUpdate(nil)
	case MsgOpenPullRequestInDetailFullscreen:
		program.applyOpenPullRequestInDetailFullscreen(actual)
		return handledUpdate(nil)
	case MsgExitReviewMode:
		program.applyExitReviewMode()
		return handledUpdate(nil)
	case MsgToggleHelp:
		program.applyToggleHelp()
		return handledUpdate(nil)
	case MsgCloseHelp:
		program.applyCloseHelp()
		return handledUpdate(nil)
	case MsgHelpPageNavigationRequested:
		return handledUpdate(program.applyHelpPageNavigationRequested(actual))
	case MsgAdjustFocusedPane:
		program.applyAdjustFocusedPane(actual)
		return handledUpdate(nil)
	case MsgLineNavigationRequested:
		return handledUpdate(program.applyLineNavigationRequested(actual))
	case MsgPageNavigationRequested:
		return handledUpdate(program.applyPageNavigationRequested(actual))
	case MsgPageNavigationResolved:
		return handledUpdate(program.applyPageNavigationResolved(actual))
	case MsgSideListViewportRequested:
		return handledUpdate(program.applySideListViewportRequested(actual))
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeDetailMotionAndLiveSync(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgDetailViewportRequested:
		return handledUpdate(program.applyDetailViewportRequested(actual))
	case MsgDetailViewportResolved:
		program.applyDetailViewportResolved(actual)
		return handledUpdate(nil)
	case MsgFocusDetailRenderedLineResolved:
		program.applyFocusDetailRenderedLineResolved(actual)
		return handledUpdate(nil)
	case MsgDetailMotionRequested:
		return handledUpdate(program.applyDetailMotionRequested(actual))
	case MsgDetailYankRequested:
		return handledUpdate(program.applyDetailYankRequested(actual))
	case MsgDetailMotionResolved:
		return handledUpdate(program.applyDetailMotionResolved(actual))
	case MsgToggleInlineConversationVisibilityResolved:
		program.applyToggleInlineConversationVisibilityResolved(actual)
		return handledUpdate(nil)
	case MsgSetAllDetailFoldsResolved:
		program.applySetAllDetailFoldsResolved(actual)
		return handledUpdate(nil)
	case MsgDetailViewSyncPlanResolved:
		program.applyDetailViewSyncPlanResolved(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}
