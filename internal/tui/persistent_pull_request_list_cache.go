package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func (program *Program) pullRequestsFromCache(tab PullRequestTab) ([]githubdomain.PullRequest, bool) {
	if program.pullRequestCache == nil || !program.canHydratePullRequestsFromCache(tab) {
		return nil, false
	}

	search, searchBacked := program.searchBackedPullRequestSearch(tab)
	if !searchBacked {
		return nil, false
	}

	pullRequests, ok, actualErr := program.pullRequestCache.PullRequests(search)
	if actualErr != nil || !ok {
		return nil, false
	}
	return pullRequests, true
}

func (program *Program) canHydratePullRequestsFromCache(tab PullRequestTab) bool {
	rows := program.model.PullRequestRows(tab)
	if len(rows) == 0 {
		return true
	}

	return len(rows) == 1 && program.isPullRequestLoadingItem(rows[0].Item)
}

func (program *Program) cachePullRequests(tab PullRequestTab, pullRequests []githubdomain.PullRequest) {
	search, ok := program.searchBackedPullRequestSearch(tab)
	if !ok {
		return
	}
	program.queuePersistentCacheShellAction(savePullRequestsPersistentCacheAction{search: search, pullRequests: clonePersistentCachePullRequests(pullRequests)})
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
