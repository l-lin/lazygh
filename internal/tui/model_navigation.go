package tui

func (model *Model) detailItem() (Item, bool) {
	resolver := model.ScreenState().MainViewResolver()
	switch resolver.SourceView.Focus {
	case FocusPullRequestsView:
		row, ok := model.SelectedPullRequestRow()
		if !ok {
			return Item{}, false
		}
		return row.Item, true
	case FocusNotificationsView:
		row, ok := model.SelectedNotificationRow()
		if !ok {
			return Item{}, false
		}
		return row.Item, true
	default:
		return itemAt(model.users, model.selectedUserIndex)
	}
}

func (model *Model) currentSideFocus() Focus {
	return model.ScreenState().ActiveSideView().Focus
}

func (model *Model) setSideFocus(focus Focus) {
	if focus != FocusUserView && focus != FocusPullRequestsView && focus != FocusNotificationsView {
		return
	}
	if model.paneLayoutSize == PaneLayoutFullscreen && focus != model.fullscreenPane {
		return
	}

	model.applyBrowserScreenState(model.ScreenState().FocusViewNumber(browserSideViewNumber(focus)))
}

func (model *Model) adjustSelectionBy(change int) {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), change)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(change)
	case FocusNotificationsView:
		model.adjustNotificationSelection(change)
	}
}

func (model *Model) adjustPullRequestSelection(change int) {
	tab := model.ActivePullRequestTab()
	selectedIndex := model.selectedPullRequestIndexes[tab]
	visibleIndexes := model.visiblePullRequestIndexes(tab)
	model.selectedPullRequestIndexes[tab] = adjustVisibleSelection(selectedIndex, visibleIndexes, change)
}

func (model *Model) adjustNotificationSelection(change int) {
	model.selectedNotificationIndex = adjustVisibleSelection(model.selectedNotificationIndex, model.visibleNotificationIndexes(), change)
}

func adjustVisibleSelection(selectedIndex int, visibleIndexes []int, change int) int {
	if len(visibleIndexes) == 0 {
		return selectedIndex
	}

	visibleSelectionIndex := indexOfInt(visibleIndexes, selectedIndex)
	if visibleSelectionIndex < 0 {
		visibleSelectionIndex = 0
	}

	visibleSelectionIndex = clampIndex(visibleSelectionIndex+change, len(visibleIndexes))
	return visibleIndexes[visibleSelectionIndex]
}

func firstVisibleIndex(selectedIndex int, visibleIndexes []int) int {
	if len(visibleIndexes) == 0 {
		return selectedIndex
	}

	return visibleIndexes[0]
}

func lastVisibleIndex(selectedIndex int, visibleIndexes []int) int {
	if len(visibleIndexes) == 0 {
		return selectedIndex
	}

	return visibleIndexes[len(visibleIndexes)-1]
}

func (tab PullRequestTab) Label() string {
	return "Pull Requests"
}
