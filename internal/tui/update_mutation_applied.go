package tui

func (program *Program) applyPullRequestLifecycleApplied(message MsgPullRequestLifecycleApplied) {
	program.applyVisiblePullRequestLifecycleMutation(message.Summary, message.State, message.IsDraft)
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestAutoMergeApplied(message MsgPullRequestAutoMergeApplied) {
	program.applyVisiblePullRequestAutoMergeMutation(message.Summary, message.Enabled)
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestMergeWhenReadyApplied(message MsgPullRequestMergeWhenReadyApplied) {
	program.applyVisiblePullRequestMergeWhenReady(message.Summary, message.OptimisticState)
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestMergeQueueApplied(message MsgPullRequestMergeQueueApplied) {
	program.applyVisiblePullRequestMergeQueueMutation(message.Summary, message.InQueue)
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestBranchUpdated(message MsgPullRequestBranchUpdated) {
	program.applyVisiblePullRequestBranchUpdate(message.Summary)
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestInvalidatedWithFeedback(message MsgPullRequestInvalidatedWithFeedback) {
	program.invalidatePullRequestDetail(message.Repository, message.Number)
	if message.InvalidateDiff {
		program.invalidatePullRequestDiff(message.Repository, message.Number)
	}
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestAssigneesUpdated(message MsgPullRequestAssigneesUpdated) {
	program.optimisticallyUpdatePullRequestAssignees(message.Repository, message.Number, message.AddLogins, message.RemoveLogins)
	program.setSuccessFeedback(message.FeedbackTarget, message.Message)
}

func (program *Program) applyPullRequestCommentDeleted(message MsgPullRequestCommentDeleted) {
	program.optimisticallyDeletePullRequestComment(message.Target)
	program.setSuccessFeedback(FocusDetailView, pullRequestCommentDeletedSuccessMessage)
}

func (program *Program) applyInlineCommentDeleted(message MsgInlineCommentDeleted) {
	program.optimisticallyDeleteReviewComment(message.Target)
	program.setSuccessFeedback(FocusDetailView, inlineCommentDeletedSuccessMessage)
}

func (program *Program) applyInlineCommentResolutionApplied(message MsgInlineCommentResolutionApplied) {
	program.optimisticallySetReviewThreadResolved(message.Target, message.Resolved)
	feedbackMessage := inlineCommentResolvedSuccessMessage
	if !message.Resolved {
		feedbackMessage = inlineCommentUnresolvedSuccessMessage
	}
	program.setSuccessFeedback(message.FeedbackTarget, feedbackMessage)
}

func (program *Program) applyReviewSessionStarted(message MsgReviewSessionStarted) {
	program.startReviewSession(message.Summary, message.PendingReviewID)
}

func (program *Program) applyReactionAdded(message MsgReactionAdded) {
	program.optimisticallyAddReaction(message.Target, message.Content)
	program.setSuccessFeedback(message.FeedbackTarget, pullRequestReactionAddedSuccessMessage)
}

func (program *Program) applyReactionRemoved(message MsgReactionRemoved) {
	program.optimisticallyRemoveReaction(message.Target)
	program.setSuccessFeedback(message.FeedbackTarget, pullRequestReactionRemovedSuccessMessage)
}

func (program *Program) applyPendingPullRequestReviewCanceled(message MsgPendingPullRequestReviewCanceled) []Cmd {
	program.invalidatePullRequestDetail(message.Target.repository, message.Target.number)
	program.invalidatePullRequestDiff(message.Target.repository, message.Target.number)
	program.setPendingPullRequestReviewStateByIdentity(message.Target.repository, message.Target.number, "")
	program.setSuccessFeedback(message.Target.sourceFocus, pendingPullRequestReviewCanceledMessage)
	return []Cmd{reloadPullRequestsTabCmd{tab: program.model.ActivePullRequestTab()}}
}

func (program *Program) applyPullRequestCommentSubmitted(message MsgPullRequestCommentSubmitted) {
	program.optimisticallyAppendPullRequestComment(message.Target, message.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: message.FeedbackTarget, Message: pullRequestCommentSuccessMessage, Success: true})
}

func (program *Program) applyPullRequestCommentUpdated(message MsgPullRequestCommentUpdated) {
	program.optimisticallyUpdatePullRequestComment(message.Target, message.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestCommentUpdatedSuccessMessage, Success: true})
}

func (program *Program) applyInlineCommentUpdated(message MsgInlineCommentUpdated) {
	program.optimisticallyUpdateReviewComment(message.Target, message.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: inlineCommentUpdatedSuccessMessage, Success: true})
}

func (program *Program) applyInlineCommentReplySubmitted(message MsgInlineCommentReplySubmitted) {
	program.optimisticallyAppendInlineCommentReply(message.Target, message.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestInlineCommentReplySuccessMessage, Success: true})
}

func (program *Program) applyReviewInlineCommentSubmitted(message MsgReviewInlineCommentSubmitted) {
	program.setPendingPullRequestReviewStateByIdentity(message.Target.repository, message.Target.number, message.Target.pendingReview)
	program.optimisticallyAppendInlineComment(message.Target, message.Body)
	program.exitDetailVisualMode()
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestReviewInlineCommentSuccessMessage, Success: true})
}
