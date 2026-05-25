package tui

import "strings"

func searchMatchChooserForDirection(direction searchRepeatDirection) searchMatchIndexChooser {
	if direction == searchRepeatBackward {
		return searchMatchIndexBefore
	}
	return searchMatchIndexAfter
}

func (program *Program) applyProjectedScreenState(state ScreenState) {
	application := projectScreenStateApplication(state)
	program.model.ApplyProjectedScreenState(state)
	if application.hasDetailTab {
		program.detailState.activeTab = application.activeDetailTab
	}
}

func (program *Program) applyMoveSideSelection(message MsgMoveSideSelection) {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return
	}
	program.model.adjustSelectionBy(message.Delta)
}

func (program *Program) applyMoveSideSelectionToTop() {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return
	}
	program.model.MoveSelectionToTop()
}

func (program *Program) applyMoveSideSelectionToBottom() {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return
	}
	program.model.MoveSelectionToBottom()
}

func (program *Program) applyAdvancePullRequestTab(message MsgAdvancePullRequestTab) []Cmd {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return nil
	}

	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() || message.Delta == 0 {
		return nil
	}

	if message.Delta > 0 {
		program.model.NextPullRequestTab()
	} else {
		program.model.PreviousPullRequestTab()
	}
	return []Cmd{reloadPullRequestsTabCmd{tab: program.model.ActivePullRequestTab()}}
}

func (program *Program) applyOpenDetailRequested() {
	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.detailTransitionBlocked() {
		return
	}
	program.model.OpenDetail()
}

func (program *Program) applyCloseDetailRequested() {
	program.clearPendingSelectionPrefix()
	if program.detailTransitionBlocked() {
		return
	}
	if program.model.Focus() == FocusDetailView && program.detailState.viewState.mode.isVisual() {
		program.detailState.viewState.exitVisualMode()
		return
	}
	if program.model.Focus() == FocusDetailView && program.detailState.viewState.hasPendingYank() {
		program.detailState.viewState.clearPendingPrefix()
		return
	}

	program.detailState.viewState.clearPendingPrefix()
	program.model.CloseDetail()
}

func (program *Program) applySearchDraftChanged(message MsgSearchDraftChanged) {
	program.model.UpdateSearchDraft(message.Query)
}

func (program *Program) applyStartReviewFileTreeSearch(message MsgStartReviewFileTreeSearch) {
	inputContext := program.inputContext()
	if program.mainPaneActionBlocked() || !inputContext.SearchUsesReviewTree || (inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusUserView) {
		return
	}
	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	program.model.StartSearchForReviewTree(program.model.ActivePullRequestTab())
	program.applySearchDraftChanged(MsgSearchDraftChanged{Query: message.Query})
	program.searchWidget.editor = newLineEditor(message.Query)
}

func (program *Program) applySubmitReviewFileTreeSearch() {
	if !program.activeSearchIsReviewFileTreeSearch() {
		return
	}
	program.model.SubmitSearch()
	program.followSubmittedReviewFileTreeSearch(program.model.ReviewTreeSearchQuery())
	program.searchWidget.editor = nil
}

func (program *Program) applyCancelReviewFileTreeSearch() {
	if !program.activeSearchIsReviewFileTreeSearch() {
		return
	}
	program.model.CancelSearch()
	program.searchWidget.editor = nil
}

func (program *Program) applyOpenPullRequestInBrowserView(message MsgOpenPullRequestInBrowserView) {
	summary := message.Summary
	program.pinOpenedPullRequestSummary(MyPullRequestsTab, summary)
	program.model.SetActivePullRequestTab(MyPullRequestsTab)
	program.model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(summary)})
	program.model.SelectPullRequestIndex(MyPullRequestsTab, 0)
	program.setPullRequestsLoadStarted(MyPullRequestsTab, true)
	program.setPullRequestsLoading(MyPullRequestsTab, false)
	program.setPullRequestsCount(MyPullRequestsTab, 1, true)
	program.navigationState.reviewSession = reviewSessionState{}
	program.invalidateReviewDiffRenderCache()
	program.detailState.activeTab = DescriptionDetailTab
	program.detailState.viewState.reset()
	program.detailState.viewState.clearPendingPrefix()
	program.clearPendingSelectionPrefix()
	program.invalidatePullRequestDetailDocumentCache()
	Update(program, MsgOpenPullRequestInDetailFullscreen{SideFocus: FocusPullRequestsView})
}

func (program *Program) applyOpenPullRequestInDetailFullscreen(message MsgOpenPullRequestInDetailFullscreen) {
	program.model.FocusDetailFullscreenFromSideFocus(message.SideFocus)
}

func (program *Program) applyFeedbackSet(message MsgFeedbackSet) {
	program.setFeedback(message.Target, strings.TrimSpace(message.Message))
}

func (program *Program) applyAdvanceDetailTab(message MsgAdvanceDetailTab) {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return
	}
	if program.overlayState.helpVisible || program.model.SearchActive() || !program.shouldShowPullRequestDetailTabs() {
		return
	}
	program.detailState.viewState.clearPendingPrefix()
	count := len(browserDetailTabs)
	if count == 0 || message.Delta == 0 {
		return
	}
	index := int(program.detailState.activeTab)
	if message.Delta > 0 {
		program.detailState.activeTab = DetailTab((index + 1) % count)
		return
	}
	program.detailState.activeTab = DetailTab((index + count - 1) % count)
}

