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
			program.detailState.viewState.moveDown(document, viewportHeight)
		})
	}

	return program.dispatch(gui, MsgMoveSideSelection{Delta: 1})
}

func (program *Program) moveSelectionUp(gui *gocui.Gui, view *gocui.View) error {
	if program.model.Focus() == FocusDetailView || program.actionContext().IsReviewContext() {
		return program.handleSelectionChange(gui, view, -1, func(document detailDocument, viewportHeight int) {
			program.detailState.viewState.moveUp(document, viewportHeight)
		})
	}

	return program.dispatch(gui, MsgMoveSideSelection{Delta: -1})
}

func (program *Program) moveDetailViewDown(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.PaneVisible(FocusDetailView) {
		return nil
	}

	return program.mutateDetailViewState(gui, nil, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.scrollDown(document, viewportHeight)
	})
}

func (program *Program) moveDetailViewUp(gui *gocui.Gui, _ *gocui.View) error {
	if !program.model.PaneVisible(FocusDetailView) {
		return nil
	}

	return program.mutateDetailViewState(gui, nil, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.scrollUp(document, viewportHeight)
	})
}

func (program *Program) pageDown(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, pageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.pageDown(document, viewportHeight)
	})
}

func (program *Program) pageUp(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, -pageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.pageUp(document, viewportHeight)
	})
}

func (program *Program) fullPageDown(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, fullPageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.fullPageDown(document, viewportHeight)
	})
}

func (program *Program) fullPageUp(gui *gocui.Gui, view *gocui.View) error {
	actualView := program.resolveView(gui, view, program.currentViewName())
	pageSize := viewPageSize(actualView)
	return program.handlePageChange(gui, actualView, -fullPageDelta(pageSize), func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.fullPageUp(document, viewportHeight)
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
		program.detailState.viewState.recenter(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToViewportTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.placeCursorAtViewportTop(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToViewportBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.placeCursorAtViewportBottom(document, viewportHeight)
	})
}

func (program *Program) moveSideSelectionToTop(gui *gocui.Gui, _ *gocui.View) error {
	if program.actionContext().IsReviewContext() {
		return program.dispatch(gui, MsgMoveReviewSelectionToTop{})
	}

	return program.dispatch(gui, MsgMoveSideSelectionToTop{})
}

func (program *Program) moveSideSelectionToBottom(gui *gocui.Gui, _ *gocui.View) error {
	if program.actionContext().IsReviewContext() {
		return program.dispatch(gui, MsgMoveReviewSelectionToBottom{})
	}

	return program.dispatch(gui, MsgMoveSideSelectionToBottom{})
}

func (program *Program) moveDetailCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveLeft(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveRight(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToRowStart(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToRowEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToTop(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToBottom(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterExclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToNextWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToWordEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterExclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToNextBigWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToBigWordEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToPreviousWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionCharacterInclusive, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.moveToPreviousBigWord(document, viewportHeight)
	})
}

func (program *Program) enterDetailVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.enterVisualMode()
		program.syncCurrentDetailViewport(document, viewportHeight)
	})
}

func (program *Program) enterDetailLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailState.viewState.enterLineVisualMode(document)
		program.syncCurrentDetailViewport(document, viewportHeight)
	})
}

func (program *Program) nextPullRequestTab(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgAdvancePullRequestTab{Delta: 1})
}

func (program *Program) previousPullRequestTab(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgAdvancePullRequestTab{Delta: -1})
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
	return program.dispatch(gui, MsgOpenDetailRequested{})
}

func (program *Program) closeDetail(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgCloseDetailRequested{})
}

func (program *Program) openSearch(gui *gocui.Gui, _ *gocui.View) error {
	return program.openSearchWithInitialQuery(gui, "")
}

func (program *Program) searchWordUnderCursorForward(gui *gocui.Gui, view *gocui.View) error {
	return program.searchWordUnderCursor(gui, view, false)
}

func (program *Program) searchWordUnderCursorBackward(gui *gocui.Gui, view *gocui.View) error {
	return program.searchWordUnderCursor(gui, view, true)
}

func (program *Program) searchWordUnderCursor(gui *gocui.Gui, view *gocui.View, reverse bool) error {
	return program.dispatch(gui, MsgSearchWordUnderCursor{View: view, Reverse: reverse})
}

func (program *Program) openSearchWithInitialQuery(gui *gocui.Gui, query string) error {
	if program.inputContext().SearchUsesReviewTree {
		return program.dispatch(gui, MsgStartReviewFileTreeSearch{Query: query})
	}
	return program.dispatch(gui, MsgOpenSearch{Query: query})
}

func (program *Program) submitSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.activeSearchIsReviewFileTreeSearch() {
		return program.dispatch(gui, MsgSubmitReviewFileTreeSearch{})
	}
	return program.dispatch(gui, MsgSubmitSearch{})
}

func (program *Program) cancelSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.activeSearchIsReviewFileTreeSearch() {
		return program.dispatch(gui, MsgCancelReviewFileTreeSearch{})
	}
	return program.dispatch(gui, MsgCancelSearch{})
}

func (program *Program) closeSearch(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgCloseSearch{})
}

func (program *Program) toggleHelp(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgToggleHelp{})
}

func (program *Program) closeHelp(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgCloseHelp{})
}
