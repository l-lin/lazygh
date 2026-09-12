package tui

func (program *Program) applyPullRequestSearchesApplied(message MsgPullRequestSearchesApplied) {
	program.setRuntimePullRequestSearches(message.Searches)
	program.resetPullRequestListLoadState()
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withoutPullRequestMemberships()
	})
	program.queuePullRequestSearchMembershipReconciliation()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestSearches))
	program.syncPastedPullRequestTab()
}
