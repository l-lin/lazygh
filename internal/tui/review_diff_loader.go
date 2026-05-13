package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestDiffResult struct {
	data                    reviewDiffData
	err                     error
	sourceUpdatedAt         string
	needsRefresh            bool
	fileTeamOwnersAttempted bool
}

func (program *Program) maybeLoadSelectedPullRequestDiff(gui *gocui.Gui) {
	program.executeWorkflowCommands(gui, program.reviewStore.planSelectedPullRequestDiffLoad(program, gui))
}

func (program *Program) loadPullRequestDiff(gui *gocui.Gui, summary any) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	repository := pullRequestRepositoryName(summaryValue.Repository)
	rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, summaryValue.Number)
	key := pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)
	result := pullRequestDiffResult{err: err, sourceUpdatedAt: pullRequestSummaryVersion(summaryValue)}
	if err == nil {
		rawDiff = program.withPullRequestDiffFileTeamOwners(repository, summaryValue.Number, rawDiff)
		result.data = buildReviewDiffData(rawDiff)
		result.needsRefresh = false
		result.fileTeamOwnersAttempted = rawDiff.FileTeamOwnersAttempted
		program.cachePullRequestDiff(summaryValue, rawDiff)
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
		cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(summaryValue)
		cachedResult.needsRefresh = false
		cachedResult.fileTeamOwnersAttempted = cachedResult.fileTeamOwnersAttempted || program.shouldLoadPullRequestDiffTeamOwners()
		program.pullRequestDiffCache[key] = cachedResult
		program.invalidatePullRequestDetailDocumentCache()
		return program.refreshViews(gui)
	})
}

func (program *Program) selectedPullRequestSummaryForDiff() (githubdomain.PullRequest, bool) {
	if program.reviewModeActive() {
		summary := program.reviewSession.summary
		if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubdomain.PullRequest{}, false
		}
		return summary, true
	}
	if !program.shouldShowPullRequestDetailTabs() || program.activeDetailTab != ChangesDetailTab {
		return githubdomain.PullRequest{}, false
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok || pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return githubdomain.PullRequest{}, false
	}
	return summary, true
}

func (program *Program) pullRequestDiffForSummary(summary any) (pullRequestDiffResult, bool) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return pullRequestDiffResult{}, false
	}
	result, ok := program.pullRequestDiffCache[pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)]
	return result, ok
}

func (program *Program) invalidatePullRequestDiff(repository string, number int) {
	delete(program.pullRequestDiffCache, strings.TrimSpace(repository)+fmt.Sprintf("#%d", number))
	program.invalidatePersistentPullRequest(repository, number)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}
