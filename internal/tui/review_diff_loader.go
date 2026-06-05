package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestDiffResult struct {
	data                    reviewDiffData
	err                     error
	sourceUpdatedAt         string
	needsRefresh            bool
	fileTeamOwnersAttempted bool
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
	key := strings.TrimSpace(repository) + fmt.Sprintf("#%d", number)
	program.updateReviewStore(func(store reviewStore) reviewStore {
		return store.withoutPullRequestDiff(key)
	})
	program.invalidatePullRequestStoryReview(repository, number)
	program.invalidatePersistentPullRequest(repository, number)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}
