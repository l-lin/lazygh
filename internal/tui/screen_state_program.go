package tui

func (program *Program) screenState() ScreenState {
	state := program.baseScreenState()
	for _, overlay := range program.visibleScreenOverlays() {
		state = state.WithOverlay(overlay)
	}
	return state
}

func (program *Program) baseScreenState() ScreenState {
	return program.modeDescriptor().ScreenState(program)
}

func (program *Program) detailScreenTabs() []TabState {
	visibleTabs := program.visibleDetailTabs()
	tabs := make([]TabState, 0, len(visibleTabs))
	for _, tab := range visibleTabs {
		tabs = append(tabs, TabState{Label: program.detailTabLabel(tab)})
	}
	return tabs
}

func (program *Program) browserShowsPullRequestDetailTabs() bool {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
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
	if program.overlayState.helpVisible {
		overlays = append(overlays, OverlayState{Kind: OverlayKindHelp})
	}
	return overlays
}
