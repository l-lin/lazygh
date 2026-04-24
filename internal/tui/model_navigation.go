package tui

func (model *Model) detailItem() (Item, bool) {
	switch model.currentSideFocus() {
	case FocusPullRequestsView:
		row, ok := model.SelectedPullRequestRow()
		if !ok {
			return Item{}, false
		}
		return row.Item, true
	default:
		return itemAt(model.users, model.selectedUserIndex)
	}
}

func (model *Model) currentSideFocus() Focus {
	if model.focus == FocusDetailView {
		return model.lastSideFocus
	}

	if model.focus == FocusPullRequestsView {
		return FocusPullRequestsView
	}

	return FocusUserView
}

func (model *Model) setSideFocus(focus Focus) {
	if focus != FocusUserView && focus != FocusPullRequestsView {
		return
	}
	if model.paneLayoutSize == PaneLayoutFullscreen && focus != model.fullscreenPane {
		return
	}

	model.focus = focus
	model.lastSideFocus = focus
}

func (model *Model) adjustSelectionBy(change int) {
	switch model.focus {
	case FocusUserView:
		model.selectedUserIndex = adjustVisibleSelection(model.selectedUserIndex, model.visibleUserIndexes(), change)
	case FocusPullRequestsView:
		model.adjustPullRequestSelection(change)
	}
}

func (model *Model) adjustPullRequestSelection(change int) {
	tab := model.activePullRequestTab
	selectedIndex := model.selectedPullRequestIndexes[tab]
	visibleIndexes := model.visiblePullRequestIndexes(tab)
	model.selectedPullRequestIndexes[tab] = adjustVisibleSelection(selectedIndex, visibleIndexes, change)
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

func (tab PullRequestTab) Label() string {
	switch tab {
	case RequestedPullRequestsTab:
		return "Requested"
	case MyPullRequestsTab:
		return "My PRs"
	default:
		return "Pull Requests"
	}
}
