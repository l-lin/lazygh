package tui

import "github.com/jesseduffield/gocui"

func (program *Program) handleSelectionChange(gui *gocui.Gui, view *gocui.View, sideChange int, mutateDetail func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewState(gui, view, mutateDetail)
	}
	if program.actionContext().IsReviewContext() {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.adjustReviewSessionSelection(sideChange)
		return program.refreshViewsIfGUI(gui)
	}

	program.model.adjustSelectionBy(sideChange)
	return nil
}

func (program *Program) handlePageChange(gui *gocui.Gui, view *gocui.View, sideChange int, mutateDetail func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewState(gui, view, mutateDetail)
	}
	if program.actionContext().IsReviewContext() {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.adjustReviewSessionSelection(sideChange)
	} else {
		program.model.adjustSelectionBy(sideChange)
	}

	viewName, selectedVisibleLine, lineCount := program.currentSideListState()
	return program.recenterListSelection(gui, view, viewName, selectedVisibleLine, lineCount)
}

func (program *Program) clearPendingSelectionPrefix() {
	program.pendingSelectionKeySequence.clear()
}

func (program *Program) armOrHandleSelectionKeySequence(target keySequenceTarget, handle func() error) error {
	if target.viewName == "" {
		program.clearPendingSelectionPrefix()
		return nil
	}
	if !program.pendingSelectionKeySequence.armOrConsume(target) {
		return nil
	}

	return handle()
}

func (program *Program) refreshViewsIfGUI(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) applyProjectedScreenState(state ScreenState) {
	program.model.focus = state.ActiveView().Focus
	program.model.lastSideFocus = state.ActiveSideView().Focus

	if state.Mode != ScreenModeBrowser {
		return
	}
	if pullRequestView, ok := state.ViewByNumber(sidePanelPullRequestsViewNumber); ok {
		program.model.activePullRequestTab = PullRequestTab(clampScreenTabIndex(pullRequestView.ActiveTab, len(pullRequestView.Tabs)))
	}
	if mainView, ok := state.ViewByNumber(mainPanelViewNumber); ok && len(mainView.Tabs) > 0 {
		program.activeDetailTab = DetailTab(clampScreenTabIndex(mainView.ActiveTab, len(mainView.Tabs)))
	}
}

func (program *Program) applyModeScreenState(gui *gocui.Gui, state ScreenState) error {
	program.applyProjectedScreenState(state)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) focusPanelViewNumber(gui *gocui.Gui, viewNumber int) error {
	state := program.screenState()
	targetView, ok := state.ViewByNumber(viewNumber)
	if !ok {
		return nil
	}
	if program.model.PaneLayoutSize() == PaneLayoutFullscreen && program.model.FullscreenPane() != targetView.Focus {
		return nil
	}
	return program.applyModeScreenState(gui, state.FocusViewNumber(viewNumber))
}

func (program *Program) resolveView(gui *gocui.Gui, view *gocui.View, fallbackName string) *gocui.View {
	if view != nil {
		return view
	}
	if gui == nil || fallbackName == "" {
		return nil
	}

	actualView, actualErr := gui.View(fallbackName)
	if actualErr != nil {
		return nil
	}
	return actualView
}

func (program *Program) recenterListSelection(gui *gocui.Gui, view *gocui.View, fallbackName string, selectedVisibleLine int, lineCount int) error {
	return program.placeListSelection(gui, view, fallbackName, selectedVisibleLine, lineCount, viewportPlacementCenter)
}

func (program *Program) placeListSelection(gui *gocui.Gui, view *gocui.View, fallbackName string, selectedVisibleLine int, lineCount int, placement viewportPlacement) error {
	if lineCount < 1 {
		return nil
	}

	actualView := program.resolveView(gui, view, fallbackName)
	viewName := fallbackName
	if actualView != nil && actualView.Name() != "" {
		viewName = actualView.Name()
	}
	program.setPendingListViewportPlacement(viewName, placement)
	program.placeListLine(actualView, selectedVisibleLine, lineCount, placement)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) currentSideListState() (string, int, int) {
	if program.actionContext().IsReviewContext() && program.model.Focus() == FocusPullRequestsView {
		if tree, _, ok := program.reviewSessionCurrentTree(); ok && len(tree.Rows) > 0 {
			return viewPullRequestsName, program.reviewSessionSelectedVisibleLine(), len(tree.Rows)
		}

		items := program.reviewSessionFiles()
		return viewPullRequestsName, program.reviewSessionSelectedVisibleLine(), len(items)
	}

	switch program.model.Focus() {
	case FocusPullRequestsView:
		return viewPullRequestsName, program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()), len(program.model.CurrentPullRequests())
	case FocusNotificationsView:
		return viewNotificationsName, program.model.SelectedVisibleNotificationIndex(), len(program.model.VisibleNotifications())
	case FocusUserView:
		return viewUserName, program.model.SelectedVisibleUserIndex(), len(program.model.VisibleUsers())
	default:
		return "", 0, 0
	}
}

func (program *Program) mutateDetailViewState(gui *gocui.Gui, view *gocui.View, mutate func(detailDocument, int)) error {
	if actualErr := program.mutateDetailViewStateWithoutRefresh(gui, view, mutate); actualErr != nil {
		return actualErr
	}

	return program.refreshDetailView(gui)
}

func (program *Program) mutateDetailViewStateWithoutRefresh(gui *gocui.Gui, view *gocui.View, mutate func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	actualView := view
	if actualView == nil && gui != nil {
		if detailView, actualErr := gui.View(viewDetailName); actualErr == nil {
			actualView = detailView
		}
	}

	viewportHeight := viewPageSize(actualView)
	detailDocument := program.currentDetailDocument(actualView)
	program.syncDetailViewState(detailDocument, viewportHeight)
	mutate(detailDocument, viewportHeight)
	program.syncDetailViewState(detailDocument, viewportHeight)
	program.syncActionsPopupSearch()
	return nil
}

func (program *Program) refreshDetailView(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}
	if actualErr := program.refreshExistingView(gui, viewDetailName, program.configureDetailView, program.renderDetailView); actualErr != nil {
		return actualErr
	}

	return program.syncCurrentView(gui)
}

func (program *Program) sideViewCyclingBlocked() bool {
	return program.model.PaneLayoutSize() == PaneLayoutFullscreen || program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) mainPaneActionBlocked() bool {
	return program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) detailTransitionBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) helpToggleBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) selectionChangeBlocked() bool {
	return program.model.SearchActive() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) scrollReadOnlyView(gui *gocui.Gui, view *gocui.View, fallbackName string, delta int) error {
	actualView := program.resolveView(gui, view, fallbackName)
	if actualView == nil {
		return nil
	}

	originX, originY := actualView.Origin()
	maxOriginY := maxInt(0, len(actualView.BufferLines())-viewPageSize(actualView))
	actualView.SetOrigin(originX, clampInt(originY+delta, 0, maxOriginY))
	return nil
}

func viewPageSize(view *gocui.View) int {
	if view == nil {
		return 1
	}

	return view.InnerHeight()
}
