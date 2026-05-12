package tui

func (program *Program) screenState() ScreenState {
	state := program.baseScreenState()
	for _, overlay := range program.visibleScreenOverlays() {
		state = state.WithOverlay(overlay)
	}
	return state
}

func (program *Program) baseScreenState() ScreenState {
	if program.reviewSession.active {
		mode := ScreenModeReview
		if program.reviewSession.mode == reviewSessionModeStory {
			mode = ScreenModeStoryReview
		}
		return newReviewScreenStateWithSideFocus(mode, program.model.Focus(), program.model.currentSideFocus())
	}

	state := program.model.ScreenState()
	if program.browserShowsPullRequestDetailTabs() {
		state = state.WithViewTabs(mainPanelViewNumber, int(program.activeDetailTab), program.detailScreenTabs())
	}
	return state
}

func (program *Program) detailScreenTabs() []TabState {
	tabs := make([]TabState, 0, len(browserDetailTabs))
	for _, tab := range browserDetailTabs {
		tabs = append(tabs, TabState{Label: tab.Label()})
	}
	return tabs
}

func (program *Program) browserShowsPullRequestDetailTabs() bool {
	if program.reviewSession.active {
		return false
	}

	switch program.model.ScreenState().MainViewResolver().SourceView.Focus {
	case FocusPullRequestsView:
		_, ok := program.model.SelectedPullRequestSummary()
		return ok
	case FocusNotificationsView:
		notification, ok := program.model.SelectedNotification()
		if !ok {
			return false
		}
		_, ok = notification.PullRequestSummary()
		return ok
	default:
		return false
	}
}

func (program *Program) visibleScreenOverlays() []OverlayState {
	overlays := make([]OverlayState, 0, 6)
	if program.pullRequestBuildRunPopupVisible() {
		overlays = append(overlays, OverlayState{Kind: OverlayKindBuildInfo})
	}
	if program.searchPromptVisible() {
		overlays = append(overlays, OverlayState{Kind: OverlayKindSearch})
	}
	if program.model.ActionsPopupVisible() {
		overlays = append(overlays, OverlayState{Kind: OverlayKindActionsPopup})
	}
	if program.model.ActionsPopupSearchActive() {
		overlays = append(overlays, OverlayState{Kind: OverlayKindActionsPopupSearch})
	}
	if program.modalEditorVisible() {
		overlays = append(overlays, OverlayState{Kind: OverlayKindModalEditor})
	}
	if program.helpVisible {
		overlays = append(overlays, OverlayState{Kind: OverlayKindHelp})
	}
	return overlays
}
