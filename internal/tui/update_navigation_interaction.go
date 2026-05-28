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

func (program *Program) applySearchEditorInputRequested(message MsgSearchEditorInputRequested) {
	if !program.searchWidget.hasEditor() {
		program.searchWidget.openEditor(program.model.SearchDraft())
	}
	if !program.searchWidget.editor.ApplyIntent(message.Intent) {
		return
	}
	program.applySearchDraftChanged(MsgSearchDraftChanged{Query: program.searchWidget.editor.Text()})
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
	program.searchWidget.openEditor(message.Query)
}

func (program *Program) applySubmitReviewFileTreeSearch() {
	if !program.activeSearchIsReviewFileTreeSearch() {
		return
	}
	program.model.SubmitSearch()
	program.followSubmittedReviewFileTreeSearch(program.model.ReviewTreeSearchQuery())
	program.searchWidget.clearEditor()
}

func (program *Program) applyCancelReviewFileTreeSearch() {
	if !program.activeSearchIsReviewFileTreeSearch() {
		return
	}
	program.model.CancelSearch()
	program.searchWidget.clearEditor()
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
	program.resetDetailViewState()
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

func (program *Program) applyLineNavigationRequested(message MsgLineNavigationRequested) []Cmd {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() || message.Delta == 0 {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return detailMotionCommandsForLineDelta(message.Delta)
	}
	if program.actionContext().IsReviewContext() {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.adjustReviewSessionSelection(message.Delta)
		return nil
	}
	program.model.adjustSelectionBy(message.Delta)
	return nil
}

func (program *Program) applyPageNavigationRequested(message MsgPageNavigationRequested) []Cmd {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		if operation, ok := detailMotionOperationForPageNavigationKind(message.Kind); ok {
			return []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: operation}}
		}
		return nil
	}
	return []Cmd{pageNavigationCmd{Kind: message.Kind}}
}

func (program *Program) applyPageNavigationResolved(message MsgPageNavigationResolved) []Cmd {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}

	delta := pageNavigationDelta(message.Kind, message.PageSize)
	if delta == 0 {
		return nil
	}
	if program.actionContext().IsReviewContext() {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.adjustReviewSessionSelection(delta)
		return []Cmd{sideListViewportCmd{Placement: viewportPlacementCenter}}
	}

	program.model.adjustSelectionBy(delta)
	return []Cmd{sideListViewportCmd{Placement: viewportPlacementCenter}}
}

func (program *Program) applySideListViewportRequested(message MsgSideListViewportRequested) []Cmd {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	return []Cmd{sideListViewportCmd{Placement: message.Placement}}
}

func (program *Program) applyDetailViewportRequested(message MsgDetailViewportRequested) []Cmd {
	if !program.model.PaneVisible(FocusDetailView) {
		return nil
	}
	return []Cmd{detailViewportCmd{Operation: message.Operation}}
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

func (program *Program) applySearchWordUnderCursor(message MsgSearchWordUnderCursor) []Cmd {
	program.clearPendingSelectionPrefix()
	if program.mainPaneActionBlocked() {
		return nil
	}

	return []Cmd{resolveDetailSearchWordCmd{Reverse: message.Reverse}}
}

func (program *Program) applyRepeatDetailSearchRequested(message MsgRepeatDetailSearchRequested) []Cmd {
	if program.model.Focus() != FocusDetailView || program.detailState.viewState.mode != detailNormalMode {
		return nil
	}
	if strings.TrimSpace(program.model.DetailSearchQuery()) == "" {
		return nil
	}

	reverse := message.Direction == searchRepeatBackward
	if program.searchWidget.detailReversed {
		reverse = !reverse
	}
	return []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationRepeatSearch, Reverse: reverse}}
}

func (program *Program) applyDetailSearchWordResolved(message MsgDetailSearchWordResolved) []Cmd {
	program.detailState = program.detailState.synced(program.currentDetailIdentity(), message.Document, message.ViewportHeight, program.model.DetailSearchQuery())
	query, ok := message.Document.wordAt(program.detailState.viewState.cursor)
	if !ok {
		return nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
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
	program.searchWidget.clearEditor()
	return []Cmd{detailMotionCmd{Target: detailMotionTargetDetail, Operation: detailMotionOperationFollowSubmittedSearch, Reverse: message.Reverse}}
}

func detailMotionCommandsForLineDelta(delta int) []Cmd {
	if delta == 0 {
		return nil
	}

	operation := detailMotionOperationMoveDown
	steps := delta
	if delta < 0 {
		operation = detailMotionOperationMoveUp
		steps = -delta
	}

	actual := make([]Cmd, 0, steps)
	for range steps {
		actual = append(actual, detailMotionCmd{Target: detailMotionTargetDetail, Operation: operation})
	}
	return actual
}

func detailMotionOperationForPageNavigationKind(kind pageNavigationKind) (detailMotionOperation, bool) {
	switch kind {
	case pageNavigationKindHalfDown:
		return detailMotionOperationPageDown, true
	case pageNavigationKindHalfUp:
		return detailMotionOperationPageUp, true
	case pageNavigationKindFullDown:
		return detailMotionOperationFullPageDown, true
	case pageNavigationKindFullUp:
		return detailMotionOperationFullPageUp, true
	default:
		return detailMotionOperationMoveDown, false
	}
}
