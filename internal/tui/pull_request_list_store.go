package tui

func (store pullRequestListStore) pullRequestsLoadStarted(tab PullRequestTab) bool {
	switch tab {
	case MyPullRequestsTab:
		return store.myPullRequestsLoadStarted
	case RequestedPullRequestsTab:
		return store.requestedPullRequestsLoadStarted
	default:
		return store.additionalPullRequestsLoadStarted[tab]
	}
}

func (store pullRequestListStore) withPullRequestsLoadStarted(tab PullRequestTab, value bool) pullRequestListStore {
	switch tab {
	case MyPullRequestsTab:
		store.myPullRequestsLoadStarted = value
	case RequestedPullRequestsTab:
		store.requestedPullRequestsLoadStarted = value
	default:
		store.additionalPullRequestsLoadStarted = copyPullRequestTabBoolState(store.additionalPullRequestsLoadStarted)
		store.additionalPullRequestsLoadStarted[tab] = value
	}
	return store
}

func (store pullRequestListStore) withPullRequestsLoading(tab PullRequestTab, value bool) pullRequestListStore {
	switch tab {
	case MyPullRequestsTab:
		store.myPullRequestsLoading = value
	case RequestedPullRequestsTab:
		store.requestedPullRequestsLoading = value
	default:
		store.additionalPullRequestsLoading = copyPullRequestTabBoolState(store.additionalPullRequestsLoading)
		store.additionalPullRequestsLoading[tab] = value
	}
	return store
}

func (store pullRequestListStore) withPullRequestsCount(tab PullRequestTab, count int, known bool) pullRequestListStore {
	switch tab {
	case MyPullRequestsTab:
		store.myPullRequestsCount = count
		store.myPullRequestsCountKnown = known
	case RequestedPullRequestsTab:
		store.requestedPullRequestsCount = count
		store.requestedPullRequestsCountKnown = known
	default:
		store.additionalPullRequestsCounts = copyPullRequestCountStates(store.additionalPullRequestsCounts)
		store.additionalPullRequestsCounts[tab] = pullRequestCountState{count: count, known: known}
	}
	return store
}

func (store pullRequestListStore) withLoadStateReset() pullRequestListStore {
	store.myPullRequestsLoadStarted = false
	store.requestedPullRequestsLoadStarted = false
	store.myPullRequestsLoading = false
	store.requestedPullRequestsLoading = false
	store.myPullRequestsCount = 0
	store.myPullRequestsCountKnown = false
	store.requestedPullRequestsCount = 0
	store.requestedPullRequestsCountKnown = false
	store.additionalPullRequestsLoadStarted = map[PullRequestTab]bool{}
	store.additionalPullRequestsLoading = map[PullRequestTab]bool{}
	store.additionalPullRequestsCounts = map[PullRequestTab]pullRequestCountState{}
	return store
}

func copyPullRequestTabBoolState(source map[PullRequestTab]bool) map[PullRequestTab]bool {
	copied := make(map[PullRequestTab]bool, len(source))
	for tab, value := range source {
		copied[tab] = value
	}
	return copied
}

func copyPullRequestCountStates(source map[PullRequestTab]pullRequestCountState) map[PullRequestTab]pullRequestCountState {
	copied := make(map[PullRequestTab]pullRequestCountState, len(source))
	for tab, state := range source {
		copied[tab] = state
	}
	return copied
}
