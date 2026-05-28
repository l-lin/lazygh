package tui

import "github.com/jesseduffield/gocui"

func (program *Program) refreshDetailView(gui *gocui.Gui) error {
	if gui == nil {
		return nil
	}
	if actualErr := program.refreshExistingView(gui, viewDetailName, program.configureDetailView, program.renderDetailView); actualErr != nil {
		return actualErr
	}
	return program.syncShellState(gui)
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

func (program *Program) reloadActivePullRequestsTab(gui *gocui.Gui) {
	if gui == nil {
		return
	}

	program.executeWorkflowPlan(gui, program.pullRequestListReloadPlan(program.model.ActivePullRequestTab()))
	_ = program.afterStateChange(gui)
}

func (program *Program) maybeLoadSelectedPullRequestDetail(gui *gocui.Gui) {
	program.executeWorkflowPlan(gui, program.selectedPullRequestDetailLoadPlan())
}

func (program *Program) maybeLoadSelectedPullRequestDiff(gui *gocui.Gui) {
	program.executeWorkflowPlan(gui, program.selectedPullRequestDiffLoadPlan())
}
