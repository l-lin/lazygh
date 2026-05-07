package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
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

	switch tab {
	case CommentsDetailTab:
		commentCount, ok := program.selectedPullRequestDetailCommentCount()
		if !ok {
			return label
		}
		return fmt.Sprintf("%s (%d)", label, commentCount)
	case CommitsDetailTab:
		commitCount, ok := program.selectedPullRequestDetailCommitCount()
		if !ok {
			return label
		}
		return fmt.Sprintf("%s (%d)", label, commitCount)
	default:
		return label
	}
}

func (program *Program) selectedPullRequestDetailCommentCount() (int, bool) {
	detail, ok := program.selectedPullRequestDetailForTabs()
	if !ok {
		return 0, false
	}
	return pullRequestDetailCommentCount(detail), true
}

func (program *Program) selectedPullRequestDetailCommitCount() (int, bool) {
	detail, ok := program.selectedPullRequestDetailForTabs()
	if !ok {
		return 0, false
	}
	return len(detail.Commits), true
}

func (program *Program) selectedPullRequestDetailForTabs() (githubcli.PullRequestDetail, bool) {
	if program.reviewSession.active || program.model.currentSideFocus() != FocusPullRequestsView {
		return githubcli.PullRequestDetail{}, false
	}
	row, ok := program.model.SelectedPullRequestRow()
	if !ok || row.Summary == nil {
		return githubcli.PullRequestDetail{}, false
	}
	result, ok := program.pullRequestDetailForSummary(*row.Summary)
	if !ok || result.err != nil {
		return githubcli.PullRequestDetail{}, false
	}
	return result.detail, true
}
