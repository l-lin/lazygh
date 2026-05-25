package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) maybeLoadSelectedPullRequestDetail(gui *gocui.Gui) {
	program.executeCmds(gui, program.detailStore.planSelectedPullRequestDetailLoad(program, gui))
}

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

func (program *Program) currentDetailIdentity() string {
	actionContext := program.actionContext()
	if actionContext.IsReviewContext() {
		return program.reviewSessionDetailIdentity()
	}

	switch actionContext.MainView.ContentKind {
	case MainContentKindPullRequestDetail:
		if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
			if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
				if actionContext.MainView.SourceView.Focus == FocusNotificationsView {
					return fmt.Sprintf("notification-pr:%s:tab:%d", key, program.detailState.activeTab)
				}
				return fmt.Sprintf("pr:%s:tab:%d", key, program.detailState.activeTab)
			}
		}
		return fmt.Sprintf("pr-state:%d:%d", program.model.ActivePullRequestTab(), program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()))
	case MainContentKindNotificationDetail:
		if notification, ok := program.model.SelectedNotification(); ok {
			return fmt.Sprintf("notification:%s", strings.TrimSpace(notification.ID))
		}
		return fmt.Sprintf("notification-state:%d", program.model.SelectedNotificationIndex())
	default:
		return fmt.Sprintf("user:%d", program.model.SelectedUserIndex())
	}
}

func (program *Program) invalidatePullRequestDetail(repository string, number int) {
	delete(program.pullRequestDetailCache, strings.TrimSpace(repository)+fmt.Sprintf("#%d", number))
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
