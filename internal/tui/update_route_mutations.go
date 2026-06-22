package tui

func (program *Program) routePullRequestFeatureRequests(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgClearCacheRequested:
		return handledUpdate(program.applyClearCacheRequested())
	case MsgStartPullRequestReviewRequested:
		return handledUpdate(program.applyStartPullRequestReviewRequested(actual))
	case MsgOpenPullRequestInBrowserRequested:
		return handledUpdate(program.applyOpenPullRequestInBrowserRequested(actual))
	case MsgApprovePullRequestRequested:
		return handledUpdate(program.applyApprovePullRequestRequested(actual))
	case MsgReRequestPullRequestReviewRequested:
		return handledUpdate(program.applyReRequestPullRequestReviewRequested(actual))
	case MsgPullRequestLifecycleMutationRequested:
		return handledUpdate(program.applyPullRequestLifecycleMutationRequested(actual))
	case MsgPullRequestAutoMergeMutationRequested:
		return handledUpdate(program.applyPullRequestAutoMergeMutationRequested(actual))
	case MsgPullRequestMergeWhenReadyRequested:
		return handledUpdate(program.applyPullRequestMergeWhenReadyRequested(actual))
	case MsgPullRequestMergeQueueMutationRequested:
		return handledUpdate(program.applyPullRequestMergeQueueMutationRequested(actual))
	case MsgPullRequestBranchUpdateRequested:
		return handledUpdate(program.applyPullRequestBranchUpdateRequested(actual))
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routeMutationApplyResultsAndOptimisticFollowUp(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgPullRequestLifecycleApplied:
		program.applyPullRequestLifecycleApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestAutoMergeApplied:
		program.applyPullRequestAutoMergeApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestMergeWhenReadyApplied:
		program.applyPullRequestMergeWhenReadyApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestMergeQueueApplied:
		program.applyPullRequestMergeQueueApplied(actual)
		return handledUpdate(nil)
	case MsgPullRequestBranchUpdated:
		program.applyPullRequestBranchUpdated(actual)
		return handledUpdate(nil)
	case MsgPullRequestInvalidatedWithFeedback:
		program.applyPullRequestInvalidatedWithFeedback(actual)
		return handledUpdate(nil)
	case MsgPullRequestAssigneesUpdated:
		program.applyPullRequestAssigneesUpdated(actual)
		return handledUpdate(nil)
	case MsgReviewSessionStarted:
		program.applyReviewSessionStarted(actual)
		return handledUpdate(nil)
	case MsgReactionAdded:
		program.applyReactionAdded(actual)
		return handledUpdate(nil)
	case MsgReactionRemoved:
		program.applyReactionRemoved(actual)
		return handledUpdate(nil)
	case MsgPendingPullRequestReviewCanceled:
		return handledUpdate(program.applyPendingPullRequestReviewCanceled(actual))
	case MsgPullRequestCustomSearchSubmitRequested:
		return handledUpdate(program.applyPullRequestCustomSearchSubmitRequested(actual))
	case MsgPullRequestCustomSearchSubmitted:
		return handledUpdate(program.applyPullRequestCustomSearchSubmitted(actual))
	case MsgOpenAssigneePickerRequested:
		return handledUpdate(program.applyOpenAssigneePickerRequested(actual))
	case MsgToggleAssigneePickerSelectionRequested:
		program.applyToggleAssigneePickerSelectionRequested(actual)
		return handledUpdate(nil)
	case MsgSubmitAssigneePickerRequested:
		return handledUpdate(program.applySubmitAssigneePickerRequested(actual))
	case MsgOpenReactionPickerRequested:
		program.applyOpenReactionPickerRequested(actual)
		return handledUpdate(nil)
	case MsgAddReactionRequested:
		return handledUpdate(program.applyAddReactionRequested(actual))
	case MsgOpenThemePickerRequested:
		program.applyOpenThemePickerRequested()
		return handledUpdate(nil)
	case MsgThemePresetSelected:
		return handledUpdate(program.applyThemePresetSelected(actual))
	case MsgThemePresetSaved:
		return handledUpdate(program.applyThemePresetSaved(actual))
	case MsgPersistentCacheCleared:
		return handledUpdate(program.applyPersistentCacheCleared(actual))
	case MsgRefreshPullRequestListRequested:
		return handledUpdate(program.applyRefreshPullRequestListRequested())
	case MsgRefreshPullRequestRequested:
		return handledUpdate(program.applyRefreshPullRequestRequested(actual))
	case MsgPullRequestTitleEditApplied:
		return handledUpdate(program.applyPullRequestTitleEditApplied(actual))
	case MsgPullRequestDescriptionEditApplied:
		return handledUpdate(program.applyPullRequestDescriptionEditApplied(actual))
	case MsgPullRequestCommentSubmitted:
		program.applyPullRequestCommentSubmitted(actual)
		return handledUpdate(nil)
	case MsgPullRequestCommentUpdated:
		program.applyPullRequestCommentUpdated(actual)
		return handledUpdate(nil)
	case MsgInlineCommentUpdated:
		program.applyInlineCommentUpdated(actual)
		return handledUpdate(nil)
	case MsgInlineCommentReplySubmitted:
		program.applyInlineCommentReplySubmitted(actual)
		return handledUpdate(nil)
	case MsgReviewInlineCommentSubmitted:
		program.applyReviewInlineCommentSubmitted(actual)
		return handledUpdate(nil)
	case MsgPullRequestCommentDeleted:
		program.applyPullRequestCommentDeleted(actual)
		return handledUpdate(nil)
	case MsgInlineCommentDeleted:
		program.applyInlineCommentDeleted(actual)
		return handledUpdate(nil)
	case MsgInlineCommentResolutionApplied:
		program.applyInlineCommentResolutionApplied(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}

func (program *Program) routePopupEditorSubmissionAndMutationRequests(msg Msg) updateResult {
	switch actual := msg.(type) {
	case MsgPullRequestCommentSubmitRequested:
		return handledUpdate(program.applyPullRequestCommentSubmitRequested(actual))
	case MsgPullRequestReviewCommentSubmitRequested:
		return handledUpdate(program.applyPullRequestReviewCommentSubmitRequested(actual))
	case MsgPullRequestRequestChangesSubmitRequested:
		return handledUpdate(program.applyPullRequestRequestChangesSubmitRequested(actual))
	case MsgPullRequestTitleEditRequested:
		return handledUpdate(program.applyPullRequestTitleEditRequested(actual))
	case MsgPullRequestDescriptionEditRequested:
		return handledUpdate(program.applyPullRequestDescriptionEditRequested(actual))
	case MsgPullRequestCommentUpdateRequested:
		return handledUpdate(program.applyPullRequestCommentUpdateRequested(actual))
	case MsgPullRequestCommentDeleteRequested:
		return handledUpdate(program.applyPullRequestCommentDeleteRequested(actual))
	case MsgInlineCommentUpdateRequested:
		return handledUpdate(program.applyInlineCommentUpdateRequested(actual))
	case MsgInlineCommentDeleteRequested:
		return handledUpdate(program.applyInlineCommentDeleteRequested(actual))
	case MsgInlineCommentReplySubmitRequested:
		return handledUpdate(program.applyInlineCommentReplySubmitRequested(actual))
	case MsgInlineCommentResolutionRequested:
		return handledUpdate(program.applyInlineCommentResolutionRequested(actual))
	case MsgReviewInlineCommentSubmitRequested:
		return handledUpdate(program.applyReviewInlineCommentSubmitRequested(actual))
	case MsgReviewInlineCommentPendingReviewPrepared:
		return handledUpdate(program.applyReviewInlineCommentPendingReviewPrepared(actual))
	case MsgPendingPullRequestReviewSubmitRequested:
		return handledUpdate(program.applyPendingPullRequestReviewSubmitRequested(actual))
	case MsgReactionRemovalRequested:
		return handledUpdate(program.applyReactionRemovalRequested(actual))
	case MsgPullRequestSquashMergeRequested:
		return handledUpdate(program.applyPullRequestSquashMergeRequested(actual))
	case MsgCancelPendingPullRequestReviewRequested:
		return handledUpdate(program.applyCancelPendingPullRequestReviewRequested(actual))
	case MsgPendingPullRequestReviewSubmitted:
		program.applyPendingPullRequestReviewSubmitted(actual)
		return handledUpdate(nil)
	default:
		return ignoredUpdate()
	}
}
