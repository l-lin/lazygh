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

func (program *Program) moveSelectionDown(_ *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailDownLine(view)
		return nil
	}

	program.model.MoveSelectionDown()
	return nil
}

func (program *Program) moveSelectionUp(_ *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailUpLine(view)
		return nil
	}

	program.model.MoveSelectionUp()
	return nil
}

func (program *Program) pageDown(_ *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailDown(view)
		return nil
	}

	program.model.PageDown(viewPageSize(view))
	return nil
}

func (program *Program) pageUp(_ *gocui.Gui, view *gocui.View) error {
	if program.selectionChangeBlocked() {
		return nil
	}
	if program.model.Focus() == FocusDetailView {
		program.scrollDetailUp(view)
		return nil
	}

	program.model.PageUp(viewPageSize(view))
	return nil
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

	program.model.FocusUserView()
	return program.syncCurrentView(gui)
}

func (program *Program) focusPullRequestsView(gui *gocui.Gui, _ *gocui.View) error {
	if program.mainPaneActionBlocked() {
		return nil
	}

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

	program.model.CloseDetail()
	return program.syncCurrentView(gui)
}

func (program *Program) openSearch(gui *gocui.Gui, _ *gocui.View) error {
	if program.mainPaneActionBlocked() {
		return nil
	}

	program.model.StartSearch()
	program.searchEditor = newLineEditor("")
	return program.layout(gui)
}

func (program *Program) submitSearch(gui *gocui.Gui, _ *gocui.View) error {
	program.model.SubmitSearch()
	return program.closeSearch(gui)
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

func (program *Program) scrollDetailDown(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollDown(pageDelta(view.InnerHeight()))
}

func (program *Program) scrollDetailUp(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollUp(pageDelta(view.InnerHeight()))
}

func (program *Program) scrollDetailDownLine(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollDown(1)
}

func (program *Program) scrollDetailUpLine(view *gocui.View) {
	if view == nil {
		return
	}

	view.ScrollUp(1)
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
