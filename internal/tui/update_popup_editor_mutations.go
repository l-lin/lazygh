package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyPullRequestCommentSubmitRequested(message MsgPullRequestCommentSubmitRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pullRequestCommentSubmitRequest{target: message.Target, body: message.Body, feedbackTarget: message.FeedbackTarget})
}

func (program *Program) applyPullRequestReviewCommentSubmitRequested(message MsgPullRequestReviewCommentSubmitRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pullRequestReviewCommentSubmitRequest{target: message.Target, body: message.Body, feedbackTarget: message.FeedbackTarget})
}

func (program *Program) applyPullRequestRequestChangesSubmitRequested(message MsgPullRequestRequestChangesSubmitRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pullRequestRequestChangesSubmitRequest{target: message.Target, body: message.Body, feedbackTarget: message.FeedbackTarget})
}

func (program *Program) applyPullRequestTitleEditRequested(message MsgPullRequestTitleEditRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pullRequestTitleEditSubmitRequest{target: message.Target, title: message.Title, feedbackTarget: message.FeedbackTarget})
}

func (program *Program) applyPullRequestDescriptionEditRequested(message MsgPullRequestDescriptionEditRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pullRequestDescriptionEditSubmitRequest{target: message.Target, body: message.Body, feedbackTarget: message.FeedbackTarget})
}

func (program *Program) applyPullRequestCommentUpdateRequested(message MsgPullRequestCommentUpdateRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pullRequestCommentUpdateSubmitRequest{target: message.Target, body: message.Body})
}

func (program *Program) applyPullRequestCommentDeleteRequested(message MsgPullRequestCommentDeleteRequested) []Cmd {
	if strings.TrimSpace(message.Target.commentID) == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if !program.hasPullRequestMutations() {
		program.setActionsPopupErrorMessage("github loader is unavailable")
		return nil
	}

	return program.queueActionsPopupAsyncRequest(deletePullRequestCommentPopupRequest{target: message.Target})
}

func (program *Program) applyInlineCommentUpdateRequested(message MsgInlineCommentUpdateRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(inlineCommentUpdateSubmitRequest{target: message.Target, body: message.Body})
}

func (program *Program) applyInlineCommentDeleteRequested(message MsgInlineCommentDeleteRequested) []Cmd {
	if strings.TrimSpace(message.Target.commentID) == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if !program.hasReviewMutations() {
		program.setActionsPopupErrorMessage("github loader is unavailable")
		return nil
	}

	return program.queueActionsPopupAsyncRequest(deleteInlineCommentPopupRequest{target: message.Target})
}

func (program *Program) applyInlineCommentReplySubmitRequested(message MsgInlineCommentReplySubmitRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(inlineCommentReplySubmitRequest{target: message.Target, body: message.Body})
}

func (program *Program) applyInlineCommentResolutionRequested(message MsgInlineCommentResolutionRequested) []Cmd {
	if strings.TrimSpace(message.Target.threadID) == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if !program.hasReviewMutations() {
		program.setActionsPopupErrorMessage("github loader is unavailable")
		return nil
	}

	previousCollapsed, rollbackCollapsedState := program.currentInlineCommentResolutionCollapsed(message.Target)
	if rollbackCollapsedState {
		program.applyInlineCommentResolutionCollapsed(message.Target, message.Resolved)
	}
	return program.queueActionsPopupAsyncRequest(inlineCommentResolutionPopupRequest{
		target:                 message.Target,
		resolved:               message.Resolved,
		feedbackTarget:         program.model.Focus(),
		previousCollapsed:      previousCollapsed,
		rollbackCollapsedState: rollbackCollapsedState,
	})
}

func (program *Program) applyReviewInlineCommentSubmitRequested(message MsgReviewInlineCommentSubmitRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(reviewInlineCommentSubmitRequest{target: message.Target, body: message.Body})
}

func (program *Program) applyReviewInlineCommentPendingReviewPrepared(message MsgReviewInlineCommentPendingReviewPrepared) []Cmd {
	program.setPendingPullRequestReviewStateByIdentity(message.Target.repository, message.Target.number, message.Target.pendingReview)
	return program.queueModalEditorSubmitRequest(preparedReviewInlineCommentSubmitRequest{target: message.Target, body: message.Body})
}

func (program *Program) applyPendingPullRequestReviewSubmitRequested(message MsgPendingPullRequestReviewSubmitRequested) []Cmd {
	return program.queueModalEditorSubmitRequest(pendingPullRequestReviewSubmitRequest{target: message.Target, event: message.Event, body: message.Body, feedbackTarget: message.FeedbackTarget})
}

func pendingReviewSubmitError(event githubdomain.PullRequestReviewEvent, feedbackTarget Focus, err error) error {
	if event == githubdomain.PullRequestReviewEventRequestChanges {
		return newModalEditorStatusLineError(feedbackTarget, err)
	}
	return err
}

func (program *Program) applyReactionRemovalRequested(message MsgReactionRemovalRequested) []Cmd {
	if strings.TrimSpace(message.Target.subjectID) == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if !reactionGroupViewerHasReacted(message.Target.reactionGroups, message.Target.content) {
		program.applyActionsPopupClosedWithFeedback(MsgActionsPopupClosedWithFeedback{Target: program.model.Focus(), Message: pullRequestReactionAlreadyRemovedMessage})
		return nil
	}
	if !program.hasReactionMutations() {
		program.setActionsPopupErrorMessage("github loader is unavailable")
		return nil
	}

	return program.queueActionsPopupAsyncRequest(removeReactionPopupRequest{target: message.Target, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyPullRequestSquashMergeRequested(message MsgPullRequestSquashMergeRequested) []Cmd {
	if strings.TrimSpace(program.actionsPopupWidget.pendingConfirmationActionID) != squashMergePullRequestActionTitle {
		program.setActionsPopupPendingConfirmation(squashMergePullRequestActionTitle)
		return nil
	}
	program.clearActionsPopupPendingConfirmation()

	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if !program.hasPullRequestMutations() {
		program.setActionsPopupErrorMessage("github loader is unavailable")
		return nil
	}

	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	return program.queueActionsPopupAsyncRequest(pullRequestSquashMergePopupRequest{repository: repository, number: number, summary: message.Summary, feedbackTarget: program.model.Focus()})
}
