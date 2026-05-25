package tui

import "github.com/jesseduffield/gocui"

func (program *Program) handleSelectionChange(gui *gocui.Gui, view *gocui.View, sideChange int, mutateDetail func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, mutateDetail)
	}
	if program.actionContext().IsReviewContext() {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.adjustReviewSessionSelection(sideChange)
		return program.refreshShell(gui)
	}

	program.applyMoveSideSelection(MsgMoveSideSelection{Delta: sideChange})
	return nil
}

func (program *Program) handlePageChange(gui *gocui.Gui, view *gocui.View, sideChange int, mutateDetail func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewStateForYankMotion(gui, view, detailYankMotionLinewise, mutateDetail)
	}
	if program.actionContext().IsReviewContext() {
		if program.model.Focus() != FocusPullRequestsView {
			return nil
		}
		program.adjustReviewSessionSelection(sideChange)
	} else {
		program.applyMoveSideSelection(MsgMoveSideSelection{Delta: sideChange})
	}

	viewName, selectedVisibleLine, lineCount := program.currentSideListState()
	return program.recenterListSelection(gui, view, viewName, selectedVisibleLine, lineCount)
}

func (program *Program) clearPendingSelectionPrefix() {
	program.navigationState.pendingSelectionKeySequence.clear()
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
		return viewNotificationsName, program.model.SelectedNotificationIndex(), len(program.model.Notifications())
	case FocusUserView:
		return viewUserName, program.model.SelectedUserIndex(), len(program.model.Users())
	default:
		return "", 0, 0
	}
}

func (program *Program) sideViewCyclingBlocked() bool {
	return program.model.PaneLayoutSize() == PaneLayoutFullscreen || program.overlayState.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.modalEditorVisible() || program.pullRequestBuildRunPopupVisible()
}

func (program *Program) mainPaneActionBlocked() bool {
	return program.overlayState.helpVisible || program.model.SearchActive() || program.model.ActionsPopupVisible() || program.pullRequestBuildRunPopupVisible()
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
