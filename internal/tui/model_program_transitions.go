package tui

func (model *Model) ApplyProjectedScreenState(state ScreenState) {
	application := projectScreenStateApplication(state)
	model.focus = application.focus
	model.lastSideFocus = application.lastSideFocus
	if application.hasPullRequestTab {
		model.activePullRequestTab = application.activePullRequestTab
	}
}

func (model *Model) SetActivePullRequestTab(tab PullRequestTab) {
	if len(model.pullRequestTabs) == 0 {
		return
	}

	state := model.ScreenState().WithViewTabs(sidePanelPullRequestsViewNumber, int(tab), model.pullRequestScreenTabs())
	model.applyBrowserScreenState(state)
}

func (model *Model) SelectPullRequestIndex(tab PullRequestTab, index int) {
	tabIndex := int(tab)
	if tabIndex < 0 || tabIndex >= len(model.pullRequestTabs) {
		return
	}

	model.selectedPullRequestIndexes[tab] = clampIndex(index, len(model.pullRequestRows(tab)))
}

func (model *Model) SelectNotificationIndex(index int) {
	model.selectedNotificationIndex = clampIndex(index, len(model.notifications))
	model.clampSearchSelectionForNotificationsView()
}

func (model *Model) SelectUserIndex(index int) {
	model.selectedUserIndex = clampIndex(index, len(model.users))
	model.clampSearchSelectionForUserView()
}

func (model *Model) SetPullRequestSearchQuery(tab PullRequestTab, query string) {
	tabIndex := int(tab)
	if tabIndex < 0 || tabIndex >= len(model.pullRequestTabs) {
		return
	}

	model.pullRequestSearchQueries[tab] = query
	model.clampSearchSelectionForPullRequestTab(tab)
}

func (model *Model) ClearPullRequestSearchQuery(tab PullRequestTab) {
	model.SetPullRequestSearchQuery(tab, "")
}

func (model *Model) CloseSearchPrompt() {
	if !model.searchActive {
		return
	}

	model.searchActive = false
	model.searchDraft = ""
}

func (model *Model) SetPaneLayoutSize(size PaneLayoutSize) {
	model.paneLayoutSize = size
}

func (model *Model) SetFullscreenPane(focus Focus) {
	if !isMainPaneFocus(focus) {
		return
	}
	model.fullscreenPane = focus
}

func (model *Model) SetDetailFullscreenReturnSize(size PaneLayoutSize) {
	model.detailFullscreenReturnSize = size
}

func (model *Model) FocusDetailFullscreenFromSideFocus(sideFocus Focus) {
	returnSize := model.paneLayoutSize
	if returnSize == PaneLayoutFullscreen {
		returnSize = PaneLayoutDefault
	}

	model.detailFullscreenReturnSize = returnSize
	model.paneLayoutSize = PaneLayoutFullscreen
	model.fullscreenPane = FocusDetailView
	model.applyBrowserScreenState(newBrowserScreenStateWithSideFocus(FocusDetailView, browserSideFocus(sideFocus), model.activePullRequestTab, nil, model.pullRequestScreenTabs()))
}
