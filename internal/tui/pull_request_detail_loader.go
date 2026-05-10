package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func (program *Program) maybeLoadSelectedPullRequestDetail(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" || program.pullRequestDetailLoadInFlight[key] {
		return
	}

	program.hydratePullRequestDetailFromCache(summary)
	cachedResult, cached := program.pullRequestDetailForSummary(summary)
	if !program.pullRequestDetailNeedsRefresh(summary, cachedResult, cached) || program.githubLoader == nil {
		return
	}

	program.pullRequestDetailLoadInFlight[key] = true
	program.asyncRunner.Go(func() {
		program.loadPullRequestDetail(gui, summary)
	})
}

func (program *Program) loadPullRequestDetail(gui *gocui.Gui, summary githubcli.PullRequest) {
	repository := pullRequestRepositoryName(summary.Repository)
	detail, err := program.githubLoader.GetPullRequestDetail(repository, summary.Number)
	pendingReviewState := pendingPullRequestReviewState{}
	pendingReviewStateKnown := false
	if pendingReviewID, found, pendingReviewErr := program.githubLoader.GetPendingPullRequestReviewID(repository, summary.Number); pendingReviewErr == nil {
		pendingReviewStateKnown = true
		if found {
			pendingReviewState.id = strings.TrimSpace(pendingReviewID)
		}
	}
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	result := pullRequestDetailResult{err: err, sourceUpdatedAt: pullRequestSummaryVersion(summary)}
	if err == nil {
		clonedDetail := clonePullRequestDetail(detail)
		result.detail = clonedDetail
		result.needsRefresh = false
		program.cachePullRequestDetail(summary, clonedDetail)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.pullRequestDetailLoadInFlight, key)
		if pendingReviewStateKnown {
			program.pendingPullRequestReviewCache[key] = pendingReviewState
		}
		if err == nil || !program.canKeepPullRequestDetailOnRefreshError(key) {
			program.pullRequestDetailCache[key] = result
			program.invalidatePullRequestDetailDocumentCache()
			return program.refreshViews(gui)
		}

		cachedResult := program.pullRequestDetailCache[key]
		cachedResult.sourceUpdatedAt = pullRequestSummaryVersion(summary)
		cachedResult.needsRefresh = false
		program.pullRequestDetailCache[key] = cachedResult
		return program.refreshViews(gui)
	})
}

func (program *Program) selectedPullRequestSummaryForDetail() (githubcli.PullRequest, bool) {
	if program.reviewSession.active {
		summary := program.reviewSession.summary
		if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubcli.PullRequest{}, false
		}
		return summary, true
	}
	switch program.model.currentSideFocus() {
	case FocusPullRequestsView:
		summary, ok := program.model.SelectedPullRequestSummary()
		if !ok {
			return githubcli.PullRequest{}, false
		}
		if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubcli.PullRequest{}, false
		}
		return summary, true
	case FocusNotificationsView:
		notification, ok := program.model.SelectedNotification()
		if !ok {
			return githubcli.PullRequest{}, false
		}
		summary, ok := notification.PullRequestSummary()
		if !ok || pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubcli.PullRequest{}, false
		}
		return summary, true
	default:
		return githubcli.PullRequest{}, false
	}
}

func (program *Program) pullRequestDetailForSummary(summary githubcli.PullRequest) (pullRequestDetailResult, bool) {
	result, ok := program.pullRequestDetailCache[pullRequestDetailKey(summary.Repository, summary.Number)]
	return result, ok
}

func (program *Program) currentDetailIdentity() string {
	if program.reviewSession.active {
		return program.reviewSessionDetailIdentity()
	}

	switch program.model.currentSideFocus() {
	case FocusPullRequestsView:
		if summary, ok := program.model.SelectedPullRequestSummary(); ok {
			if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
				return fmt.Sprintf("pr:%s:tab:%d", key, program.activeDetailTab)
			}
		}
		return fmt.Sprintf("pr-state:%d:%d", program.model.ActivePullRequestTab(), program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()))
	case FocusNotificationsView:
		if summary, ok := program.selectedPullRequestSummaryForDetail(); ok {
			if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
				return fmt.Sprintf("notification-pr:%s:tab:%d", key, program.activeDetailTab)
			}
		}
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

func pullRequestDetailKey(repository githubcli.Repository, number int) string {
	repositoryName := strings.TrimSpace(pullRequestRepositoryName(repository))
	if repositoryName == "" || repositoryName == "-" || number <= 0 {
		return ""
	}

	return fmt.Sprintf("%s#%d", repositoryName, number)
}
