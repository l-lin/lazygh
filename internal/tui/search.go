package tui

import "strings"

type SearchTargetKind int

const (
	SearchTargetUser SearchTargetKind = iota
	SearchTargetPullRequests
	SearchTargetReviewTree
	SearchTargetNotifications
	SearchTargetDetail
)

func (model *Model) SearchActive() bool {
	return model.searchActive
}

func (model *Model) SearchTarget() Focus {
	return searchTargetFocus(model.searchTargetKind)
}

func (model *Model) SearchTargetKind() SearchTargetKind {
	return model.searchTargetKind
}

func (model *Model) SearchTargetPullRequestTab() PullRequestTab {
	return model.searchTargetPullRequestTab
}

func (model *Model) SearchDraft() string {
	return model.searchDraft
}

func (model *Model) ReviewTreeSearchQuery() string {
	if model.searchActive && model.searchTargetKind == SearchTargetReviewTree {
		return model.searchDraft
	}
	return model.reviewTreeSearchQuery
}

func (model *Model) StartSearch() {
	model.StartSearchForTarget(model.Focus(), model.ActivePullRequestTab())
}

func (model *Model) StartSearchForTarget(target Focus, pullRequestTab PullRequestTab) {
	model.startSearch(searchTargetKindForFocus(target), pullRequestTab)
}

func (model *Model) StartSearchForReviewTree(pullRequestTab PullRequestTab) {
	model.startSearch(SearchTargetReviewTree, pullRequestTab)
}

func (model *Model) UpdateSearchDraft(query string) {
	if !model.searchActive {
		return
	}

	model.searchDraft = query
	model.clampSearchSelectionForTarget(model.SearchTarget(), query)
}

func (model *Model) SubmitSearch() {
	if !model.searchActive {
		return
	}

	targetKind := model.searchTargetKind
	targetFocus := searchTargetFocus(targetKind)
	tab := model.searchTargetPullRequestTab
	selectedPullRequestIndex := model.selectedPullRequestIndexes[tab]

	switch targetKind {
	case SearchTargetPullRequests:
		model.pullRequestSearchQueries[tab] = model.searchDraft
	case SearchTargetReviewTree:
		model.reviewTreeSearchQuery = model.searchDraft
	case SearchTargetDetail:
		model.detailSearchQuery = model.searchDraft
	case SearchTargetNotifications:
		model.notificationSearchQuery = model.searchDraft
	default:
		model.userSearchQuery = model.searchDraft
	}

	model.searchActive = false
	model.searchDraft = ""
	if targetKind == SearchTargetPullRequests {
		model.followSubmittedPullRequestSearch(tab, selectedPullRequestIndex)
		return
	}
	model.clampSearchSelectionForTarget(targetFocus, model.appliedSearchQuery(targetKind, tab))
}

func (model *Model) CancelSearch() {
	if !model.searchActive {
		return
	}

	targetKind := model.searchTargetKind
	targetFocus := searchTargetFocus(targetKind)
	tab := model.searchTargetPullRequestTab
	query := model.appliedSearchQuery(targetKind, tab)

	model.searchActive = false
	model.searchDraft = ""
	model.clampSearchSelectionForTarget(targetFocus, query)
}

func (model *Model) VisibleUsers() []Item {
	return filterItemsByIndexes(model.users, model.visibleUserIndexes())
}

func (model *Model) VisiblePullRequests() []Item {
	tab := model.ActivePullRequestTab()
	return filterItemsByIndexes(model.CurrentPullRequests(), model.visiblePullRequestIndexes(tab))
}

func (model *Model) VisibleNotifications() []Item {
	return filterItemsByIndexes(model.Notifications(), model.visibleNotificationIndexes())
}

func (model *Model) SelectedVisibleUserIndex() int {
	return model.selectedVisibleIndex(model.selectedUserIndex, model.visibleUserIndexes())
}

func (model *Model) SelectedVisiblePullRequestIndex(tab PullRequestTab) int {
	return model.selectedVisibleIndex(model.selectedPullRequestIndexes[tab], model.visiblePullRequestIndexes(tab))
}

