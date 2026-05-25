package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type MsgPullRequestCommentSubmitRequested struct {
	Target         pullRequestCommentTarget
	Body           string
	FeedbackTarget Focus
}

type MsgPullRequestReviewCommentSubmitRequested struct {
	Target         pullRequestActionTarget
	Body           string
	FeedbackTarget Focus
}

type MsgPullRequestRequestChangesSubmitRequested struct {
	Target         pullRequestActionTarget
	Body           string
	FeedbackTarget Focus
}

type MsgPullRequestTitleEditRequested struct {
	Target         pullRequestActionTarget
	Title          string
	FeedbackTarget Focus
}

type MsgPullRequestDescriptionEditRequested struct {
	Target         pullRequestActionTarget
	Body           string
	FeedbackTarget Focus
}

type MsgPullRequestCommentUpdateRequested struct {
	Target pullRequestCommentEditActionTarget
	Body   string
}

type MsgPullRequestCommentDeleteRequested struct {
	Target pullRequestCommentEditActionTarget
}

type MsgInlineCommentUpdateRequested struct {
	Target pullRequestReviewCommentActionTarget
	Body   string
}

type MsgInlineCommentDeleteRequested struct {
	Target pullRequestReviewCommentActionTarget
}

type MsgInlineCommentReplySubmitRequested struct {
	Target pullRequestReviewThreadReplyTarget
	Body   string
}

type MsgInlineCommentResolutionRequested struct {
	Target   pullRequestReviewThreadActionTarget
	Resolved bool
}

type MsgReviewInlineCommentSubmitRequested struct {
	Target pullRequestInlineCommentTarget
	Body   string
}

type MsgPendingPullRequestReviewSubmitRequested struct {
	Target         pendingPullRequestReviewTarget
	Event          githubdomain.PullRequestReviewEvent
	Body           string
	FeedbackTarget Focus
}

type MsgReactionRemovalRequested struct {
	Target pullRequestReactionRemovalTarget
}

type MsgPullRequestSquashMergeRequested struct {
	Target  pullRequestActionTarget
	Summary githubdomain.PullRequest
}

func (MsgPullRequestCommentSubmitRequested) isMsg()        {}
func (MsgPullRequestReviewCommentSubmitRequested) isMsg()  {}
func (MsgPullRequestRequestChangesSubmitRequested) isMsg() {}
func (MsgPullRequestTitleEditRequested) isMsg()            {}
func (MsgPullRequestDescriptionEditRequested) isMsg()      {}
func (MsgPullRequestCommentUpdateRequested) isMsg()        {}
func (MsgPullRequestCommentDeleteRequested) isMsg()        {}
func (MsgInlineCommentUpdateRequested) isMsg()             {}
func (MsgInlineCommentDeleteRequested) isMsg()             {}
func (MsgInlineCommentReplySubmitRequested) isMsg()        {}
func (MsgInlineCommentResolutionRequested) isMsg()         {}
func (MsgReviewInlineCommentSubmitRequested) isMsg()       {}
func (MsgPendingPullRequestReviewSubmitRequested) isMsg()  {}
func (MsgReactionRemovalRequested) isMsg()                 {}
func (MsgPullRequestSquashMergeRequested) isMsg()          {}
