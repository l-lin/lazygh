package tui

type projectedScreenStateApplication struct {
	focus                Focus
	lastSideFocus        Focus
	activePullRequestTab PullRequestTab
	hasPullRequestTab    bool
	activeDetailTab      DetailTab
	hasDetailTab         bool
}

func projectScreenStateApplication(state ScreenState) projectedScreenStateApplication {
	application := projectedScreenStateApplication{
		focus:         state.ActiveView().Focus,
		lastSideFocus: state.ActiveSideView().Focus,
	}
	if state.Mode != ScreenModeBrowser {
		return application
	}
	if pullRequestView, ok := state.ViewByNumber(sidePanelPullRequestsViewNumber); ok {
		application.activePullRequestTab = PullRequestTab(clampScreenTabIndex(pullRequestView.ActiveTab, len(pullRequestView.Tabs)))
		application.hasPullRequestTab = true
	}
	if mainView, ok := state.ViewByNumber(mainPanelViewNumber); ok && len(mainView.Tabs) > 0 {
		application.activeDetailTab = DetailTab(clampScreenTabIndex(mainView.ActiveTab, len(mainView.Tabs)))
		application.hasDetailTab = true
	}
	return application
}
