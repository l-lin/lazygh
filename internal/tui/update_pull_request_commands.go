package tui

func (program *Program) applyPullRequestSearchesApplied(message MsgPullRequestSearchesApplied) {
	program.setRuntimePullRequestSearches(message.Searches)
	program.resetPullRequestListLoadState()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestSearches))
	program.syncPastedPullRequestTab()
}
