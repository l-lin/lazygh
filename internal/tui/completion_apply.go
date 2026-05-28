package tui

func (program *Program) applyActionsPopupAsyncCompletion(completion actionsPopupAsyncCompletion) []Cmd {
	switch actual := completion.(type) {
	case nil:
		return nil
	case reviewSessionStartedCompletion:
		program.applyReviewSessionStarted(MsgReviewSessionStarted(actual))
		return nil
	case feedbackSetCompletion:
		program.applyFeedbackSet(MsgFeedbackSet(actual))
		return nil
	case pullRequestInvalidatedWithFeedbackCompletion:
		program.applyPullRequestInvalidatedWithFeedback(MsgPullRequestInvalidatedWithFeedback(actual))
		return nil
	case pullRequestLifecycleAppliedCompletion:
		program.applyPullRequestLifecycleApplied(MsgPullRequestLifecycleApplied(actual))
		return nil
	case pullRequestAutoMergeAppliedCompletion:
		program.applyPullRequestAutoMergeApplied(MsgPullRequestAutoMergeApplied(actual))
		return nil
	case pullRequestBranchUpdatedCompletion:
		program.applyPullRequestBranchUpdated(MsgPullRequestBranchUpdated(actual))
		return nil
	case pullRequestAssigneesUpdatedCompletion:
		program.applyPullRequestAssigneesUpdated(MsgPullRequestAssigneesUpdated(actual))
		return nil
	case reactionAddedCompletion:
		program.applyReactionAdded(MsgReactionAdded(actual))
		return nil
	case pendingPullRequestReviewCanceledCompletion:
		return program.applyPendingPullRequestReviewCanceled(MsgPendingPullRequestReviewCanceled(actual))
	case pullRequestCommentDeletedCompletion:
		program.applyPullRequestCommentDeleted(MsgPullRequestCommentDeleted(actual))
		return nil
	case inlineCommentDeletedCompletion:
		program.applyInlineCommentDeleted(MsgInlineCommentDeleted(actual))
		return nil
	case inlineCommentResolutionAppliedCompletion:
		program.applyInlineCommentResolutionApplied(MsgInlineCommentResolutionApplied(actual))
		return nil
	case reactionRemovedCompletion:
		program.applyReactionRemoved(MsgReactionRemoved(actual))
		return nil
	default:
		return nil
	}
}

func (program *Program) applyModalEditorSubmitCompletion(completion modalEditorSubmitCompletion) []Cmd {
	switch actual := completion.(type) {
	case nil:
		return nil
	case openPullRequestInBrowserViewCompletion:
		program.applyOpenPullRequestInBrowserView(MsgOpenPullRequestInBrowserView(actual))
		return nil
	case pullRequestCustomSearchSubmittedCompletion:
		return program.applyPullRequestCustomSearchSubmitted(MsgPullRequestCustomSearchSubmitted(actual))
	case pullRequestCommentSubmittedCompletion:
		program.applyPullRequestCommentSubmitted(MsgPullRequestCommentSubmitted(actual))
		return nil
	case pullRequestInvalidatedWithFeedbackCompletion:
		program.applyPullRequestInvalidatedWithFeedback(MsgPullRequestInvalidatedWithFeedback(actual))
		return nil
	case pullRequestTitleEditAppliedCompletion:
		return program.applyPullRequestTitleEditApplied(MsgPullRequestTitleEditApplied(actual))
	case pullRequestDescriptionEditAppliedCompletion:
		return program.applyPullRequestDescriptionEditApplied(MsgPullRequestDescriptionEditApplied(actual))
	case pullRequestCommentUpdatedCompletion:
		program.applyPullRequestCommentUpdated(MsgPullRequestCommentUpdated(actual))
		return nil
	case inlineCommentUpdatedCompletion:
		program.applyInlineCommentUpdated(MsgInlineCommentUpdated(actual))
		return nil
	case inlineCommentReplySubmittedCompletion:
		program.applyInlineCommentReplySubmitted(MsgInlineCommentReplySubmitted(actual))
		return nil
	case reviewInlineCommentPendingReviewPreparedCompletion:
		return program.applyReviewInlineCommentPendingReviewPrepared(MsgReviewInlineCommentPendingReviewPrepared(actual))
	case reviewInlineCommentSubmittedCompletion:
		program.applyReviewInlineCommentSubmitted(MsgReviewInlineCommentSubmitted(actual))
		return nil
	case pendingPullRequestReviewSubmittedCompletion:
		program.applyPendingPullRequestReviewSubmitted(MsgPendingPullRequestReviewSubmitted(actual))
		return nil
	default:
		return nil
	}
}

func modalEditorRemainsOpenAfterCompletion(completion modalEditorSubmitCompletion) bool {
	switch completion.(type) {
	case reviewInlineCommentPendingReviewPreparedCompletion:
		return true
	default:
		return false
	}
}