func (model *Model) SelectedVisibleNotificationIndex() int {
	return model.selectedVisibleIndex(model.selectedNotificationIndex, model.visibleNotificationIndexes())
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

func (model *Model) NotificationSearchQuery() string {
	return model.effectiveSearchQuery(FocusNotificationsView, MyPullRequestsTab)
}

func (model *Model) visibleUserIndexes() []int {
	return matchingItemIndexes(model.users, model.UserSearchQuery())
}

func (model *Model) visiblePullRequestIndexes(tab PullRequestTab) []int {
	return matchingItemIndexes(model.PullRequests(tab), model.PullRequestSearchQuery(tab))
}

func (model *Model) visibleNotificationIndexes() []int {
	return matchingItemIndexes(model.Notifications(), model.NotificationSearchQuery())
}

func (model *Model) clampSearchSelectionForUserView() {
	model.clampSearchSelectionForTarget(FocusUserView, model.UserSearchQuery())
}

func (model *Model) clampSearchSelectionForPullRequestTab(tab PullRequestTab) {
	model.clampSearchSelectionForTarget(FocusPullRequestsView, model.PullRequestSearchQuery(tab))
}

func (model *Model) clampSearchSelectionForNotificationsView() {
	model.clampSearchSelectionForTarget(FocusNotificationsView, model.NotificationSearchQuery())
}

func (model *Model) clampSearchSelectionForTarget(target Focus, query string) {
	if strings.TrimSpace(query) == "" {
		return
	}

	switch target {
	case FocusPullRequestsView:
		return
	case FocusDetailView:
		return
	case FocusNotificationsView:
		visibleIndexes := matchingItemIndexes(model.Notifications(), query)
		if len(visibleIndexes) == 0 {
			return
		}
		if indexOfInt(visibleIndexes, model.selectedNotificationIndex) < 0 {
			model.selectedNotificationIndex = visibleIndexes[0]
		}
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

func (model *Model) followSubmittedPullRequestSearch(tab PullRequestTab, startIndex int) {
	matchIndexes := model.visiblePullRequestIndexes(tab)
	matchIndex := searchMatchIndexAtOrAfter(matchIndexes, startIndex)
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return
	}

	model.selectedPullRequestIndexes[tab] = matchIndexes[matchIndex]
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
	if model.searchActive && model.SearchTarget() == target {
		if target != FocusPullRequestsView || model.searchTargetPullRequestTab == tab {
			return model.searchDraft
		}
	}

	return model.appliedSearchQuery(searchTargetKindForFocus(target), tab)
}

func (model *Model) appliedSearchQuery(targetKind SearchTargetKind, tab PullRequestTab) string {
	switch targetKind {
	case SearchTargetPullRequests:
		return model.pullRequestSearchQueries[tab]
	case SearchTargetReviewTree:
		return model.reviewTreeSearchQuery
	case SearchTargetDetail:
		return model.detailSearchQuery
	case SearchTargetNotifications:
		return model.notificationSearchQuery
	default:
		return model.userSearchQuery
	}
}

func searchTargetKindForFocus(focus Focus) SearchTargetKind {
	switch focus {
	case FocusPullRequestsView:
		return SearchTargetPullRequests
	case FocusDetailView:
		return SearchTargetDetail
	case FocusNotificationsView:
		return SearchTargetNotifications
	default:
		return SearchTargetUser
	}
}

func searchTargetFocus(kind SearchTargetKind) Focus {
	switch kind {
	case SearchTargetPullRequests, SearchTargetReviewTree:
		return FocusPullRequestsView
	case SearchTargetDetail:
		return FocusDetailView
	case SearchTargetNotifications:
		return FocusNotificationsView
	default:
		return FocusUserView
	}
}

func (model *Model) startSearch(targetKind SearchTargetKind, pullRequestTab PullRequestTab) {
	model.searchActive = true
	model.searchTargetKind = targetKind
	if len(model.pullRequestTabs) > 0 {
		model.searchTargetPullRequestTab = PullRequestTab(clampIndex(int(pullRequestTab), len(model.pullRequestTabs)))
	} else {
		model.searchTargetPullRequestTab = 0
	}
	targetFocus := searchTargetFocus(targetKind)
	model.clearAppliedSearchQueriesForOtherViews(targetFocus)
	model.searchDraft = ""
	model.clampSearchSelectionForTarget(targetFocus, model.searchDraft)
}

func (model *Model) ClearReviewTreeSearchQuery() {
	model.reviewTreeSearchQuery = ""
}

func (model *Model) ClearPaneSearchQueries() {
	model.userSearchQuery = ""
	model.reviewTreeSearchQuery = ""
	model.detailSearchQuery = ""
	model.notificationSearchQuery = ""
	for tab := range model.pullRequestSearchQueries {
		model.pullRequestSearchQueries[tab] = ""
	}
}

func (model *Model) clearAppliedSearchQueriesForOtherViews(target Focus) {
	switch target {
	case FocusPullRequestsView:
		model.userSearchQuery = ""
		model.detailSearchQuery = ""
		model.notificationSearchQuery = ""
	case FocusDetailView:
		model.userSearchQuery = ""
		model.notificationSearchQuery = ""
		model.reviewTreeSearchQuery = ""
		for tab := range model.pullRequestSearchQueries {
			model.pullRequestSearchQueries[tab] = ""
		}
	case FocusNotificationsView:
		model.userSearchQuery = ""
		model.detailSearchQuery = ""
		model.reviewTreeSearchQuery = ""
		for tab := range model.pullRequestSearchQueries {
			model.pullRequestSearchQueries[tab] = ""
		}
	default:
		model.detailSearchQuery = ""
		model.notificationSearchQuery = ""
		model.reviewTreeSearchQuery = ""
		for tab := range model.pullRequestSearchQueries {
			model.pullRequestSearchQueries[tab] = ""
		}
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
