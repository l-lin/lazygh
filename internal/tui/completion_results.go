package tui

type actionsPopupAsyncCompletion interface {
	isActionsPopupAsyncCompletion()
}

type modalEditorSubmitCompletion interface {
	isModalEditorSubmitCompletion()
}

type reviewSessionStartedCompletion MsgReviewSessionStarted

type feedbackSetCompletion MsgFeedbackSet

type pullRequestInvalidatedWithFeedbackCompletion MsgPullRequestInvalidatedWithFeedback

type pullRequestLifecycleAppliedCompletion MsgPullRequestLifecycleApplied

type pullRequestAutoMergeAppliedCompletion MsgPullRequestAutoMergeApplied

type pullRequestBranchUpdatedCompletion MsgPullRequestBranchUpdated

type pullRequestAssigneesUpdatedCompletion MsgPullRequestAssigneesUpdated

type reactionAddedCompletion MsgReactionAdded

type pendingPullRequestReviewCanceledCompletion MsgPendingPullRequestReviewCanceled

type pullRequestCommentDeletedCompletion MsgPullRequestCommentDeleted

type inlineCommentDeletedCompletion MsgInlineCommentDeleted

type inlineCommentResolutionAppliedCompletion MsgInlineCommentResolutionApplied

type reactionRemovedCompletion MsgReactionRemoved

type openPullRequestInBrowserViewCompletion MsgOpenPullRequestInBrowserView

type pullRequestCustomSearchSubmittedCompletion MsgPullRequestCustomSearchSubmitted

type pullRequestCommentSubmittedCompletion MsgPullRequestCommentSubmitted

type pullRequestTitleEditAppliedCompletion MsgPullRequestTitleEditApplied

type pullRequestDescriptionEditAppliedCompletion MsgPullRequestDescriptionEditApplied

type pullRequestCommentUpdatedCompletion MsgPullRequestCommentUpdated

type inlineCommentUpdatedCompletion MsgInlineCommentUpdated

type inlineCommentReplySubmittedCompletion MsgInlineCommentReplySubmitted

type reviewInlineCommentPendingReviewPreparedCompletion MsgReviewInlineCommentPendingReviewPrepared

type reviewInlineCommentSubmittedCompletion MsgReviewInlineCommentSubmitted

type pendingPullRequestReviewSubmittedCompletion MsgPendingPullRequestReviewSubmitted

func (reviewSessionStartedCompletion) isActionsPopupAsyncCompletion()               {}
func (feedbackSetCompletion) isActionsPopupAsyncCompletion()                        {}
func (pullRequestInvalidatedWithFeedbackCompletion) isActionsPopupAsyncCompletion() {}
func (pullRequestInvalidatedWithFeedbackCompletion) isModalEditorSubmitCompletion() {}
func (pullRequestLifecycleAppliedCompletion) isActionsPopupAsyncCompletion()        {}
func (pullRequestAutoMergeAppliedCompletion) isActionsPopupAsyncCompletion()        {}
func (pullRequestBranchUpdatedCompletion) isActionsPopupAsyncCompletion()           {}
func (pullRequestAssigneesUpdatedCompletion) isActionsPopupAsyncCompletion()        {}
func (reactionAddedCompletion) isActionsPopupAsyncCompletion()                      {}
func (pendingPullRequestReviewCanceledCompletion) isActionsPopupAsyncCompletion()   {}
func (pullRequestCommentDeletedCompletion) isActionsPopupAsyncCompletion()          {}
func (inlineCommentDeletedCompletion) isActionsPopupAsyncCompletion()               {}
func (inlineCommentResolutionAppliedCompletion) isActionsPopupAsyncCompletion()     {}
func (reactionRemovedCompletion) isActionsPopupAsyncCompletion()                    {}

func (openPullRequestInBrowserViewCompletion) isModalEditorSubmitCompletion()             {}
func (pullRequestCustomSearchSubmittedCompletion) isModalEditorSubmitCompletion()         {}
func (pullRequestCommentSubmittedCompletion) isModalEditorSubmitCompletion()              {}
func (pullRequestTitleEditAppliedCompletion) isModalEditorSubmitCompletion()              {}
func (pullRequestDescriptionEditAppliedCompletion) isModalEditorSubmitCompletion()        {}
func (pullRequestCommentUpdatedCompletion) isModalEditorSubmitCompletion()                {}
func (inlineCommentUpdatedCompletion) isModalEditorSubmitCompletion()                     {}
func (inlineCommentReplySubmittedCompletion) isModalEditorSubmitCompletion()              {}
func (reviewInlineCommentPendingReviewPreparedCompletion) isModalEditorSubmitCompletion() {}
func (reviewInlineCommentSubmittedCompletion) isModalEditorSubmitCompletion()             {}
func (pendingPullRequestReviewSubmittedCompletion) isModalEditorSubmitCompletion()        {}
