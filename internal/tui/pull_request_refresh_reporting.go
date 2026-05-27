package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

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
	if program == nil || pendingOperations <= 0 {
		return
	}

	program.feedbackMessage = ""
	program.manualRefreshState.feedback = &manualRefreshFeedbackState{
		successMessage:    strings.TrimSpace(successMessage),
		pendingOperations: pendingOperations,
	}
}

func (program *Program) completeManualRefreshOperation(err error) manualRefreshFeedbackCompletion {
	if program == nil || program.manualRefreshState.feedback == nil {
		return manualRefreshFeedbackCompletion{}
	}

	state := program.manualRefreshState.feedback
	completion := manualRefreshFeedbackCompletion{}
	if err != nil {
		if !state.failed {
			program.feedbackMessage = ""
			completion.popupError = strings.TrimSpace(normalizeGHCommandError(err).Error())
		}
		state.failed = true
	}
	if state.pendingOperations > 0 {
		state.pendingOperations--
	}
	if state.pendingOperations > 0 {
		return completion
	}
	if !state.failed && state.successMessage != "" {
		completion.successMessage = state.successMessage
	}
	program.manualRefreshState.feedback = nil
	return completion
}

func (program *Program) markManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil {
		return false
	}
	if program.manualRefreshState.pullRequestListPending == nil {
		program.manualRefreshState.pullRequestListPending = map[PullRequestTab]bool{}
	}
	program.manualRefreshState.pullRequestListPending[tab] = true
	return true
}

func (program *Program) consumeManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil || program.manualRefreshState.pullRequestListPending == nil {
		return false
	}
	pending := program.manualRefreshState.pullRequestListPending[tab]
	delete(program.manualRefreshState.pullRequestListPending, tab)
	return pending
}

func (program *Program) markManualPullRequestDetailRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}
	if program.manualRefreshState.pullRequestDetailPending == nil {
		program.manualRefreshState.pullRequestDetailPending = map[string]bool{}
	}
	if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
		program.manualRefreshState.pullRequestDetailPending[key] = true
		return true
	}
	return false
}

func (program *Program) consumeManualPullRequestDetailRefresh(key string) bool {
	if program == nil || program.manualRefreshState.pullRequestDetailPending == nil || key == "" {
		return false
	}
	pending := program.manualRefreshState.pullRequestDetailPending[key]
	delete(program.manualRefreshState.pullRequestDetailPending, key)
	return pending
}

func (program *Program) markManualPullRequestDiffRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}
	if program.manualRefreshState.pullRequestDiffPending == nil {
		program.manualRefreshState.pullRequestDiffPending = map[string]bool{}
	}
	if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
		program.manualRefreshState.pullRequestDiffPending[key] = true
		return true
	}
	return false
}

func (program *Program) consumeManualPullRequestDiffRefresh(key string) bool {
	if program == nil || program.manualRefreshState.pullRequestDiffPending == nil || key == "" {
		return false
	}
	pending := program.manualRefreshState.pullRequestDiffPending[key]
	delete(program.manualRefreshState.pullRequestDiffPending, key)
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
