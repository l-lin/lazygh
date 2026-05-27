package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) beginManualPullRequestListRefresh(tab PullRequestTab, successMessage string) {
	if program == nil {
		return
	}

	pendingOperations := 0
	if program.hasPullRequestListQueries() && program.markManualPullRequestListRefresh(tab) {
		pendingOperations++
	}
	program.beginManualRefresh(successMessage, pendingOperations)
}

func (program *Program) beginManualPullRequestRefresh(summary githubdomain.PullRequest, tab PullRequestTab) {
	if program == nil {
		return
	}

	pendingOperations := 0
	if program.hasDetailQueries() && program.markManualPullRequestDetailRefresh(summary) {
		pendingOperations++
	}
	if program.reviewModeActive() {
		if program.hasDetailQueries() && program.markManualPullRequestDiffRefresh(summary) {
			pendingOperations++
		}
	} else if program.hasPullRequestListQueries() && program.markManualPullRequestListRefresh(tab) {
		pendingOperations++
	}
	program.beginManualRefresh(pullRequestRefreshSuccessMessage, pendingOperations)
}

func (program *Program) beginManualNotificationsRefresh() {
	if program == nil {
		return
	}

	pendingOperations := 0
	if !program.reviewModeActive() && program.hasNotificationQueries() && program.markManualNotificationRefresh() {
		pendingOperations++
	}
	program.beginManualRefresh(notificationsRefreshSuccessMessage, pendingOperations)
}
