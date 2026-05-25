package tui

import appconfig "github.com/l-lin/lazygh/internal/config"

type pullRequestCountState struct {
	count int
	known bool
}

func (program *Program) ApplyPullRequestSearches(searches []appconfig.PullRequestSearch) {
	Update(program, MsgPullRequestSearchesApplied{Searches: appconfig.ResolvePullRequestSearches(searches)})
}

func pullRequestTabSeedsForSearches(searches []appconfig.PullRequestSearch) []PullRequestTabSeed {
	seeds := make([]PullRequestTabSeed, 0, len(searches))
	for _, search := range searches {
		state := buildPullRequestListState(search)
		seeds = append(seeds, PullRequestTabSeed{Label: search.Label, PullRequests: []Item{pullRequestLoadingItem(state)}})
	}
	return seeds
}

func (program *Program) pullRequestSearch(tab PullRequestTab) appconfig.PullRequestSearch {
	searches := appconfig.ResolvePullRequestSearches(program.runtimeConfig.pullRequestSearches)
	index := int(tab)
	if index < 0 || index >= len(searches) {
		return searches[0]
	}
	return searches[index]
}

func (program *Program) pullRequestListState(tab PullRequestTab) pullRequestListState {
	return buildPullRequestListState(program.pullRequestSearch(tab))
}

func (program *Program) pullRequestLoadingItem(tab PullRequestTab) Item {
	return pullRequestLoadingItem(program.pullRequestListState(tab))
}

func (program *Program) pullRequestsTabLabels() []string {
	tabs := program.model.PullRequestTabs()
	labels := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		labels = append(labels, program.pullRequestsTabLabel(tab))
	}
	return labels
}

func (program *Program) pullRequestsTabLabel(tab PullRequestTab) string {
	label := program.model.PullRequestTabLabel(tab)
	count, ok := program.pullRequestsCount(tab)
	if !ok {
		return label
	}

	return label + " (" + itoa(count) + ")"
}

func (program *Program) pullRequestsCount(tab PullRequestTab) (int, bool) {
	switch tab {
	case MyPullRequestsTab:
		return program.myPullRequestsCount, program.myPullRequestsCountKnown
	case RequestedPullRequestsTab:
		return program.requestedPullRequestsCount, program.requestedPullRequestsCountKnown
	default:
		state, ok := program.additionalPullRequestsCounts[tab]
		return state.count, ok && state.known
	}
}

func (program *Program) isPullRequestLoadingItem(item Item) bool {
	for _, tab := range program.model.PullRequestTabs() {
		state := program.pullRequestListState(tab)
		if item.Title == state.loadingTitle && item.Detail == state.loadingDetail {
			return true
		}
	}
	return false
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}

	negative := value < 0
	if negative {
		value = -value
	}

	buffer := [20]byte{}
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		index--
		buffer[index] = '-'
	}
	return string(buffer[index:])
}
