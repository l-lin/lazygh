package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) applyOpenPullRequestInBrowserRequested(message MsgOpenPullRequestInBrowserRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	return []Cmd{actionsPopupAsyncWorkCmd{Command: openPullRequestInBrowserCommand(repository, number), Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.pullRequestMutations.OpenPullRequestInBrowser(repository, number); err != nil {
			return nil, err
		}
		return actionsPopupAsyncFeedbackSuccess{Message: pullRequestBrowserOpenSuccessMessage}, nil
	}}}
}

func (program *Program) applyApprovePullRequestRequested(message MsgApprovePullRequestRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	return []Cmd{actionsPopupAsyncWorkCmd{Command: approvePullRequestCommand(repository, number), Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.reviewMutations.ApprovePullRequest(repository, number); err != nil {
			return nil, err
		}
		return actionsPopupAsyncInvalidatePullRequestSuccess{
			Repository:     repository,
			Number:         number,
			InvalidateDiff: true,
			Message:        pullRequestReviewSuccessMessage,
		}, nil
	}}}
}

func (program *Program) applyReRequestPullRequestReviewRequested(message MsgReRequestPullRequestReviewRequested) []Cmd {
	repository, number, reviewerLogin, ok := popupPullRequestReviewerRequestIdentity(message.Target)
	if !ok {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	return []Cmd{actionsPopupAsyncWorkCmd{Command: requestPullRequestReviewerCommand(repository, number, reviewerLogin), Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.pullRequestMutations.RequestPullRequestReviewer(repository, number, reviewerLogin); err != nil {
			return nil, err
		}
		return actionsPopupAsyncInvalidatePullRequestSuccess{
			Repository: repository,
			Number:     number,
			Message:    pullRequestReviewReRequestedSuccessMessage,
		}, nil
	}}}
}

func (program *Program) applyPullRequestLifecycleMutationRequested(message MsgPullRequestLifecycleMutationRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	command := pullRequestLifecycleMutationCommand(message.Kind, repository, number)
	if command == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	return []Cmd{actionsPopupAsyncWorkCmd{Command: command, Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.runPullRequestLifecycleMutation(message.Kind, repository, number); err != nil {
			return nil, err
		}
		return actionsPopupAsyncPullRequestLifecycleSuccess{
			Summary: message.Summary,
			State:   message.State,
			IsDraft: message.IsDraft,
			Message: message.SuccessMessage,
		}, nil
	}}}
}

func (program *Program) applyPullRequestAutoMergeMutationRequested(message MsgPullRequestAutoMergeMutationRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	command := pullRequestAutoMergeMutationCommand(message.Kind, repository, number)
	if command == "" {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	return []Cmd{actionsPopupAsyncWorkCmd{Command: command, Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := program.runPullRequestAutoMergeMutation(message.Kind, repository, number); err != nil {
			return nil, err
		}
		return actionsPopupAsyncPullRequestAutoMergeSuccess{
			Summary: message.Summary,
			Enabled: message.Enabled,
			Message: message.SuccessMessage,
		}, nil
	}}}
}

func (program *Program) applyPullRequestBranchUpdateRequested(message MsgPullRequestBranchUpdateRequested) []Cmd {
	repository, number, ok := popupPullRequestActionTargetIdentity(message.Target)
	if !ok || !popupPullRequestSummaryValid(message.Summary) {
		program.actionsPopupWidget.errorMessage = errActionsPopupActionUnavailable.Error()
		return nil
	}

	return []Cmd{actionsPopupAsyncWorkCmd{Command: updatePullRequestBranchCommand(repository, number), Async: true, Work: func(program *Program) (actionsPopupAsyncSuccess, error) {
		if err := normalizedPullRequestMutationError(program.pullRequestMutations.UpdatePullRequestBranch(repository, number), "gh pr update-branch"); err != nil {
			return nil, err
		}
		return actionsPopupAsyncPullRequestBranchUpdateSuccess{Summary: message.Summary, Message: pullRequestBranchUpdatedSuccessMessage}, nil
	}}}
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

func (program *Program) runPullRequestLifecycleMutation(kind pullRequestLifecycleMutationKind, repository string, number int) error {
	switch kind {
	case pullRequestLifecycleMutationReadyForReview:
		return normalizedPullRequestMutationError(program.pullRequestMutations.MarkPullRequestReadyForReview(repository, number), "gh pr ready")
	case pullRequestLifecycleMutationConvertToDraft:
		return normalizedPullRequestMutationError(program.pullRequestMutations.ConvertPullRequestToDraft(repository, number), "gh pr ready")
	case pullRequestLifecycleMutationClose:
		return normalizedPullRequestMutationError(program.pullRequestMutations.ClosePullRequest(repository, number), "gh pr close")
	case pullRequestLifecycleMutationReopen:
		return normalizedPullRequestMutationError(program.pullRequestMutations.ReopenPullRequest(repository, number), "gh pr reopen")
	default:
		return errActionsPopupActionUnavailable
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

func (program *Program) runPullRequestAutoMergeMutation(kind pullRequestAutoMergeMutationKind, repository string, number int) error {
	switch kind {
	case pullRequestAutoMergeMutationEnable:
		return normalizedPullRequestMutationError(program.pullRequestMutations.EnablePullRequestAutoMerge(repository, number), "gh pr merge")
	case pullRequestAutoMergeMutationDisable:
		return normalizedPullRequestMutationError(program.pullRequestMutations.DisablePullRequestAutoMerge(repository, number), "gh pr merge")
	default:
		return errActionsPopupActionUnavailable
	}
}
