package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyOpenPullRequestInBrowserRequested(message MsgOpenPullRequestInBrowserRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(openPullRequestInBrowserPopupRequest{repository: repository, number: number, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyApprovePullRequestRequested(message MsgApprovePullRequestRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(approvePullRequestPopupRequest{repository: repository, number: number, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyReRequestPullRequestReviewRequested(message MsgReRequestPullRequestReviewRequested) []Cmd {
	repository, number, reviewerLogin, ok := popupPullRequestReviewerRequestIdentity(message.Target)
	if !ok {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(reRequestPullRequestReviewPopupRequest{repository: repository, number: number, reviewerLogin: reviewerLogin, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyPullRequestLifecycleMutationRequested(message MsgPullRequestLifecycleMutationRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	command := pullRequestLifecycleMutationCommand(message.Kind, repository, number)
	if command == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(pullRequestLifecycleMutationPopupRequest{kind: message.Kind, repository: repository, number: number, summary: message.Summary, state: message.State, isDraft: message.IsDraft, successMessage: message.SuccessMessage, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyPullRequestAutoMergeMutationRequested(message MsgPullRequestAutoMergeMutationRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	command := pullRequestAutoMergeMutationCommand(message.Kind, repository, number)
	if command == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(pullRequestAutoMergeMutationPopupRequest{kind: message.Kind, repository: repository, number: number, summary: message.Summary, enabled: message.Enabled, successMessage: message.SuccessMessage, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyPullRequestMergeQueueMutationRequested(message MsgPullRequestMergeQueueMutationRequested) []Cmd {
	pullRequestID, ok := popupPullRequestActionTargetPullRequestID(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}
	if pullRequestMergeQueueMutationCommand(message.Kind) == "" {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	snapshot := program.capturePullRequestMergeQueueMutationSnapshot(message.Summary)
	program.clearPendingSelectionPrefix()
	program.closeActionsPopupState()
	program.applyVisiblePullRequestMergeQueueMutation(message.Summary, message.InQueue)
	return program.queueActionsPopupAsyncRequest(pullRequestMergeQueueMutationPopupRequest{kind: message.Kind, pullRequestID: pullRequestID, summary: message.Summary, inQueue: message.InQueue, successMessage: message.SuccessMessage, feedbackTarget: program.model.Focus(), rollbackSnapshot: snapshot})
}

func (program *Program) applyPullRequestBranchUpdateRequested(message MsgPullRequestBranchUpdateRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.setActionsPopupErrorMessage(errActionsPopupActionUnavailable.Error())
		return nil
	}

	return program.queueActionsPopupAsyncRequest(pullRequestBranchUpdatePopupRequest{repository: repository, number: number, summary: message.Summary, feedbackTarget: program.model.Focus()})
}

func (program *Program) applyPendingPullRequestReviewSubmitted(message MsgPendingPullRequestReviewSubmitted) {
	target := message.Target
	program.invalidatePullRequestDetail(target.repository, target.number)
	program.invalidatePullRequestDiff(target.repository, target.number)
	program.setPendingPullRequestReviewStateByIdentity(target.repository, target.number, "")
	program.restorePullRequestBrowserFromReviewMode()
	program.setFeedback(target.sourceFocus, pullRequestReviewSuccessMessage)
}

func popupPullRequestActionTargetIdentity(target pullRequestActionTarget) (string, int, bool) {
	repository := strings.TrimSpace(target.repository)
	if pullRequestKeyFromIdentity(repository, target.number) == "" {
		return "", 0, false
	}
	return repository, target.number, true
}

func popupPullRequestReviewerRequestIdentity(target pullRequestReviewerRequestTarget) (string, int, string, bool) {
	repository := strings.TrimSpace(target.repository)
	reviewerLogin := strings.TrimSpace(target.reviewerLogin)
	if pullRequestKeyFromIdentity(repository, target.number) == "" || reviewerLogin == "" {
		return "", 0, "", false
	}
	return repository, target.number, reviewerLogin, true
}

func popupPullRequestActionTargetPullRequestID(target pullRequestActionTarget) (string, bool) {
	pullRequestID := strings.TrimSpace(target.pullRequestID)
	return pullRequestID, pullRequestID != ""
}

func popupPullRequestSummaryValid(summary githubdomain.PullRequest) bool {
	return pullRequestDetailKey(summary.Repository, summary.Number) != ""
}

func pullRequestLifecycleMutationCommand(kind pullRequestLifecycleMutationKind, repository string, number int) string {
	switch kind {
	case pullRequestLifecycleMutationReadyForReview:
		return pullRequestReadyCommand(repository, number, false)
	case pullRequestLifecycleMutationConvertToDraft:
		return pullRequestReadyCommand(repository, number, true)
	case pullRequestLifecycleMutationClose:
		return closePullRequestCommand(repository, number)
	case pullRequestLifecycleMutationReopen:
		return reopenPullRequestCommand(repository, number)
	default:
		return ""
	}
}

func pullRequestAutoMergeMutationCommand(kind pullRequestAutoMergeMutationKind, repository string, number int) string {
	switch kind {
	case pullRequestAutoMergeMutationEnable:
		return enablePullRequestAutoMergeCommand(repository, number)
	case pullRequestAutoMergeMutationDisable:
		return disablePullRequestAutoMergeCommand(repository, number)
	default:
		return ""
	}
}

func pullRequestMergeQueueMutationCommand(kind pullRequestMergeQueueMutationKind) string {
	switch kind {
	case pullRequestMergeQueueMutationEnqueue, pullRequestMergeQueueMutationDequeue:
		return formatStatusLineCommand("gh", "api", "graphql")
	default:
		return ""
	}
}
