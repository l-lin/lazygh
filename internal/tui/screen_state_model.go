package tui

func (model *Model) ScreenState() ScreenState {
	return model.browserScreenState()
}

func (model *Model) browserScreenState() ScreenState {
	return newBrowserScreenStateWithSideFocus(
		model.focus,
		model.legacySideFocus(),
		model.activePullRequestTab,
		nil,
		model.pullRequestScreenTabs(),
	)
}

func (model *Model) legacySideFocus() Focus {
	switch model.lastSideFocus {
	case FocusPullRequestsView, FocusNotificationsView:
		return model.lastSideFocus
	default:
		return FocusUserView
	}
}

func (model *Model) pullRequestScreenTabs() []TabState {
	if len(model.pullRequestTabs) == 0 {
		return nil
	}

	tabs := make([]TabState, 0, len(model.pullRequestTabs))
	for _, tab := range model.pullRequestTabs {
		tabs = append(tabs, TabState{Label: tab.label})
	}
	return tabs
}

func (model *Model) applyBrowserScreenState(state ScreenState) {
	model.focus = state.ActiveView().Focus
	model.lastSideFocus = state.ActiveSideView().Focus
	if pullRequestView, ok := state.ViewByNumber(sidePanelPullRequestsViewNumber); ok {
		model.activePullRequestTab = PullRequestTab(clampScreenTabIndex(pullRequestView.ActiveTab, len(pullRequestView.Tabs)))
	}
}

func (model *Model) FocusViewNumber(number int) {
	switch number {
	case mainPanelViewNumber:
		model.FocusDetailView()
	case sidePanelPullRequestsViewNumber:
		model.FocusPullRequestsView()
	case sidePanelNotificationsViewNumber:
		model.FocusNotificationsView()
	default:
		model.FocusUserView()
	}
}
