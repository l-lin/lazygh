package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgPullRequestLifecycleApplied struct {
	Summary        githubdomain.PullRequest
	State          string
	IsDraft        bool
	FeedbackTarget Focus
	Message        string
}

type MsgPullRequestAutoMergeApplied struct {
	Summary        githubdomain.PullRequest
	Enabled        bool
	FeedbackTarget Focus
	Message        string
}

type MsgPullRequestBranchUpdated struct {
	Summary        githubdomain.PullRequest
	FeedbackTarget Focus
	Message        string
}

type MsgPullRequestInvalidatedWithFeedback struct {
	Repository     string
	Number         int
	InvalidateDiff bool
	FeedbackTarget Focus
	Message        string
}

type MsgPullRequestAssigneesUpdated struct {
	Repository     string
	Number         int
	AddLogins      []string
	RemoveLogins   []string
	FeedbackTarget Focus
	Message        string
}

type MsgPullRequestCommentDeleted struct {
	Target pullRequestCommentEditActionTarget
}

type MsgInlineCommentDeleted struct {
	Target pullRequestReviewCommentActionTarget
}

type MsgInlineCommentResolutionApplied struct {
	Target         pullRequestReviewThreadActionTarget
	Resolved       bool
	FeedbackTarget Focus
}

type MsgReviewSessionStarted struct {
	Summary         githubdomain.PullRequest
	PendingReviewID string
}

type MsgReactionAdded struct {
	Target         pullRequestReactionActionTarget
	Content        githubdomain.ReactionContent
	FeedbackTarget Focus
}

type MsgReactionRemoved struct {
	Target         pullRequestReactionRemovalTarget
	FeedbackTarget Focus
}

type MsgPendingPullRequestReviewCanceled struct {
	Target pendingPullRequestReviewActionTarget
}

type MsgPullRequestCommentSubmitted struct {
	Target         pullRequestCommentTarget
	Body           string
	FeedbackTarget Focus
}

type MsgPullRequestCommentUpdated struct {
	Target pullRequestCommentEditActionTarget
	Body   string
}

type MsgInlineCommentUpdated struct {
	Target pullRequestReviewCommentActionTarget
	Body   string
}

type MsgInlineCommentReplySubmitted struct {
	Target pullRequestReviewThreadReplyTarget
	Body   string
}

type MsgReviewInlineCommentPendingReviewPrepared struct {
	Target pullRequestInlineCommentTarget
	Body   string
}

type MsgReviewInlineCommentSubmitted struct {
	Target pullRequestInlineCommentTarget
	Body   string
}

func (MsgPullRequestLifecycleApplied) isMsg()              {}
func (MsgPullRequestAutoMergeApplied) isMsg()              {}
func (MsgPullRequestBranchUpdated) isMsg()                 {}
func (MsgPullRequestInvalidatedWithFeedback) isMsg()       {}
func (MsgPullRequestAssigneesUpdated) isMsg()              {}
func (MsgPullRequestCommentDeleted) isMsg()                {}
func (MsgInlineCommentDeleted) isMsg()                     {}
func (MsgInlineCommentResolutionApplied) isMsg()           {}
func (MsgReviewSessionStarted) isMsg()                     {}
func (MsgReactionAdded) isMsg()                            {}
func (MsgReactionRemoved) isMsg()                          {}
func (MsgPendingPullRequestReviewCanceled) isMsg()         {}
func (MsgPullRequestCommentSubmitted) isMsg()              {}
func (MsgPullRequestCommentUpdated) isMsg()                {}
func (MsgInlineCommentUpdated) isMsg()                     {}
func (MsgInlineCommentReplySubmitted) isMsg()              {}
func (MsgReviewInlineCommentPendingReviewPrepared) isMsg() {}
func (MsgReviewInlineCommentSubmitted) isMsg()             {}
