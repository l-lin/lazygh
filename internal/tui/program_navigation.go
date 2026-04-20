package tui

import "github.com/jesseduffield/gocui"

func (program *Program) quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (program *Program) nextSideView(gui *gocui.Gui, _ *gocui.View) error {
	if program.sideViewCyclingBlocked() {
		return nil
	}

	program.model.NextSideView()
	return program.syncCurrentView(gui)
}

func (program *Program) previousSideView(gui *gocui.Gui, _ *gocui.View) error {
	if program.sideViewCyclingBlocked() {
		return nil
	}

	program.model.PreviousSideView()
	return program.syncCurrentView(gui)
}

func (program *Program) moveSelectionDown(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
			program.detailViewState.moveDown(document, viewportHeight)
		})
	}

	program.model.MoveSelectionDown()
	return nil
}

func (program *Program) moveSelectionUp(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
			program.detailViewState.moveUp(document, viewportHeight)
		})
	}

	program.model.MoveSelectionUp()
	return nil
}

func (program *Program) pageDown(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
			program.detailViewState.pageDown(document, viewportHeight)
		})
	}

	program.model.PageDown(viewPageSize(view))
	return nil
}

func (program *Program) pageUp(gui *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
			program.detailViewState.pageUp(document, viewportHeight)
		})
	}

	program.model.PageUp(viewPageSize(view))
	return nil
}

func (program *Program) moveDetailCursorLeft(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveLeft(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorRight(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveRight(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowStart(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToRowStart(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToRowEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToRowEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToTop(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.handleGoToTopPrefix(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToBottom(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToBottom(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToNextWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToNextWord(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToWordEnd(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToWordEnd(document, viewportHeight)
	})
}

func (program *Program) moveDetailCursorToPreviousWord(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.moveToPreviousWord(document, viewportHeight)
	})
}

func (program *Program) enterDetailVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.enterVisualMode()
		program.detailViewState.sync(document, viewportHeight)
	})
}

func (program *Program) enterDetailLineVisualMode(gui *gocui.Gui, view *gocui.View) error {
	return program.mutateDetailViewState(gui, view, func(document detailDocument, viewportHeight int) {
		program.detailViewState.enterLineVisualMode(document)
		program.detailViewState.sync(document, viewportHeight)
	})
}

func (program *Program) nextPullRequestTab(gui *gocui.Gui, _ *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}

	program.model.NextPullRequestTab()
	program.reloadActivePullRequestsTab(gui)
	return nil
}

func (program *Program) previousPullRequestTab(gui *gocui.Gui, _ *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}

	program.model.PreviousPullRequestTab()
	program.reloadActivePullRequestsTab(gui)
	return nil
}

func (program *Program) focusDetailView(gui *gocui.Gui, _ *gocui.View) error {
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.model.FocusDetailView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusUserView(gui *gocui.Gui, _ *gocui.View) error {
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.model.FocusUserView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusPullRequestsView(gui *gocui.Gui, _ *gocui.View) error {
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.model.FocusPullRequestsView()
	return program.syncCurrentView(gui)
}

func (program *Program) openDetail(gui *gocui.Gui, _ *gocui.View) error {
	if program.detailTransitionBlocked() {
		return nil
	}

	program.model.OpenDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) closeDetail(gui *gocui.Gui, _ *gocui.View) error {
	if program.detailTransitionBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView && program.detailViewState.mode.isVisual() {
		program.detailViewState.exitVisualMode()
		return program.refreshDetailView(gui)
	}

	program.detailViewState.clearPendingPrefix()
	program.model.CloseDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) openSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.model.StartSearch()
	program.searchEditor = newLineEditor("")
	return program.layout(gui)
}

func (program *Program) submitSearch(gui *gocui.Gui, _ *gocui.View) error {
	target := program.model.SearchTarget()
	program.model.SubmitSearch()
	program.searchEditor = nil

	actualErr := gui.DeleteView(viewSearchName)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return actualErr
	}
	if target == FocusDetailView {
		if actualErr := program.followSubmittedDetailSearch(gui); actualErr != nil {
			return actualErr
		}
	}

	return program.refreshViews(gui)
}

func (program *Program) cancelSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.model.CancelSearch()
	return program.closeSearch(gui)
}

func (program *Program) closeSearch(gui *gocui.Gui) error {
	program.searchEditor = nil

	actualErr := gui.DeleteView(viewSearchName)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return actualErr
	}

	return program.refreshViews(gui)
}

func (program *Program) toggleHelp(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpToggleBlocked() {
		return nil
	}

	program.helpVisible = !program.helpVisible
	if !program.helpVisible {
		return program.closeHelp(gui, nil)
	}

	return program.layout(gui)
}

func (program *Program) closeHelp(gui *gocui.Gui, _ *gocui.View) error {
	program.helpVisible = false
	actualErr := gui.DeleteView(viewHelpName)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		return actualErr
	}

	return program.syncCurrentView(gui)
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
