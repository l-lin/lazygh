package tui

import "codeberg.org/l-lin/lazygh/internal/githubcli"

func (program *Program) hydratePullRequestsFromCache(tab PullRequestTab) bool {
	if program.pullRequestCache == nil || !program.canHydratePullRequestsFromCache(tab) {
		return false
	}

	pullRequests, ok, actualErr := program.pullRequestCache.PullRequests(program.pullRequestSearch(tab))
	if actualErr != nil || !ok {
		return false
	}

	program.setPullRequestsCount(tab, len(pullRequests), true)
	program.model.SetPullRequestRows(tab, program.pullRequestRowsForTab(tab, pullRequests, nil))
	return true
}

func (program *Program) canHydratePullRequestsFromCache(tab PullRequestTab) bool {
	rows := program.model.PullRequestRows(tab)
	if len(rows) == 0 {
		return true
	}

	return len(rows) == 1 && program.isPullRequestLoadingItem(rows[0].Item)
}

func (program *Program) cachePullRequests(tab PullRequestTab, pullRequests []githubcli.PullRequest) {
	if program.pullRequestCache == nil {
		return
	}

	_ = program.pullRequestCache.SavePullRequests(program.pullRequestSearch(tab), pullRequests)
}

func (program *Program) shouldPreservePullRequestRowsOnRefreshError(tab PullRequestTab) bool {
	rows := program.model.PullRequestRows(tab)
	if len(rows) == 0 {
		return false
	}
	if len(rows) > 1 {
		return true
	}

	item := rows[0].Item
	return !program.isPullRequestLoadingItem(item) && !program.isPullRequestErrorItem(tab, item)
}

func (program *Program) isPullRequestErrorItem(tab PullRequestTab, item Item) bool {
	state := program.pullRequestListState(tab)
	title := item.Title
	switch title {
	case state.unauthenticatedTitle, state.unavailableTitle, state.genericErrorTitle:
		return true
	default:
		return false
	}
}
