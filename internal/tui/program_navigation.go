package tui

import "github.com/jesseduffield/gocui"

func (program *Program) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (program *Program) nextSideView(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.sideViewCyclingBlocked() {
		return nil
	}

	program.model.NextSideView()
	if program.reviewSession.active {
		return program.refreshViewsIfGUI(gui)
	}
	return program.syncCurrentView(gui)
}

func (program *Program) previousSideView(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.sideViewCyclingBlocked() {
		return nil
	}

	program.model.PreviousSideView()
	if program.reviewSession.active {
		return program.refreshViewsIfGUI(gui)
	}
	return program.syncCurrentView(gui)
}

func (program *Program) moveSelectionDown(gui *gocui.Gui, view *gocui.View) error {
	return program.handleSelectionChange(gui, view, 1, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveDown(document, viewportHeight)
	})
}

func (program *Program) moveSelectionUp(gui *gocui.Gui, view *gocui.View) error {
	return program.handleSelectionChange(gui, view, -1, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveUp(document, viewportHeight)
	})
}

func (program *Program) moveDetailViewDown(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.PaneVisible(FocusDetailView) {
		return nil
	}

	return program.mutateDetailViewState(gui, nil, func(document detailDocument, viewportHeight int) {
		program.detailViewState.scrollDown(document, viewportHeight)
	})
}

func (program *Program) moveDetailViewUp(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.PaneVisible(FocusDetailView) {
		return nil
	}

	return program.mutateDetailViewState(gui, nil, func(document detailDocument, viewportHeight int) {
		program.detailViewState.scrollUp(document, viewportHeight)
	})
}

func (program *Program) pageDown(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, pageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailViewState.pageDown(document, viewportHeight)
	})
}

func (program *Program) pageUp(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, -pageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailViewState.pageUp(document, viewportHeight)
	})
}

func (program *Program) recenterSideSelection(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	viewName, selectedVisibleLine, lineCount := program.currentSideListState()
	target := keySequenceTargetFor(viewName, keymapScopeSide, "recenter_selection")
	return program.armOrHandleSelectionKeySequence(target, func() error {
		return program.recenterListSelection(gui, view, viewName, selectedVisibleLine, lineCount)
	})
}

func (program *Program) recenterDetailView(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.recenter(document, viewportHeight)
	})
}

func (program *Program) moveSideSelectionToTop(gui *gocui.Gui, _ *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}

	target := keySequenceTargetFor(program.currentSideViewName(), keymapScopeSide, "move_selection_to_top")
	return program.armOrHandleSelectionKeySequence(target, func() error {
		if program.reviewSession.active {
			if program.model.Focus() != FocusPullRequestsView {
				return nil
			}
			program.moveReviewSessionSelectionToTop()
			return program.refreshViewsIfGUI(gui)
		}

		program.model.MoveSelectionToTop()
		return nil
	})
}

func (program *Program) moveSideSelectionToBottom(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.reviewSession.active {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.moveReviewSessionSelectionToBottom()
		return program.refreshViewsIfGUI(gui)
	}

	program.model.MoveSelectionToBottom()
	return nil
}

func (program *Program) moveDetailCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveLeft(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveRight(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToRowStart(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToRowEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.handleGoToTopPrefix(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToBottom(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToNextWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToWordEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToPreviousWord(document, viewportHeight)
	})
}

func (program *Program) enterDetailVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.enterVisualMode()
		program.detailViewState.sync(document, viewportHeight)
	})
}

func (program *Program) enterDetailLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.enterLineVisualMode(document)
		program.detailViewState.sync(document, viewportHeight)
	})
}

func (program *Program) nextPullRequestTab(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewSession.active {
		return program.handleReviewFileMotionPrefix(gui, view, reviewNavigationForward)
	}

	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}

	program.model.NextPullRequestTab()
	program.reloadActivePullRequestsTab(gui)
	return nil
}

func (program *Program) previousPullRequestTab(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewSession.active {
		return program.handleReviewFileMotionPrefix(gui, view, reviewNavigationBackward)
	}

	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}

	program.model.PreviousPullRequestTab()
	program.reloadActivePullRequestsTab(gui)
	return nil
}

func (program *Program) focusDetailView(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.model.FocusDetailView()
	if program.reviewSession.active {
		return program.refreshViewsIfGUI(gui)
	}
	return program.syncCurrentView(gui)
}

func (program *Program) focusUserView(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.model.FocusUserView()
	if program.reviewSession.active {
		return program.refreshViewsIfGUI(gui)
	}
	return program.syncCurrentView(gui)
}

func (program *Program) focusPullRequestsView(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.model.FocusPullRequestsView()
	if program.reviewSession.active {
		return program.refreshViewsIfGUI(gui)
	}
	return program.syncCurrentView(gui)
}

func (program *Program) openDetail(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.detailTransitionBlocked() {
		return nil
	}

	program.model.OpenDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) closeDetail(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.detailTransitionBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView && program.detailViewState.mode.isVisual() {
		program.detailViewState.exitVisualMode()
		return program.refreshDetailView(gui)
	}

	program.detailViewState.clearPendingPrefix()
	program.model.CloseDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) openSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() || (program.reviewSession.active && program.model.Focus() == FocusUserView) {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	if program.reviewSession.active && program.model.Focus() == FocusPullRequestsView {
		program.startReviewFileTreeSearch()
	} else {
		if program.reviewSession.active && program.model.Focus() == FocusDetailView {
			program.reviewSession.fileTreeSearchQuery = ""
		}
		program.model.StartSearch()
	}
	program.searchEditor = newLineEditor("")
	return program.layout(gui)
}

func (program *Program) submitSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.activeSearchIsReviewFileTreeSearch() {
		program.submitReviewFileTreeSearch()
		program.searchEditor = nil
		return program.refreshViewsIfGUI(gui)
	}

	target := program.model.SearchTarget()
	program.model.SubmitSearch()
	program.searchEditor = nil

	if target == FocusDetailView {
		if actualErr := program.followSubmittedDetailSearch(gui); actualErr != nil {
			return actualErr
		}
	}

	return program.refreshViewsIfGUI(gui)
}

func (program *Program) cancelSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.activeSearchIsReviewFileTreeSearch() {
		program.cancelReviewFileTreeSearch()
		return program.closeSearch(gui)
	}

	program.model.CancelSearch()
	return program.closeSearch(gui)
}

func (program *Program) closeSearch(gui *gocui.Gui) error {
	program.searchEditor = nil
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) toggleHelp(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	if program.helpToggleBlocked() {
		return nil
	}

	program.helpVisible = !program.helpVisible
	if !program.helpVisible {
		return program.closeHelp(gui, nil)
	}

	return program.layout(gui)
}

func (program *Program) closeHelp(gui *gocui.Gui, _ *gocui.View) error {
	program.clearPendingSelectionPrefix()
	program.detailViewState.clearPendingPrefix()
	program.helpVisible = false
	return program.refreshViewsIfGUI(gui)
}
