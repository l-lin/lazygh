package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type modalEditorSubmitSuccess interface {
	apply(*Program) []Cmd
}

type pullRequestCommentSubmitSuccess struct {
	Target         pullRequestCommentTarget
	Body           string
	FeedbackTarget Focus
}

func (success pullRequestCommentSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyAppendPullRequestComment(success.Target, success.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: success.FeedbackTarget, Message: pullRequestCommentSuccessMessage})
	return nil
}

type pullRequestReviewCommentSubmitSuccess struct {
	Repository string
	Number     int
}

func (success pullRequestReviewCommentSubmitSuccess) apply(program *Program) []Cmd {
	return actionsPopupAsyncInvalidatePullRequestSuccess{
		Repository:     success.Repository,
		Number:         success.Number,
		InvalidateDiff: true,
		Message:        pullRequestReviewSuccessMessage,
	}.apply(program)
}

type pullRequestRequestChangesSubmitSuccess struct {
	Repository string
	Number     int
}

func (success pullRequestRequestChangesSubmitSuccess) apply(program *Program) []Cmd {
	return actionsPopupAsyncInvalidatePullRequestSuccess{
		Repository:     success.Repository,
		Number:         success.Number,
		InvalidateDiff: true,
		Message:        pullRequestReviewSuccessMessage,
	}.apply(program)
}

type pullRequestTitleEditSubmitSuccess struct {
	Target         pullRequestActionTarget
	Title          string
	FeedbackTarget Focus
}

func (success pullRequestTitleEditSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	return program.applyPullRequestTitleEditApplied(MsgPullRequestTitleEditApplied{Target: success.Target, Title: success.Title, FeedbackTarget: success.FeedbackTarget})
}

type pullRequestDescriptionEditSubmitSuccess struct {
	Target         pullRequestActionTarget
	Body           string
	FeedbackTarget Focus
}

func (success pullRequestDescriptionEditSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	return program.applyPullRequestDescriptionEditApplied(MsgPullRequestDescriptionEditApplied{Target: success.Target, Body: success.Body, FeedbackTarget: success.FeedbackTarget})
}

type pullRequestCommentUpdateSubmitSuccess struct {
	Target pullRequestCommentEditActionTarget
	Body   string
}

func (success pullRequestCommentUpdateSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyUpdatePullRequestComment(success.Target, success.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestCommentUpdatedSuccessMessage})
	return nil
}

type inlineCommentUpdateSubmitSuccess struct {
	Target pullRequestReviewCommentActionTarget
	Body   string
}

func (success inlineCommentUpdateSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyUpdateReviewComment(success.Target, success.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: inlineCommentUpdatedSuccessMessage})
	return nil
}

type inlineCommentReplySubmitSuccess struct {
	Target pullRequestReviewThreadReplyTarget
	Body   string
}

func (success inlineCommentReplySubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyAppendInlineCommentReply(success.Target, success.Body)
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestInlineCommentReplySuccessMessage})
	return nil
}

type reviewInlineCommentSubmitSuccess struct {
	Target pullRequestInlineCommentTarget
	Body   string
}

func (success reviewInlineCommentSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.optimisticallyAppendInlineComment(success.Target, success.Body)
	program.detailState.viewState.exitVisualMode()
	program.applyFeedbackSet(MsgFeedbackSet{Target: FocusDetailView, Message: pullRequestReviewInlineCommentSuccessMessage})
	return nil
}

type pendingPullRequestReviewSubmitSuccess struct {
	Target pendingPullRequestReviewTarget
}

func (success pendingPullRequestReviewSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.applyPendingPullRequestReviewSubmitted(MsgPendingPullRequestReviewSubmitted{Target: success.Target})
	return nil
}

type pullRequestCustomSearchSubmitSuccess struct {
	Criteria string
}

func (success pullRequestCustomSearchSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	return program.applyPullRequestCustomSearchSubmitted(MsgPullRequestCustomSearchSubmitted{Criteria: success.Criteria})
}

type openPullRequestByURLSubmitSuccess struct {
	Summary githubdomain.PullRequest
}

func (success openPullRequestByURLSubmitSuccess) apply(program *Program) []Cmd {
	if program == nil {
		return nil
	}
	program.applyOpenPullRequestInBrowserView(MsgOpenPullRequestInBrowserView{Summary: success.Summary})
	return nil
}
