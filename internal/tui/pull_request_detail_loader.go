package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) selectedPullRequestSummaryForDetail() (githubdomain.PullRequest, bool) {
	actionContext := program.actionContext()
	if actionContext.IsReviewContext() {
		summary, ok := toDomainPullRequestSummary(program.navigationState.reviewSession.summary)
		if !ok || pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubdomain.PullRequest{}, false
		}
		return summary, true
	}
	if actionContext.MainView.ContentKind != MainContentKindPullRequestDetail {
		return githubdomain.PullRequest{}, false
	}

	switch actionContext.MainView.SourceView.Focus {
	case FocusPullRequestsView:
		summary, ok := program.model.SelectedPullRequestSummary()
		if !ok || pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubdomain.PullRequest{}, false
		}
		return summary, true
	case FocusNotificationsView:
		notification, ok := program.model.SelectedNotification()
		if !ok {
			return githubdomain.PullRequest{}, false
		}
		summary, ok := notification.PullRequestSummary()
		if !ok || pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubdomain.PullRequest{}, false
		}
		return summary, true
	default:
		return githubdomain.PullRequest{}, false
	}
}

func (program *Program) pullRequestDetailForSummary(summary any) (pullRequestDetailResult, bool) {
	summaryValue, ok := toDomainPullRequestSummary(summary)
	if !ok {
		return pullRequestDetailResult{}, false
	}
	result, ok := program.pullRequestDetailCache[pullRequestDetailKey(summaryValue.Repository, summaryValue.Number)]
	return result, ok
}

func (program *Program) invalidatePullRequestDetail(repository string, number int) {
	key := strings.TrimSpace(repository) + fmt.Sprintf("#%d", number)
	program.updateDetailStore(func(store detailStore) detailStore {
		return store.withoutPullRequestDetail(key)
	})
	program.forgetPendingPullRequestReviewState(repository, number)
	program.invalidatePersistentPullRequest(repository, number)
	program.invalidatePullRequestDetailDocumentCache()
}

func pullRequestDetailKey(repository any, number int) string {
	repositoryName := strings.TrimSpace(pullRequestRepositoryName(repository))
	if repositoryName == "" || repositoryName == "-" || number <= 0 {
		return ""
	}

	return fmt.Sprintf("%s#%d", repositoryName, number)
}
