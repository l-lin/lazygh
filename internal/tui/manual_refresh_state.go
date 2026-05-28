package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (state manualRefreshStateModel) withPullRequestListPending(tab PullRequestTab) (manualRefreshStateModel, bool) {
	state.pullRequestListPending = copyManualRefreshPendingTabs(state.pullRequestListPending)
	state.pullRequestListPending[tab] = true
	return state, true
}

func (state manualRefreshStateModel) withoutPullRequestListPending(tab PullRequestTab) (manualRefreshStateModel, bool) {
	if !state.pullRequestListPending[tab] {
		return state, false
	}
	state.pullRequestListPending = copyManualRefreshPendingTabs(state.pullRequestListPending)
	delete(state.pullRequestListPending, tab)
	return state, true
}

func (state manualRefreshStateModel) withPullRequestDetailPending(summary githubdomain.PullRequest) (manualRefreshStateModel, bool) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return state, false
	}
	state.pullRequestDetailPending = copyManualRefreshPendingKeys(state.pullRequestDetailPending)
	state.pullRequestDetailPending[key] = true
	return state, true
}

func (state manualRefreshStateModel) withoutPullRequestDetailPending(key string) (manualRefreshStateModel, bool) {
	if key == "" || !state.pullRequestDetailPending[key] {
		return state, false
	}
	state.pullRequestDetailPending = copyManualRefreshPendingKeys(state.pullRequestDetailPending)
	delete(state.pullRequestDetailPending, key)
	return state, true
}

func (state manualRefreshStateModel) withPullRequestDiffPending(summary githubdomain.PullRequest) (manualRefreshStateModel, bool) {
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return state, false
	}
	state.pullRequestDiffPending = copyManualRefreshPendingKeys(state.pullRequestDiffPending)
	state.pullRequestDiffPending[key] = true
	return state, true
}

func (state manualRefreshStateModel) withoutPullRequestDiffPending(key string) (manualRefreshStateModel, bool) {
	if key == "" || !state.pullRequestDiffPending[key] {
		return state, false
	}
	state.pullRequestDiffPending = copyManualRefreshPendingKeys(state.pullRequestDiffPending)
	delete(state.pullRequestDiffPending, key)
	return state, true
}

func (state manualRefreshStateModel) withNotificationPending() (manualRefreshStateModel, bool) {
	state.notificationPending = true
	return state, true
}

func (state manualRefreshStateModel) withoutNotificationPending() (manualRefreshStateModel, bool) {
	if !state.notificationPending {
		return state, false
	}
	state.notificationPending = false
	return state, true
}

func (state manualRefreshStateModel) withFeedbackBegun(successMessage string, pendingOperations int) (manualRefreshStateModel, bool) {
	if pendingOperations <= 0 {
		return state, false
	}
	feedbackState := manualRefreshFeedbackState{successMessage: strings.TrimSpace(successMessage), pendingOperations: pendingOperations}
	state.feedback = &feedbackState
	return state, true
}

func (state manualRefreshStateModel) withCompletedOperation(err error) (manualRefreshStateModel, manualRefreshFeedbackCompletion, bool) {
	if state.feedback == nil {
		return state, manualRefreshFeedbackCompletion{}, false
	}

	updatedFeedback, completion, active, clearFeedback := state.feedback.completed(err)
	if active {
		state.feedback = &updatedFeedback
		return state, completion, clearFeedback
	}
	state.feedback = nil
	return state, completion, clearFeedback
}

func (state manualRefreshFeedbackState) completed(err error) (manualRefreshFeedbackState, manualRefreshFeedbackCompletion, bool, bool) {
	completion := manualRefreshFeedbackCompletion{}
	clearFeedback := false
	if err != nil {
		if !state.failed {
			clearFeedback = true
			completion.popupError = strings.TrimSpace(normalizeGHCommandError(err).Error())
		}
		state.failed = true
	}
	if state.pendingOperations > 0 {
		state.pendingOperations--
	}
	if state.pendingOperations > 0 {
		return state, completion, true, clearFeedback
	}
	if !state.failed && state.successMessage != "" {
		completion.successMessage = state.successMessage
	}
	return manualRefreshFeedbackState{}, completion, false, clearFeedback
}

func copyManualRefreshPendingTabs(source map[PullRequestTab]bool) map[PullRequestTab]bool {
	copied := make(map[PullRequestTab]bool, len(source))
	for tab, pending := range source {
		copied[tab] = pending
	}
	return copied
}

func copyManualRefreshPendingKeys(source map[string]bool) map[string]bool {
	copied := make(map[string]bool, len(source))
	for key, pending := range source {
		copied[key] = pending
	}
	return copied
}
