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
	return program.dispatch(gui, MsgLineNavigationRequested{Delta: 1})
}

func (program *Program) moveSelectionUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgLineNavigationRequested{Delta: -1})
}

func (program *Program) moveDetailViewDown(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailViewportRequested{Operation: detailViewportOperationScrollDown})
}

func (program *Program) moveDetailViewUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailViewportRequested{Operation: detailViewportOperationScrollUp})
}

func (program *Program) pageDown(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgPageNavigationRequested{Kind: pageNavigationKindHalfDown})
}

func (program *Program) pageUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgPageNavigationRequested{Kind: pageNavigationKindHalfUp})
}

func (program *Program) fullPageDown(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgPageNavigationRequested{Kind: pageNavigationKindFullDown})
}

func (program *Program) fullPageUp(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgPageNavigationRequested{Kind: pageNavigationKindFullUp})
}

func (program *Program) recenterSideSelection(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSideListViewportRequested{Placement: viewportPlacementCenter})
}

func (program *Program) moveSideSelectionToViewportTop(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSideListViewportRequested{Placement: viewportPlacementTop})
}

func (program *Program) moveSideSelectionToViewportBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgSideListViewportRequested{Placement: viewportPlacementBottom})
}

func (program *Program) recenterDetailView(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailViewportRequested{Operation: detailViewportOperationRecenter})
}

func (program *Program) moveDetailCursorToViewportTop(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailViewportRequested{Operation: detailViewportOperationPlaceTop})
}

func (program *Program) moveDetailCursorToViewportBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailViewportRequested{Operation: detailViewportOperationPlaceBottom})
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
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveLeft})
}

func (program *Program) moveDetailCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveRight})
}

func (program *Program) moveDetailCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToRowStart})
}

func (program *Program) moveDetailCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToRowEnd})
}

func (program *Program) moveDetailCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToTop})
}

func (program *Program) moveDetailCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToBottom})
}

func (program *Program) moveDetailCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToNextWord})
}

func (program *Program) moveDetailCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToWordEnd})
}

func (program *Program) moveDetailCursorToNextBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToNextBigWord})
}

func (program *Program) moveDetailCursorToBigWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToBigWordEnd})
}

func (program *Program) moveDetailCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToPreviousWord})
}

func (program *Program) moveDetailCursorToPreviousBigWord(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationMoveToPreviousBigWord})
}

func (program *Program) enterDetailVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationEnterVisualMode})
}

func (program *Program) enterDetailLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgDetailMotionRequested{Target: detailMotionTargetDetail, Operation: detailMotionOperationEnterLineVisualMode})
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
	return program.dispatch(gui, MsgSearchWordUnderCursor{Reverse: reverse})
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
