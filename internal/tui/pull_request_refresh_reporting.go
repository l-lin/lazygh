package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type manualRefreshFeedbackState struct {
	successMessage    string
	pendingOperations int
	failed            bool
}

type manualRefreshFeedbackCompletion struct {
	popupError     string
	successMessage string
}

func (program *Program) beginManualRefresh(successMessage string, pendingOperations int) {
	if program == nil {
		return
	}

	started := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualStarted := state.withFeedbackBegun(successMessage, pendingOperations)
		if !actualStarted {
			return state
		}
		started = true
		return updatedState
	})
	if started {
		program.clearFeedbackMessage()
	}
}

func (program *Program) completeManualRefreshOperation(err error) manualRefreshFeedbackCompletion {
	if program == nil {
		return manualRefreshFeedbackCompletion{}
	}

	completion := manualRefreshFeedbackCompletion{}
	clearFeedback := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualCompletion, actualClearFeedback := state.withCompletedOperation(err)
		completion = actualCompletion
		clearFeedback = actualClearFeedback
		return updatedState
	})
	if clearFeedback {
		program.clearFeedbackMessage()
	}
	return completion
}

func (program *Program) markManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil {
		return false
	}

	marked := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualMarked := state.withPullRequestListPending(tab)
		if !actualMarked {
			return state
		}
		marked = true
		return updatedState
	})
	return marked
}

func (program *Program) consumeManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil {
		return false
	}

	pending := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualPending := state.withoutPullRequestListPending(tab)
		if !actualPending {
			return state
		}
		pending = true
		return updatedState
	})
	return pending
}

func (program *Program) markManualPullRequestDetailRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}

	marked := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualMarked := state.withPullRequestDetailPending(summary)
		if !actualMarked {
			return state
		}
		marked = true
		return updatedState
	})
	return marked
}

func (program *Program) consumeManualPullRequestDetailRefresh(key string) bool {
	if program == nil {
		return false
	}

	pending := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualPending := state.withoutPullRequestDetailPending(key)
		if !actualPending {
			return state
		}
		pending = true
		return updatedState
	})
	return pending
}

func (program *Program) markManualPullRequestDiffRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}

	marked := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualMarked := state.withPullRequestDiffPending(summary)
		if !actualMarked {
			return state
		}
		marked = true
		return updatedState
	})
	return marked
}

func (program *Program) consumeManualPullRequestDiffRefresh(key string) bool {
	if program == nil {
		return false
	}

	pending := false
	program.updateManualRefreshState(func(state manualRefreshStateModel) manualRefreshStateModel {
		updatedState, actualPending := state.withoutPullRequestDiffPending(key)
		if !actualPending {
			return state
		}
		pending = true
		return updatedState
	})
	return pending
}

func (program *Program) markPullRequestDetailNeedsRefresh(summary githubdomain.PullRequest) {
	if program == nil {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: program.optimisticPullRequestDetailSeed(summary)}
	}
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	result.err = nil
	program.applyPullRequestDetailCacheResult(pullRequestRepositoryName(summary.Repository), summary.Number, result, pullRequestDetailCacheApplyOptions{clearInFlight: true, invalidateDocuments: true})
}

func (program *Program) markPullRequestDiffNeedsRefresh(summary githubdomain.PullRequest) {
	if program == nil {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	result, ok := program.pullRequestDiffCache[key]
	if !ok || result.err != nil {
		return
	}
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	result.err = nil
	program.applyPullRequestDiffCacheResult(pullRequestRepositoryName(summary.Repository), summary.Number, result, pullRequestDiffCacheApplyOptions{clearInFlight: true, invalidateReviewRender: true, invalidateDetailDocs: true})
}
