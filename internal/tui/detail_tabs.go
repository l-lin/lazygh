package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) nextDetailTab(gui *gocui.Gui, view *gocui.View) error {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return nil
	}
	if program.overlayState.helpVisible || program.model.SearchActive() || !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	return program.dispatch(gui, MsgAdvanceDetailTab{Delta: 1})
}

func (program *Program) previousDetailTab(gui *gocui.Gui, view *gocui.View) error {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return nil
	}

	if program.overlayState.helpVisible || program.model.SearchActive() || !program.shouldShowPullRequestDetailTabs() {
		return nil
	}

	return program.dispatch(gui, MsgAdvanceDetailTab{Delta: -1})
}

func (program *Program) shouldShowPullRequestDetailTabs() bool {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return false
	}
	return program.mainViewResolver().ContentKind == MainContentKindPullRequestDetail
}

var browserDetailTabs = []DetailTab{DescriptionDetailTab, CommentsDetailTab, CommitsDetailTab, ChangesDetailTab}

func (program *Program) visibleDetailTabs() []DetailTab {
	visibleTabs := append([]DetailTab(nil), browserDetailTabs...)
	pullRequestKey, ok := program.currentPullRequestDetailKeyForVisibleTabs()
	if !ok {
		return visibleTabs
	}
	if program.detailState.commitDiffTab.visibleForPullRequestKey(pullRequestKey) {
		visibleTabs = append(visibleTabs, CommitChangesDetailTab)
	}
	return visibleTabs
}

func (program *Program) currentPullRequestSummaryForVisibleTabs() (githubdomain.PullRequest, bool) {
	if program == nil || program.model == nil || program.modeDescriptor().Mode() != ScreenModeBrowser {
		return githubdomain.PullRequest{}, false
	}

	switch program.model.ScreenState().MainViewResolver().SourceView.Focus {
	case FocusPullRequestsView:
		summary, ok := program.model.SelectedPullRequestSummary()
		if !ok {
			return githubdomain.PullRequest{}, false
		}
		return summary, pullRequestDetailKey(summary.Repository, summary.Number) != ""
	case FocusNotificationsView:
		notification, ok := program.model.SelectedNotification()
		if !ok {
			return githubdomain.PullRequest{}, false
		}
		summary, ok := notification.PullRequestSummary()
		if !ok {
			return githubdomain.PullRequest{}, false
		}
		return summary, pullRequestDetailKey(summary.Repository, summary.Number) != ""
	default:
		return githubdomain.PullRequest{}, false
	}
}

func (program *Program) currentPullRequestDetailKeyForVisibleTabs() (string, bool) {
	summary, ok := program.currentPullRequestSummaryForVisibleTabs()
	if !ok {
		return "", false
	}
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	return key, key != ""
}

func (program *Program) activeDetailTabIndex() int {
	return detailTabIndex(program.visibleDetailTabs(), program.detailState.activeTab)
}

func (program *Program) detailTabLabels() []string {
	visibleTabs := program.visibleDetailTabs()
	labels := make([]string, 0, len(visibleTabs))
	for _, tab := range visibleTabs {
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
	case CommitChangesDetailTab:
		if actual := strings.TrimSpace(program.detailState.commitDiffTab.shortLabel); actual != "" {
			return detailChangesIcon + " " + actual
		}
		if actual := shortPullRequestCommitOID(program.detailState.commitDiffTab.commitOID); actual != "" {
			return detailChangesIcon + " " + actual
		}
		return label
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

func (program *Program) selectedPullRequestDetailForTabs() (githubdomain.PullRequestDetail, bool) {
	if program.modeDescriptor().Mode() != ScreenModeBrowser {
		return githubdomain.PullRequestDetail{}, false
	}
	summary, ok := program.currentPullRequestSummaryForVisibleTabs()
	if !ok {
		return githubdomain.PullRequestDetail{}, false
	}
	result, ok := program.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		return githubdomain.PullRequestDetail{}, false
	}
	return result.detail, true
}
