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
	if program.reviewSession.active {
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
	if program.reviewSession.active {
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

func (program *Program) currentSideViewName() string {
	switch program.model.Focus() {
	case FocusUserView:
		return viewUserName
	case FocusPullRequestsView:
		return viewPullRequestsName
	default:
		return ""
	}
}

func (program *Program) refreshViewsIfGUI(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
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
	if lineCount < 1 {
		return nil
	}

	actualView := program.resolveView(gui, view, fallbackName)
	program.centerListLine(actualView, selectedVisibleLine, lineCount)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) currentSideListState() (string, int, int) {
	if program.reviewSession.active && program.model.Focus() == FocusPullRequestsView {
		if result, ok := program.reviewSessionDiffResult(); ok && result.err == nil && len(result.data.FileTree.Rows) > 0 {
			return viewPullRequestsName, program.reviewSessionSelectedVisibleLine(), len(result.data.FileTree.Rows)
		}

		items := program.reviewSessionFiles()
		return viewPullRequestsName, program.reviewSessionSelectedVisibleLine(), len(items)
	}

	switch program.model.Focus() {
	case FocusPullRequestsView:
		return viewPullRequestsName, program.model.SelectedVisiblePullRequestIndex(program.model.ActivePullRequestTab()), len(program.model.VisiblePullRequests())
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
	return program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible()
}

func (program *Program) mainPaneActionBlocked() bool {
	return program.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible()
}

func (program *Program) detailTransitionBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible()
}

func (program *Program) helpToggleBlocked() bool {
	return program.model.SearchActive() || program.model.ActionsPopupVisible()
}

func (program *Program) selectionChangeBlocked() bool {
	return program.model.SearchActive()
}

func viewPageSize(view *gocui.View) int {
	if view == nil {
		return 1
	}

	return view.InnerHeight()
}
