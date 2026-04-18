package tui

import "strings"

func (model *Model) SearchActive() bool {
	return model.searchActive
}

func (model *Model) SearchTarget() Focus {
	return model.searchTarget
}

func (model *Model) SearchTargetPullRequestTab() PullRequestTab {
	return model.searchTargetPullRequestTab
}

func (model *Model) SearchDraft() string {
	return model.searchDraft
}

func (model *Model) StartSearch() {
	model.searchActive = true
	model.searchTarget = model.focus
	model.searchTargetPullRequestTab = model.activePullRequestTab
	model.searchDraft = model.appliedSearchQuery(model.searchTarget, model.searchTargetPullRequestTab)
	model.clampSearchSelectionForTarget(model.searchTarget, model.searchTargetPullRequestTab, model.searchDraft)
}

func (model *Model) UpdateSearchDraft(query string) {
	if !model.searchActive {
		return
	}

	model.searchDraft = query
	model.clampSearchSelectionForTarget(model.searchTarget, model.searchTargetPullRequestTab, query)
}

func (model *Model) SubmitSearch() {
	if !model.searchActive {
		return
	}

	target := model.searchTarget
	tab := model.searchTargetPullRequestTab

	switch target {
	case FocusPullRequestsView:
		model.pullRequestSearchQueries[tab] = model.searchDraft
	case FocusDetailView:
		model.detailSearchQuery = model.searchDraft
	default:
		model.userSearchQuery = model.searchDraft
	}

	model.searchActive = false
	model.searchDraft = ""
	model.clampSearchSelectionForTarget(target, tab, model.appliedSearchQuery(target, tab))
}

func (model *Model) CancelSearch() {
	if !model.searchActive {
		return
	}

	target := model.searchTarget
	tab := model.searchTargetPullRequestTab
	query := model.appliedSearchQuery(target, tab)

	model.searchActive = false
	model.searchDraft = ""
	model.clampSearchSelectionForTarget(target, tab, query)
}

func (model *Model) VisibleUsers() []Item {
	return filterItemsByIndexes(model.users, model.visibleUserIndexes())
}

func (model *Model) VisiblePullRequests() []Item {
	return filterItemsByIndexes(model.CurrentPullRequests(), model.visiblePullRequestIndexes(model.activePullRequestTab))
}

func (model *Model) SelectedVisibleUserIndex() int {
	return model.selectedVisibleIndex(model.selectedUserIndex, model.visibleUserIndexes())
}

func (model *Model) SelectedVisiblePullRequestIndex(tab PullRequestTab) int {
	return model.selectedVisibleIndex(model.selectedPullRequestIndexes[tab], model.visiblePullRequestIndexes(tab))
}

func (model *Model) UserSearchQuery() string {
	return model.effectiveSearchQuery(FocusUserView, MyPullRequestsTab)
}

func (model *Model) DetailSearchQuery() string {
	return model.effectiveSearchQuery(FocusDetailView, MyPullRequestsTab)
}

func (model *Model) PullRequestSearchQuery(tab PullRequestTab) string {
	return model.effectiveSearchQuery(FocusPullRequestsView, tab)
}

func (model *Model) visibleUserIndexes() []int {
	return matchingItemIndexes(model.users, model.UserSearchQuery())
}

func (model *Model) visiblePullRequestIndexes(tab PullRequestTab) []int {
	return matchingItemIndexes(model.PullRequests(tab), model.PullRequestSearchQuery(tab))
}

func (model *Model) clampSearchSelectionForUserView() {
	model.clampSearchSelectionForTarget(FocusUserView, MyPullRequestsTab, model.UserSearchQuery())
}

func (model *Model) clampSearchSelectionForPullRequestTab(tab PullRequestTab) {
	model.clampSearchSelectionForTarget(FocusPullRequestsView, tab, model.PullRequestSearchQuery(tab))
}

func (model *Model) clampSearchSelectionForTarget(target Focus, tab PullRequestTab, query string) {
	if strings.TrimSpace(query) == "" {
		return
	}

	switch target {
	case FocusPullRequestsView:
		visibleIndexes := matchingItemIndexes(model.PullRequests(tab), query)
		if len(visibleIndexes) == 0 {
			return
		}
		if indexOfInt(visibleIndexes, model.selectedPullRequestIndexes[tab]) < 0 {
			model.selectedPullRequestIndexes[tab] = visibleIndexes[0]
		}
	case FocusDetailView:
		return
	default:
		visibleIndexes := matchingItemIndexes(model.users, query)
		if len(visibleIndexes) == 0 {
			return
		}
		if indexOfInt(visibleIndexes, model.selectedUserIndex) < 0 {
			model.selectedUserIndex = visibleIndexes[0]
		}
	}
}

func (model *Model) selectedVisibleIndex(selectedIndex int, visibleIndexes []int) int {
	if len(visibleIndexes) == 0 {
		return 0
	}

	visibleIndex := indexOfInt(visibleIndexes, selectedIndex)
	if visibleIndex >= 0 {
		return visibleIndex
	}

	return 0
}

func (model *Model) effectiveSearchQuery(target Focus, tab PullRequestTab) string {
	if model.searchActive && model.searchTarget == target {
		if target != FocusPullRequestsView || model.searchTargetPullRequestTab == tab {
			return model.searchDraft
		}
	}

	return model.appliedSearchQuery(target, tab)
}

func (model *Model) appliedSearchQuery(target Focus, tab PullRequestTab) string {
	switch target {
	case FocusPullRequestsView:
		return model.pullRequestSearchQueries[tab]
	case FocusDetailView:
		return model.detailSearchQuery
	default:
		return model.userSearchQuery
	}
}

func filterItemsByIndexes(items []Item, visibleIndexes []int) []Item {
	filteredItems := make([]Item, 0, len(visibleIndexes))
	for _, visibleIndex := range visibleIndexes {
		if visibleIndex < 0 || visibleIndex >= len(items) {
			continue
		}
		filteredItems = append(filteredItems, items[visibleIndex])
	}

	return filteredItems
}

func matchingItemIndexes(items []Item, query string) []int {
	visibleIndexes := make([]int, 0, len(items))
	for index, item := range items {
		if itemMatchesQuery(item, query) {
			visibleIndexes = append(visibleIndexes, index)
		}
	}

	return visibleIndexes
}

func itemMatchesQuery(item Item, query string) bool {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return true
	}

	return strings.Contains(strings.ToLower(item.Title), strings.ToLower(trimmedQuery))
}
