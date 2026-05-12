package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

type pullRequestDiffResult struct {
	data                    reviewDiffData
	err                     error
	sourceUpdatedAt         string
	needsRefresh            bool
	fileTeamOwnersAttempted bool
}

func (program *Program) maybeLoadSelectedPullRequestDiff(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	summary, ok := program.selectedPullRequestSummaryForDiff()
	if !ok {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" || program.pullRequestDiffLoadInFlight[key] {
		return
	}

	program.hydratePullRequestDiffFromCache(summary)
	cachedResult, cached := program.pullRequestDiffForSummary(summary)
	if !program.pullRequestDiffNeedsRefresh(summary, cachedResult, cached) || !program.hasDetailQueries() {
		return
	}

	program.pullRequestDiffLoadInFlight[key] = true
	program.asyncRunner.Go(func() {
		program.loadPullRequestDiff(gui, summary)
	})
}

func (program *Program) loadPullRequestDiff(gui *gocui.Gui, summary githubcli.PullRequest) {
	repository := pullRequestRepositoryName(summary.Repository)
	rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, summary.Number)
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	result := pullRequestDiffResult{err: err, sourceUpdatedAt: pullRequestSummaryVersion(summary)}
	if err == nil {
		rawDiff = program.withPullRequestDiffFileTeamOwners(repository, summary.Number, rawDiff)
		result.data = buildReviewDiffData(rawDiff)
		result.needsRefresh = false
		result.fileTeamOwnersAttempted = rawDiff.FileTeamOwnersAttempted
		program.cachePullRequestDiff(summary, rawDiff)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.pullRequestDiffLoadInFlight, key)
		if err == nil || !program.canKeepPullRequestDiffOnRefreshError(key) {
			program.pullRequestDiffCache[key] = result
			program.invalidateReviewDiffRenderCache()
			program.invalidatePullRequestDetailDocumentCache()
			program.clampReviewSessionSelection()
			return program.refreshViews(gui)
		}

		cachedResult := program.pullRequestDiffCache[key]
		cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(summary)
		cachedResult.needsRefresh = false
		cachedResult.fileTeamOwnersAttempted = cachedResult.fileTeamOwnersAttempted || program.shouldLoadPullRequestDiffTeamOwners()
		program.pullRequestDiffCache[key] = cachedResult
		program.invalidatePullRequestDetailDocumentCache()
		return program.refreshViews(gui)
	})
}

func (program *Program) selectedPullRequestSummaryForDiff() (githubcli.PullRequest, bool) {
	if program.reviewSession.active {
		summary := program.reviewSession.summary
		if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubcli.PullRequest{}, false
		}
		return summary, true
	}
	if !program.shouldShowPullRequestDetailTabs() || program.activeDetailTab != ChangesDetailTab {
		return githubcli.PullRequest{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok || pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return githubcli.PullRequest{}, false
	}
	return summary, true
}

func (program *Program) pullRequestDiffForSummary(summary githubcli.PullRequest) (pullRequestDiffResult, bool) {
	result, ok := program.pullRequestDiffCache[pullRequestDetailKey(summary.Repository, summary.Number)]
	return result, ok
}

func (program *Program) invalidatePullRequestDiff(repository string, number int) {
	delete(program.pullRequestDiffCache, strings.TrimSpace(repository)+fmt.Sprintf("#%d", number))
	program.invalidatePersistentPullRequest(repository, number)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) selectedPullRequestDiffLoading() bool {
	summary, ok := program.selectedPullRequestSummaryForDiff()
	if !ok {
		return false
	}
	return program.pullRequestDiffLoadInFlight[pullRequestDetailKey(summary.Repository, summary.Number)]
}
