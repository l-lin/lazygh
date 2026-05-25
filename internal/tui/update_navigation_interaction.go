package tui

import "strings"

func searchMatchChooserForDirection(direction searchRepeatDirection) searchMatchIndexChooser {
	if direction == searchRepeatBackward {
		return searchMatchIndexBefore
	}
	return searchMatchIndexAfter
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
	program.updateActiveSearchDraft(query)
	program.model.SubmitSearch()
	program.searchWidget.detailReversed = message.Reverse
	program.searchWidget.editor = nil
	if message.Reverse {
		_ = program.followReverseDetailSearch(program.gui)
		return
	}
	_ = program.followSubmittedDetailSearch(program.gui)
}
