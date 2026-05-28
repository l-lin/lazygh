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

	updatedState, started := program.manualRefreshState.withFeedbackBegun(successMessage, pendingOperations)
	if !started {
		return
	}
	program.clearFeedbackMessage()
	program.manualRefreshState = updatedState
}

func (program *Program) completeManualRefreshOperation(err error) manualRefreshFeedbackCompletion {
	if program == nil {
		return manualRefreshFeedbackCompletion{}
	}

	updatedState, completion, clearFeedback := program.manualRefreshState.withCompletedOperation(err)
	if clearFeedback {
		program.clearFeedbackMessage()
	}
	program.manualRefreshState = updatedState
	return completion
}

func (program *Program) markManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil {
		return false
	}
	updatedState, marked := program.manualRefreshState.withPullRequestListPending(tab)
	if !marked {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}

func (program *Program) consumeManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil {
		return false
	}
	updatedState, pending := program.manualRefreshState.withoutPullRequestListPending(tab)
	if !pending {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}

func (program *Program) markManualPullRequestDetailRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}
	updatedState, marked := program.manualRefreshState.withPullRequestDetailPending(summary)
	if !marked {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}

func (program *Program) consumeManualPullRequestDetailRefresh(key string) bool {
	if program == nil {
		return false
	}
	updatedState, pending := program.manualRefreshState.withoutPullRequestDetailPending(key)
	if !pending {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}

func (program *Program) markManualPullRequestDiffRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}
	updatedState, marked := program.manualRefreshState.withPullRequestDiffPending(summary)
	if !marked {
		return false
	}
	program.manualRefreshState = updatedState
	return true
}

func (program *Program) consumeManualPullRequestDiffRefresh(key string) bool {
	if program == nil {
		return false
	}
	updatedState, pending := program.manualRefreshState.withoutPullRequestDiffPending(key)
	if !pending {
		return false
	}
	program.manualRefreshState = updatedState
	return true
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
