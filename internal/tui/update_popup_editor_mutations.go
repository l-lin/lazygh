package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyPullRequestCommentSubmitRequested(message MsgPullRequestCommentSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestCommentSubmitRequest{target: message.Target, body: message.Body, feedbackTarget: message.FeedbackTarget}}}
}

func (program *Program) applyPullRequestReviewCommentSubmitRequested(message MsgPullRequestReviewCommentSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestReviewCommentSubmitRequest{target: message.Target, body: message.Body}}}
}

func (program *Program) applyPullRequestRequestChangesSubmitRequested(message MsgPullRequestRequestChangesSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestRequestChangesSubmitRequest{target: message.Target, body: message.Body}}}
}

func (program *Program) applyPullRequestTitleEditRequested(message MsgPullRequestTitleEditRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestTitleEditSubmitRequest{target: message.Target, title: message.Title, feedbackTarget: message.FeedbackTarget}}}
}

func (program *Program) applyPullRequestDescriptionEditRequested(message MsgPullRequestDescriptionEditRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestDescriptionEditSubmitRequest{target: message.Target, body: message.Body, feedbackTarget: message.FeedbackTarget}}}
}

func (program *Program) applyPullRequestCommentUpdateRequested(message MsgPullRequestCommentUpdateRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pullRequestCommentUpdateSubmitRequest{target: message.Target, body: message.Body}}}
}

func (program *Program) applyPullRequestCommentDeleteRequested(message MsgPullRequestCommentDeleteRequested) []Cmd {
	if strings.TrimSpace(message.Target.commentID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasPullRequestMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return program.queueActionsPopupAsyncRequest(deletePullRequestCommentPopupRequest{target: message.Target})
}

func (program *Program) applyInlineCommentUpdateRequested(message MsgInlineCommentUpdateRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: inlineCommentUpdateSubmitRequest{target: message.Target, body: message.Body}}}
}

func (program *Program) applyInlineCommentDeleteRequested(message MsgInlineCommentDeleteRequested) []Cmd {
	if strings.TrimSpace(message.Target.commentID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasReviewMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return program.queueActionsPopupAsyncRequest(deleteInlineCommentPopupRequest{target: message.Target})
}

func (program *Program) applyInlineCommentReplySubmitRequested(message MsgInlineCommentReplySubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: inlineCommentReplySubmitRequest{target: message.Target, body: message.Body}}}
}

func (program *Program) applyInlineCommentResolutionRequested(message MsgInlineCommentResolutionRequested) []Cmd {
	if strings.TrimSpace(message.Target.threadID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasReviewMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return program.queueActionsPopupAsyncRequest(inlineCommentResolutionPopupRequest{target: message.Target, resolved: message.Resolved, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyReviewInlineCommentSubmitRequested(message MsgReviewInlineCommentSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: reviewInlineCommentSubmitRequest{target: message.Target, body: message.Body}}}
}

func (program *Program) applyPendingPullRequestReviewSubmitRequested(message MsgPendingPullRequestReviewSubmitRequested) []Cmd {
	return []Cmd{modalEditorSubmitCmd{request: pendingPullRequestReviewSubmitRequest{target: message.Target, event: message.Event, body: message.Body, feedbackTarget: message.FeedbackTarget}}}
}

func pendingReviewSubmitError(event githubdomain.PullRequestReviewEvent, feedbackTarget Focus, err error) error {
	if event == githubdomain.PullRequestReviewEventRequestChanges {
		return newModalEditorStatusLineError(feedbackTarget, err)
	}
	return err
}

func (program *Program) applyReactionRemovalRequested(message MsgReactionRemovalRequested) []Cmd {
	if strings.TrimSpace(message.Target.subjectID) == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !reactionGroupViewerHasReacted(message.Target.reactionGroups, message.Target.content) {
		return Update(program, MsgActionsPopupClosedWithFeedback{Target: program.model.Focus(), Message: pullRequestReactionAlreadyRemovedMessage})
	}
	if !program.hasReactionMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	return program.queueActionsPopupAsyncRequest(removeReactionPopupRequest{target: message.Target, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyPullRequestSquashMergeRequested(message MsgPullRequestSquashMergeRequested) []Cmd {
	if strings.TrimSpace(program.actionsPopupWidget.pendingConfirmationActionID) != squashMergePullRequestActionTitle {
		program.actionsPopupWidget.pendingConfirmationActionID = squashMergePullRequestActionTitle
		program.actionsPopupWidget.errorMessage = ""
		return nil
	}
	program.clearActionsPopupPendingConfirmation()

	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}
	if !program.hasPullRequestMutations() {
		program.actionsPopupWidget.errorMessage = "github loader is unavailable"
		return nil
	}

	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	return program.queueActionsPopupAsyncRequest(pullRequestSquashMergePopupRequest{repository: repository, number: number, summary: message.Summary, feedbackTarget: program.model.Focus()})
}
