package tui

import appconfig "github.com/l-lin/lazygh/internal/config"

func (program *Program) applyPullRequestSearchesApplied(message MsgPullRequestSearchesApplied) []Cmd {
	program.setRuntimePullRequestSearches(message.Searches)
	program.resetPullRequestListLoadState()
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withoutPullRequestMemberships()
	})
	program.queuePullRequestSearchMembershipReconciliation()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.runtimeConfig.pullRequestSearches))
	program.syncPastedPullRequestTab()

	program.pullRequestRefreshScheduleGeneration++
	return []Cmd{configurePullRequestRefreshSchedulerCmd{
		searches:   append([]appconfig.PullRequestSearch(nil), program.runtimeConfig.pullRequestSearches...),
		generation: program.pullRequestRefreshScheduleGeneration,
	}}
}
