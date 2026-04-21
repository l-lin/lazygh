package tui

import "github.com/jesseduffield/gocui"

func (program *Program) nextDetailTab(gui *gocui.Gui, _ *gocui.View) error {
	if program.helpVisible || program.model.SearchActive() || !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	switch program.activeDetailTab {
	case CommentsDetailTab:
		program.activeDetailTab = DescriptionDetailTab
	default:
		program.activeDetailTab = CommentsDetailTab
	}

	return program.refreshViews(gui)
}

func (program *Program) previousDetailTab(gui *gocui.Gui, _ *gocui.View) error {
	return program.nextDetailTab(gui, nil)
}

func (program *Program) shouldShowPullRequestDetailTabs() bool {
	if program.reviewSession.active || program.model.currentSideFocus() != FocusPullRequestsView {
		return false
	}

	row, ok := program.model.SelectedPullRequestRow()
	return ok && row.Summary != nil
}

func (program *Program) detailTabLabels() []string {
	return []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label()}
}
