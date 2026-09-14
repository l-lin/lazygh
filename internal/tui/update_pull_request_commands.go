package tui

func (program *Program) applyPullRequestConfigApplied(message MsgPullRequestConfigApplied) []Cmd {
	program.cancelScheduledPullRequestRefreshBatches()
	program.setRuntimePullRequestConfig(message.Config)
	program.resetPullRequestListLoadState()
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withoutPullRequestMemberships()
	})
	program.queuePullRequestSearchMembershipReconciliation()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestConfig.Searches))
	program.syncPastedPullRequestTab()

	program.pullRequestRefreshScheduleGeneration++
	return []Cmd{configurePullRequestRefreshSchedulerCmd{
		config:     program.runtimeConfig.pullRequestConfig,
		generation: program.pullRequestRefreshScheduleGeneration,
	}}
}
