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
	program.executeCmds(gui, program.reviewStore.planSelectedPullRequestDiffLoad(program, gui))
}

func (program *Program) loadPullRequestDiff(gui *gocui.Gui, summary any) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return
	}
	repository := pullRequestRepositoryName(summaryValue.Repository)
	rawDiff, err := program.detailQueries.GetPullRequestDiff(repository, summaryValue.Number)
	if err == nil {
		rawDiff = program.withPullRequestDiffFileTeamOwners(repository, summaryValue.Number, rawDiff)
	}
	program.dispatchAsync(gui, MsgPullRequestDiffLoaded{Summary: summaryValue, RawDiff: rawDiff, Err: err})
}

func (program *Program) selectedPullRequestSummaryForDiff() (githubdomain.PullRequest, bool) {
	if program.reviewModeActive() {
		summary := program.navigationState.reviewSession.summary
		if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubdomain.PullRequest{}, false
		}
		return summary, true
	}
	if !program.shouldShowPullRequestDetailTabs() || program.detailState.activeTab != ChangesDetailTab {
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
