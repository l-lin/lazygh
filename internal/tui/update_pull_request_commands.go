package tui

import appconfig "github.com/l-lin/lazygh/internal/config"

func (program *Program) applyPullRequestSearchesApplied(message MsgPullRequestSearchesApplied) {
	program.runtimeConfig.pullRequestSearches = append([]appconfig.PullRequestSearch(nil), message.Searches...)
	program.resetPullRequestSearchState()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestSearches))
}
