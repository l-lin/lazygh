package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

var pastedPullRequestsPersistentCacheCommand = []string{"lazygh", "pasted-pull-requests"}

func pastedPullRequestsPersistentSearch() appconfig.PullRequestSearch {
	return appconfig.PullRequestSearch{
		Label:   pastedPullRequestsTabLabel,
		Command: append([]string(nil), pastedPullRequestsPersistentCacheCommand...),
	}
}

func loadPastedPullRequestsFromPersistentCache(cache persistentPullRequestCache) []githubdomain.PullRequest {
	if cache == nil {
		return nil
	}

	pullRequests, ok, actualErr := cache.PullRequests(pastedPullRequestsPersistentSearch())
	if actualErr != nil || !ok {
		return nil
	}
	return clonePersistentCachePullRequests(pullRequests)
}

func (program *Program) queueSavePastedPullRequestsPersistentCache() {
	if program == nil || program.pullRequestCache == nil {
		return
	}

	program.queuePersistentCacheShellAction(savePullRequestsPersistentCacheAction{
		search:       pastedPullRequestsPersistentSearch(),
		pullRequests: clonePersistentCachePullRequests(program.pastedPullRequests.summaries()),
	})
}
