package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
)

func (program *Program) nextDetailTab(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewSession.active {
		return program.handleReviewFileMotionPrefix(gui, view, reviewNavigationForward)
	}
	if program.helpVisible || program.model.SearchActive() || !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.activeDetailTab = DetailTab((int(program.activeDetailTab) + 1) % len(browserDetailTabs))

	return program.refreshViews(gui)
}

func (program *Program) previousDetailTab(gui *gocui.Gui, view *gocui.View) error {
	if program.reviewSession.active {
		return program.handleReviewFileMotionPrefix(gui, view, reviewNavigationBackward)
	}

	if program.helpVisible || program.model.SearchActive() || !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	program.detailViewState.clearPendingPrefix()
	program.activeDetailTab = DetailTab((int(program.activeDetailTab) + len(browserDetailTabs) - 1) % len(browserDetailTabs))
	return program.refreshViews(gui)
}

func (program *Program) shouldShowPullRequestDetailTabs() bool {
	if program.reviewSession.active || program.model.currentSideFocus() != FocusPullRequestsView {
		return false
	}

	row, ok := program.model.SelectedPullRequestRow()
	return ok && row.Summary != nil
}

var browserDetailTabs = []DetailTab{DescriptionDetailTab, CommentsDetailTab, CommitsDetailTab, ChangesDetailTab}

func (program *Program) detailTabLabels() []string {
	labels := make([]string, 0, len(browserDetailTabs))
	for _, tab := range browserDetailTabs {
		labels = append(labels, program.detailTabLabel(tab))
	}
	return labels
}

func (program *Program) detailTabLabel(tab DetailTab) string {
	label := tab.Label()
	if tab != CommentsDetailTab {
		return label
	}
	commentCount, ok := program.selectedPullRequestDetailCommentCount()
	if !ok {
		return label
	}
	return fmt.Sprintf("%s (%d)", label, commentCount)
}

func (program *Program) selectedPullRequestDetailCommentCount() (int, bool) {
	if program.reviewSession.active || program.model.currentSideFocus() != FocusPullRequestsView {
		return 0, false
	}
	row, ok := program.model.SelectedPullRequestRow()
	if !ok || row.Summary == nil {
		return 0, false
	}
	result, ok := program.pullRequestDetailForSummary(*row.Summary)
	if !ok || result.err != nil {
		return 0, false
	}
	return pullRequestDetailCommentCount(result.detail), true
}
