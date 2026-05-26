package tui

import "github.com/jesseduffield/gocui"

func (program *Program) mutateDetailViewStateForYankMotion(gui *gocui.Gui, view *gocui.View, selectionKind detailYankMotionSelectionKind, mutate func(detailDocument, int)) error {
	program.clearPendingSelectionPrefix()
	actualView := program.resolveView(gui, view, viewDetailName)
	viewportHeight := viewPageSize(actualView)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewportHeight)
	snapshot := newDetailYankSnapshot(program.detailState.viewState)
	pendingYank := program.detailState.viewState.hasPendingYank()
	mutate(document, viewportHeight)
	program.syncDetailViewState(document, viewportHeight)
	if pendingYank {
		program.finishPendingYank(document, &program.detailState.viewState, snapshot, selectionKind)
	}
	program.syncActionsPopupSearch()
	return program.refreshShell(gui)
}

func (program *Program) mutatePullRequestBuildRunPopupViewState(gui *gocui.Gui, view *gocui.View, mutate func(*detailViewState, detailDocument, int)) error {
	if err := program.mutatePullRequestBuildRunPopupViewStateWithoutRefresh(gui, view, mutate); err != nil {
		return err
	}
	return program.refreshShell(gui)
}

func (program *Program) mutatePullRequestBuildRunPopupViewStateWithoutRefresh(gui *gocui.Gui, view *gocui.View, mutate func(*detailViewState, detailDocument, int)) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)
	mutate(&popup.viewState, document, viewportHeight)
	popup.viewState.sync(document, viewportHeight)
	return nil
}

func (program *Program) mutatePullRequestBuildRunPopupViewStateForYankMotion(gui *gocui.Gui, view *gocui.View, selectionKind detailYankMotionSelectionKind, mutate func(*detailViewState, detailDocument, int)) error {
	popup := program.pullRequestBuildRunPopup
	if popup == nil {
		return nil
	}

	actualView := program.resolveView(gui, view, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	viewportHeight := viewPageSize(actualView)
	popup.viewState.sync(document, viewportHeight)
	snapshot := newDetailYankSnapshot(popup.viewState)
	pendingYank := popup.viewState.hasPendingYank()
	mutate(&popup.viewState, document, viewportHeight)
	popup.viewState.sync(document, viewportHeight)
	if pendingYank {
		program.finishPendingYank(document, &popup.viewState, snapshot, selectionKind)
	}
	return program.refreshShell(gui)
}
