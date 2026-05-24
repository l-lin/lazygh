package tui

import "github.com/jesseduffield/gocui"

func (program *Program) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (program *Program) nextSideView(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgNextSideView{})
}

func (program *Program) previousSideView(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgPreviousSideView{})
}

func (program *Program) moveSelectionDown(gui *gocui.Gui, view *gocui.View) error {
	if program.model.Focus() == FocusDetailView || program.actionContext().IsReviewContext() {
		return program.handleSelectionChange(gui, view, 1, func(document detailDocument, viewportHeight int) {
			program.detailViewState.moveDown(document, viewportHeight)
		})
	}

	return program.dispatch(gui, MsgMoveSideSelection{Delta: 1})
}

func (program *Program) moveSelectionUp(gui *gocui.Gui, view *gocui.View) error {
	if program.model.Focus() == FocusDetailView || program.actionContext().IsReviewContext() {
		return program.handleSelectionChange(gui, view, -1, func(document detailDocument, viewportHeight int) {
			program.detailViewState.moveUp(document, viewportHeight)
		})
	}

	return program.dispatch(gui, MsgMoveSideSelection{Delta: -1})
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

func (program *Program) fullPageDown(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, fullPageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailViewState.fullPageDown(document, viewportHeight)
	})
}

func (program *Program) fullPageUp(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, -fullPageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailViewState.fullPageUp(document, viewportHeight)
	})
}

func (program *Program) recenterSideSelection(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	viewName, selectedVisibleLine, lineCount := program.currentSideListState()
	return program.recenterListSelection(gui, view, viewName, selectedVisibleLine, lineCount)
}

func (program *Program) moveSideSelectionToViewportTop(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	viewName, selectedVisibleLine, lineCount := program.currentSideListState()
	return program.placeListSelection(gui, view, viewName, selectedVisibleLine, lineCount, viewportPlacementTop)
}

func (program *Program) moveSideSelectionToViewportBottom(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		program.clearPendingSelectionPrefix()
		return nil
	}

	viewName, selectedVisibleLine, lineCount := program.currentSideListState()
	return program.placeListSelection(gui, view, viewName, selectedVisibleLine, lineCount, viewportPlacementBottom)
}

func (program *Program) recenterDetailView(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.recenter(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToViewportTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.placeCursorAtViewportTop(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToViewportBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.placeCursorAtViewportBottom(document, viewportHeight)
	})
}

func (program *Program) moveSideSelectionToTop(gui *gocui.Gui, _ *gocui.View) error {
	if program.actionContext().IsReviewContext() {
		program.clearPendingSelectionPrefix()
		if program.selectionChangeBlocked() {
			return nil
		}
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.moveReviewSessionSelectionToTop()
		return program.refreshViewsIfGUI(gui)
	}

	return program.dispatch(gui, MsgMoveSideSelectionToTop{})
}

func (program *Program) moveSideSelectionToBottom(gui *gocui.Gui, _ *gocui.View) error {
	if program.actionContext().IsReviewContext() {
		program.clearPendingSelectionPrefix()
		if program.selectionChangeBlocked() {
			return nil
		}
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.moveReviewSessionSelectionToBottom()
		return program.refreshViewsIfGUI(gui)
	}

	return program.dispatch(gui, MsgMoveSideSelectionToBottom{})
}

func (program *Program) moveDetailCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveLeft(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveRight(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToRowStart(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToRowEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToTop(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToBottom(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterExclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToNextWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToWordEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterExclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToNextBigWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToBigWordEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToPreviousWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToPreviousBigWord(document, viewportHeight)
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
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return nil
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
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return nil
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
	return program.dispatch(gui, MsgFocusPanelView{Number: mainPanelViewNumber})
}

func (program *Program) focusUserView(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgFocusPanelView{Number: sidePanelUserViewNumber})
}

func (program *Program) focusPullRequestsView(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgFocusPanelView{Number: sidePanelPullRequestsViewNumber})
}

func (program *Program) focusNotificationsView(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgFocusPanelView{Number: sidePanelNotificationsViewNumber})
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
	if program.model.Focus() == FocusDetailView && program.detailViewState.hasPendingYank() {
		program.detailViewState.clearPendingPrefix()
		return program.refreshDetailView(gui)
	}

	program.detailViewState.clearPendingPrefix()
	program.model.CloseDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) openSearch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgOpenSearch{Query: ""})
}

func (program *Program) searchWordUnderCursorForward(gui *gocui.Gui, view *gocui.View) error {
	return program.searchWordUnderCursor(gui, view, false)
}

func (program *Program) searchWordUnderCursorBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.searchWordUnderCursor(gui, view, true)
}

func (program *Program) searchWordUnderCursor(gui *gocui.Gui, view *gocui.View, reverse bool) error {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() {
		return nil
	}

	actualView := program.resolveView(gui, view, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	query, ok := document.wordAt(program.detailViewState.cursor)
	if !ok {
		return nil
	}

	inputContext := program.inputContext()
	program.detailViewState.clearPendingPrefix()
	if inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusDetailView {
		program.reviewSession.fileTreeSearchQuery = ""
	}
	program.model.StartSearch()
	program.updateActiveSearchDraft(query)
	program.model.SubmitSearch()
	program.detailSearchReversed = reverse
	program.searchEditor = nil

	if reverse {
		if actualErr := program.followReverseDetailSearch(gui); actualErr != nil {
			return actualErr
		}
	} else {
		if actualErr := program.followSubmittedDetailSearch(gui); actualErr != nil {
			return actualErr
		}
	}
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) openSearchWithInitialQuery(gui *gocui.Gui, query string) error {
	return program.dispatch(gui, MsgOpenSearch{Query: query})
}

func (program *Program) submitSearch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgSubmitSearch{})
}

func (program *Program) cancelSearch(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgCancelSearch{})
}

func (program *Program) closeSearch(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgCloseSearch{})
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
