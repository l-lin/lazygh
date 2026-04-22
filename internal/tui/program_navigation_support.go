package tui

import "github.com/jesseduffield/gocui"

func (program *Program) handleSelectionChange(gui *gocui.Gui, view *gocui.View, sideChange int, mutateDetail func(detailDocument, int)) error {
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

func (program *Program) refreshViewsIfGUI(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}

	return program.refreshViews(gui)
}

func (program *Program) mutateDetailViewState(gui *gocui.Gui, view *gocui.View, mutate func(detailDocument, int)) error {
	if actualErr := program.mutateDetailViewStateWithoutRefresh(gui, view, mutate); actualErr != nil {
		return actualErr
	}

	return program.refreshDetailView(gui)
}

func (program *Program) mutateDetailViewStateWithoutRefresh(gui *gocui.Gui, view *gocui.View, mutate func(detailDocument, int)) error {
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
