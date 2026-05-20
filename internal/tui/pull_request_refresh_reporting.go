package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) markManualPullRequestListRefresh(tab PullRequestTab) {
	if program == nil {
		return
	}
	if program.manualPullRequestListRefreshErrors == nil {
		program.manualPullRequestListRefreshErrors = map[PullRequestTab]bool{}
	}
	program.manualPullRequestListRefreshErrors[tab] = true
}

func (program *Program) consumeManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil || program.manualPullRequestListRefreshErrors == nil {
		return false
	}
	pending := program.manualPullRequestListRefreshErrors[tab]
	delete(program.manualPullRequestListRefreshErrors, tab)
	return pending
}

func (program *Program) markManualPullRequestDetailRefresh(summary githubdomain.PullRequest) {
	if program == nil {
		return
	}
	if program.manualPullRequestDetailRefreshErrors == nil {
		program.manualPullRequestDetailRefreshErrors = map[string]bool{}
	}
	if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
		program.manualPullRequestDetailRefreshErrors[key] = true
	}
}

func (program *Program) consumeManualPullRequestDetailRefresh(key string) bool {
	if program == nil || program.manualPullRequestDetailRefreshErrors == nil || key == "" {
		return false
	}
	pending := program.manualPullRequestDetailRefreshErrors[key]
	delete(program.manualPullRequestDetailRefreshErrors, key)
	return pending
}

func (program *Program) markManualPullRequestDiffRefresh(summary githubdomain.PullRequest) {
	if program == nil {
		return
	}
	if program.manualPullRequestDiffRefreshErrors == nil {
		program.manualPullRequestDiffRefreshErrors = map[string]bool{}
	}
	if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
		program.manualPullRequestDiffRefreshErrors[key] = true
	}
}

func (program *Program) consumeManualPullRequestDiffRefresh(key string) bool {
	if program == nil || program.manualPullRequestDiffRefreshErrors == nil || key == "" {
		return false
	}
	pending := program.manualPullRequestDiffRefreshErrors[key]
	delete(program.manualPullRequestDiffRefreshErrors, key)
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
	program.pullRequestDetailCache[key] = result
	delete(program.pullRequestDetailLoadInFlight, key)
	program.invalidatePullRequestDetailDocumentCache()
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
	program.pullRequestDiffCache[key] = result
	delete(program.pullRequestDiffLoadInFlight, key)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}
