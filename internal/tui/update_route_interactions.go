package tui

func (program *Program) routeBrowserAndClipboardCompletions(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgOpenBrowserURLRequested:
		return handledUpdate(program.applyOpenBrowserURLRequested(actual))
	case MsgOpenBrowserURLFinished:
		program.applyOpenBrowserURLFinished(actual)
		return handledUpdate(nil)
	case MsgClipboardWriteFinished:
		program.applyClipboardWriteFinished(actual)
		return handledUpdate(nil)
	case MsgSelectedDetailClipboardPrepared:
		return handledUpdate(program.applySelectedDetailClipboardPrepared(actual))
	case MsgPullRequestBuildRunPopupClipboardPrepared:
		return handledUpdate(program.applyPullRequestBuildRunPopupClipboardPrepared(actual))
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeURLClipboardBrowserAndLinkFollowUps(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgOpenPullRequestByURLPromptRequested:
		program.applyOpenPullRequestByURLPromptRequested()
		return handledUpdate(nil)
	case MsgReadPullRequestURLFromClipboardRequested:
		return handledUpdate(program.applyReadPullRequestURLFromClipboardRequested())
	case MsgOpenPullRequestByURLSubmitRequested:
		return handledUpdate(program.applyOpenPullRequestByURLSubmitRequested(actual))
	case MsgPullRequestURLReadFromClipboard:
		program.applyPullRequestURLReadFromClipboard(actual)
		return handledUpdate(nil)
	case MsgOpenPullRequestCustomSearchEditorRequested:
		program.applyOpenPullRequestCustomSearchEditorRequested()
		return handledUpdate(nil)
	case MsgOpenPullRequestCommentComposerRequested:
		program.applyOpenPullRequestCommentComposerRequested()
		return handledUpdate(nil)
	case MsgOpenDetailPullRequestCommentRequested:
		program.applyOpenDetailPullRequestCommentRequested()
		return handledUpdate(nil)
	case MsgOpenInlineCommentReplyRequested:
		program.applyOpenInlineCommentReplyRequested()
		return handledUpdate(nil)
	case MsgOpenLinkUnderCursorRequested:
		return handledUpdate(program.applyOpenLinkUnderCursorRequested(actual))
	case MsgOpenLinkUnderCursorResolved:
		return handledUpdate(program.applyOpenLinkUnderCursorResolved(actual))
	case MsgOpenPullRequestBuildRunPopupLinkRequested:
		return handledUpdate(program.applyOpenPullRequestBuildRunPopupLinkRequested(actual))
	case MsgOpenPullRequestBuildRunPopupLinkResolved:
		return handledUpdate(program.applyOpenPullRequestBuildRunPopupLinkResolved(actual))
	case MsgCopySelectedDetailTextRequested:
		return handledUpdate([]Cmd{prepareSelectedDetailClipboardWriteCmd{Target: program.model.Focus()}})
	case MsgCopyPullRequestURLRequested:
		return handledUpdate(program.applyCopyPullRequestURLRequested(actual))
	case MsgCopyPullRequestBuildRunPopupContentRequested:
		return handledUpdate(program.applyCopyPullRequestBuildRunPopupContentRequested(actual))
	case MsgOpenNotificationInBrowserRequested:
		return handledUpdate(program.applyOpenNotificationInBrowserRequested())
	case MsgRefreshActiveViewRequested:
		return handledUpdate(program.applyRefreshActiveViewRequested())
	case MsgRefreshNotificationsRequested:
		return handledUpdate(program.applyRefreshNotificationsRequested())
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeNotificationReviewTreeAndSearchNavigation(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgNotificationReadRequested:
		return handledUpdate(program.applyNotificationReadRequested(actual))
	case MsgNotificationDoneRequested:
		return handledUpdate(program.applyNotificationDoneRequested(actual))
	case MsgAllNotificationsReadRequested:
		return handledUpdate(program.applyAllNotificationsReadRequested())
	case MsgAllNotificationsDoneRequested:
		return handledUpdate(program.applyAllNotificationsDoneRequested())
	case MsgReviewStoryRequested:
		return handledUpdate(program.applyReviewStoryRequested(actual))
	case MsgRepeatActionsPopupSearch:
		program.applyRepeatActionsPopupSearch(actual)
		return handledUpdate(nil)
	case MsgRepeatSideSearch:
		program.applyRepeatSideSearch(actual)
		return handledUpdate(nil)
	case MsgRepeatPullRequestSearch:
		program.applyRepeatPullRequestSearch(actual)
		return handledUpdate(nil)
	case MsgRepeatReviewFileTreeSearch:
		program.applyRepeatReviewFileTreeSearch(actual)
		return handledUpdate(nil)
	case MsgMoveReviewSelection:
		program.applyMoveReviewSelection(actual)
		return handledUpdate(nil)
	case MsgMoveReviewSelectionToTop:
		program.applyMoveReviewSelectionToTop()
		return handledUpdate(nil)
	case MsgMoveReviewSelectionToBottom:
		program.applyMoveReviewSelectionToBottom()
		return handledUpdate(nil)
	case MsgMoveReviewFile:
		program.applyMoveReviewFile(actual)
		return handledUpdate(nil)
	case MsgMoveReviewComment:
		return handledUpdate(program.applyMoveReviewComment(actual))
	case MsgToggleReviewTreeRowVisibility:
		program.applyToggleReviewTreeRowVisibility()
		return handledUpdate(nil)
	case MsgSetAllReviewTreeFolds:
		program.applySetAllReviewTreeFolds(actual)
		return handledUpdate(nil)
	case MsgSearchWordUnderCursor:
		return handledUpdate(program.applySearchWordUnderCursor(actual))
	case MsgRepeatDetailSearchRequested:
		return handledUpdate(program.applyRepeatDetailSearchRequested(actual))
	case MsgDetailSearchWordResolved:
		return handledUpdate(program.applyDetailSearchWordResolved(actual))
	case MsgToggleInlineConversationVisibility:
		return handledUpdate(program.applyToggleInlineConversationVisibility(actual))
	case MsgSetAllDetailFolds:
		return handledUpdate(program.applySetAllDetailFolds(actual))
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeSearchSubmissionAndPopupSearchEditor(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgSubmitSearch:
		if program.pullRequestBuildRunPopupSearchActive() {
			if popup := program.pullRequestBuildRunPopup; popup != nil {
				popup.searchActive = false
				popup.searchQuery = program.currentSearchText()
			}
			program.searchWidget.clearEditor()
			return handledUpdate([]Cmd{detailMotionCmd{Target: detailMotionTargetBuildPopup, Operation: detailMotionOperationFollowSubmittedSearch}})
		}
		if program.activeSearchIsReviewFileTreeSearch() {
			program.applySubmitReviewFileTreeSearch()
			return handledUpdate(nil)
		}

		target := program.model.SearchTarget()
		targetPullRequestTab := program.model.SearchTargetPullRequestTab()
		targetPullRequestIndex := program.model.SelectedPullRequestIndex(targetPullRequestTab)
		program.model.SubmitSearch()
		if target == FocusDetailView {
			program.searchWidget.detailReversed = false
		}
		program.searchWidget.clearEditor()

		commands := []Cmd(nil)
		if target == FocusDetailView {
			commands = append(commands, detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationFollowSubmittedSearch})
		}
		if target == FocusPullRequestsView {
			program.followSubmittedPullRequestSearch(targetPullRequestTab, targetPullRequestIndex)
		}
		return handledUpdate(commands)
	case MsgCancelSearch:
		if program.pullRequestBuildRunPopupSearchActive() {
			if popup := program.pullRequestBuildRunPopup; popup != nil {
				popup.searchActive = false
			}
			program.searchWidget.clearEditor()
			return handledUpdate(nil)
		}
		if program.activeSearchIsReviewFileTreeSearch() {
			program.applyCancelReviewFileTreeSearch()
			return handledUpdate(nil)
		}
		program.model.CancelSearch()
		program.searchWidget.clearEditor()
		return handledUpdate(nil)
	case MsgCloseSearch:
		program.searchWidget.clearEditor()
		return handledUpdate(nil)
	case MsgActionsPopupSearchInputRequested:
		return handledUpdate(program.applyActionsPopupSearchInputRequested(actual))
	case MsgPullRequestSearchesApplied:
		program.applyPullRequestSearchesApplied(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}
