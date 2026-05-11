package tui

import (
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) hydratePullRequestDetailFromCache(summary githubcli.PullRequest) bool {
	if program.pullRequestCache == nil {
		return false
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return false
	}
	if _, ok := program.pullRequestDetailCache[key]; ok {
		return false
	}

	cached, ok, actualErr := program.pullRequestCache.PullRequestDetail(pullRequestRepositoryName(summary.Repository), summary.Number)
	if actualErr != nil || !ok {
		return false
	}

	clonedDetail := clonePullRequestDetail(cached.Detail)
	program.pullRequestDetailCache[key] = pullRequestDetailResult{
		detail:          clonedDetail,
		sourceUpdatedAt: strings.TrimSpace(cached.SourceUpdatedAt),
		needsRefresh:    cachedPullRequestNeedsRefresh(summary, cached.SourceUpdatedAt) || pullRequestDetailMissingBrowserTabData(clonedDetail),
	}
	program.invalidatePullRequestDetailDocumentCache()
	return true
}

func (program *Program) pullRequestDetailNeedsRefresh(summary githubcli.PullRequest, result pullRequestDetailResult, ok bool) bool {
	if !ok || result.err != nil {
		return true
	}

	currentVersion := pullRequestSummaryVersion(summary)
	if currentVersion == "" {
		return result.needsRefresh
	}

	return result.needsRefresh || strings.TrimSpace(result.sourceUpdatedAt) != currentVersion
}

func (program *Program) canKeepPullRequestDetailOnRefreshError(key string) bool {
	result, ok := program.pullRequestDetailCache[key]
	return ok && result.err == nil
}

func (program *Program) cachePullRequestDetail(summary githubcli.PullRequest, detail githubcli.PullRequestDetail) {
	if program.pullRequestCache == nil {
		return
	}

	_ = program.pullRequestCache.SavePullRequestDetail(summary, detail)
}

func (program *Program) hydratePullRequestDiffFromCache(summary githubcli.PullRequest) bool {
	if program.pullRequestCache == nil {
		return false
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return false
	}
	if _, ok := program.pullRequestDiffCache[key]; ok {
		return false
	}

	cached, ok, actualErr := program.pullRequestCache.PullRequestDiff(pullRequestRepositoryName(summary.Repository), summary.Number)
	if actualErr != nil || !ok {
		return false
	}

	program.pullRequestDiffCache[key] = pullRequestDiffResult{
		data:                    buildReviewDiffData(cached.Diff),
		sourceUpdatedAt:         strings.TrimSpace(cached.SourceUpdatedAt),
		needsRefresh:            cachedPullRequestNeedsRefresh(summary, cached.SourceUpdatedAt),
		fileTeamOwnersAttempted: cached.Diff.FileTeamOwnersAttempted,
	}
	program.invalidateReviewDiffRenderCache()
	program.clampReviewSessionSelection()
	return true
}

func (program *Program) pullRequestDiffNeedsRefresh(summary githubcli.PullRequest, result pullRequestDiffResult, ok bool) bool {
	if !ok || result.err != nil {
		return true
	}

	if program.shouldLoadPullRequestDiffTeamOwners() && !result.fileTeamOwnersAttempted {
		return true
	}

	currentVersion := pullRequestSummaryVersion(summary)
	if currentVersion == "" {
		return result.needsRefresh
	}

	return result.needsRefresh || strings.TrimSpace(result.sourceUpdatedAt) != currentVersion
}

func (program *Program) canKeepPullRequestDiffOnRefreshError(key string) bool {
	result, ok := program.pullRequestDiffCache[key]
	return ok && result.err == nil
}

func (program *Program) cachePullRequestDiff(summary githubcli.PullRequest, diff githubcli.PullRequestDiff) {
	if program.pullRequestCache == nil {
		return
	}

	_ = program.pullRequestCache.SavePullRequestDiff(summary, diff)
}

func (program *Program) invalidatePersistentPullRequest(repository string, number int) {
	if program.pullRequestCache == nil {
		return
	}

	_ = program.pullRequestCache.InvalidatePullRequest(strings.TrimSpace(repository), number)
}

func pullRequestSummaryVersion(summary githubcli.PullRequest) string {
	return strings.TrimSpace(summary.UpdatedAt)
}

func cachedPullRequestNeedsRefresh(summary githubcli.PullRequest, cachedSourceUpdatedAt string) bool {
	currentVersion := pullRequestSummaryVersion(summary)
	if currentVersion == "" {
		return true
	}

	return strings.TrimSpace(cachedSourceUpdatedAt) != currentVersion
}

func pullRequestDetailMissingBrowserTabData(detail githubcli.PullRequestDetail) bool {
	if len(detail.Commits) == 0 {
		return true
	}

	for _, check := range detail.StatusCheckRollup {
		if pullRequestOverviewStatusForCheck(check) == pullRequestOverviewStatusPending {
			continue
		}
		if strings.TrimSpace(check.Link) == "" {
			return true
		}
	}
	return false
}