func (program *Program) applyExitReviewMode() {
	if !program.reviewModeActive() {
		return
	}
	focus := program.navigationState.reviewSession.sourceFocus
	pendingReviewID := strings.TrimSpace(program.navigationState.reviewSession.pendingReviewID)
	program.restorePullRequestBrowserFromReviewMode()
	if pendingReviewID != "" {
		program.setFeedback(focus, pendingPullRequestReviewKeptOpenMessage)
	}
}

func (program *Program) applyToggleHelp() {
	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	if program.helpToggleBlocked() {
		return
	}
	program.overlayState.helpVisible = !program.overlayState.helpVisible
}

func (program *Program) applyCloseHelp() {
	program.clearPendingSelectionPrefix()
	program.detailState.viewState.clearPendingPrefix()
	program.overlayState.helpVisible = false
}

func (program *Program) applyAdjustFocusedPane(message MsgAdjustFocusedPane) {
	if program.overlayState.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() {
		return
	}
	if message.Delta > 0 {
		program.model.GrowFocusedPane()
		return
	}
	if message.Delta < 0 {
		program.model.ShrinkFocusedPane()
	}
}

func (program *Program) applyRepeatActionsPopupSearch(message MsgRepeatActionsPopupSearch) {
	program.clearPendingSelectionPrefix()
	if !program.model.ActionsPopupVisible() || program.model.ActionsPopupSearchActive() {
		return
	}
	if strings.TrimSpace(program.model.ActionsPopupSearchQuery()) == "" {
		return
	}
	program.clearActionsPopupPendingConfirmation()
	if !program.model.followActionsPopupSearchMatch(searchMatchChooserForDirection(message.Direction)) {
		return
	}
	program.actionsPopupWidget.errorMessage = ""
}

func (program *Program) setSideSearchSelection(focus Focus, index int) {
	switch focus {
	case FocusNotificationsView:
		program.model.SelectNotificationIndex(index)
	default:
		program.model.SelectUserIndex(index)
	}
}

func (program *Program) applyRepeatSideSearch(message MsgRepeatSideSearch) {
	if program.reviewModeActive() || program.model.Focus() != message.Focus {
		return
	}
	query, matchIndexes, selectedIndex := program.sideSearchState(message.Focus)
	if strings.TrimSpace(query) == "" {
		return
	}
	matchIndex := searchMatchChooserForDirection(message.Direction)(matchIndexes, selectedIndex)
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return
	}
	program.setSideSearchSelection(message.Focus, matchIndexes[matchIndex])
}

func (program *Program) applyRepeatPullRequestSearch(message MsgRepeatPullRequestSearch) {
	if program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView {
		return
	}

	tab := program.model.ActivePullRequestTab()
	query := program.model.PullRequestSearchQuery(tab)
	if strings.TrimSpace(query) == "" {
		return
	}

	matchIndexes := program.model.visiblePullRequestIndexes(tab)
	matchIndex := searchMatchChooserForDirection(message.Direction)(matchIndexes, program.model.SelectedPullRequestIndex(tab))
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return
	}
	program.model.SelectPullRequestIndex(tab, matchIndexes[matchIndex])
}

func (program *Program) followSubmittedPullRequestSearch(tab PullRequestTab, startIndex int) {
	if program.reviewModeActive() {
		return
	}

	query := program.model.PullRequestSearchQuery(tab)
	if strings.TrimSpace(query) == "" {
		return
	}

	matchIndexes := program.model.visiblePullRequestIndexes(tab)
	matchIndex := searchMatchIndexAtOrAfter(matchIndexes, startIndex)
	if matchIndex < 0 || matchIndex >= len(matchIndexes) {
		return
	}

	program.model.SelectPullRequestIndex(tab, matchIndexes[matchIndex])
}

func (program *Program) applyRepeatReviewFileTreeSearch(message MsgRepeatReviewFileTreeSearch) {
	if !program.reviewModeActive() || program.model.Focus() != FocusPullRequestsView {
		return
	}
	query := program.model.ReviewTreeSearchQuery()
	if strings.TrimSpace(query) == "" {
		return
	}
	program.followReviewFileTreeSearch(query, searchMatchChooserForDirection(message.Direction))
}

func (program *Program) applyMoveReviewSelection(message MsgMoveReviewSelection) {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() || program.model.Focus() != FocusPullRequestsView {
		return
	}
	program.adjustReviewSessionSelection(message.Delta)
}

func (program *Program) applyMoveReviewSelectionToTop() {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() || program.model.Focus() != FocusPullRequestsView {
		return
	}
	program.moveReviewSessionSelectionToTop()
}

func (program *Program) applyMoveReviewSelectionToBottom() {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() || program.model.Focus() != FocusPullRequestsView {
		return
	}
	program.moveReviewSessionSelectionToBottom()
}

func (program *Program) applySearchWordUnderCursor(message MsgSearchWordUnderCursor) {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() {
		return
	}

	actualView := program.resolveView(program.gui, message.View, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	query, ok := document.wordAt(program.detailState.viewState.cursor)
	if !ok {
		return
	}

	inputContext := program.inputContext()
	program.detailState.viewState.clearPendingPrefix()
	if inputContext.IsReviewContext() && inputContext.ActiveView.Focus == FocusDetailView {
		program.model.ClearReviewTreeSearchQuery()
	}
	program.model.StartSearch()
	program.applySearchDraftChanged(MsgSearchDraftChanged{Query: query})
	program.model.SubmitSearch()
	program.searchWidget.detailReversed = message.Reverse
	program.searchWidget.editor = nil
	if message.Reverse {
		_ = program.followReverseDetailSearch(program.gui)
		return
	}
	_ = program.followSubmittedDetailSearch(program.gui)
}
