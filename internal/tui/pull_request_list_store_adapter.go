package tui

func (program *Program) updatePullRequestListStore(transition func(pullRequestListStore) pullRequestListStore) {
	if program == nil || program.pullRequestListStore == nil {
		return
	}

	updatedStore := transition(*program.pullRequestListStore)
	program.pullRequestListStore = &updatedStore
}

func (program *Program) setPullRequestsLoadStarted(tab PullRequestTab, value bool) {
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestsLoadStarted(tab, value)
	})
}

func (program *Program) setPullRequestsLoading(tab PullRequestTab, value bool) {
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestsLoading(tab, value)
	})
}

func (program *Program) setPullRequestsCount(tab PullRequestTab, count int, known bool) {
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestsCount(tab, count, known)
	})
}

func (program *Program) resetPullRequestListLoadState() {
	program.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withLoadStateReset()
	})
}
