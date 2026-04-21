package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func (program *Program) maybeLoadSelectedPullRequestDetail(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil {
		return
	}

	summary, ok := program.selectedPullRequestSummaryForDetail()
	if !ok {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}
	if program.pullRequestDetailLoadInFlight[key] {
		return
	}
	if _, ok := program.pullRequestDetailCache[key]; ok {
		return
	}

	program.pullRequestDetailLoadInFlight[key] = true
	program.asyncRunner.Go(func() {
		program.loadPullRequestDetail(gui, summary)
	})
}

func (program *Program) loadPullRequestDetail(gui *gocui.Gui, summary githubcli.PullRequest) {
	repository := pullRequestRepositoryName(summary.Repository)
	detail, err := program.githubLoader.GetPullRequestDetail(repository, summary.Number)
	key := pullRequestDetailKey(summary.Repository, summary.Number)

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.pullRequestDetailLoadInFlight, key)
		program.pullRequestDetailCache[key] = pullRequestDetailResult{detail: detail, err: err}
		return program.refreshViews(gui)
	})
}

func (program *Program) selectedPullRequestSummaryForDetail() (githubcli.PullRequest, bool) {
	if program.reviewSession.active {
		summary := program.reviewSession.summary
		if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
			return githubcli.PullRequest{}, false
		}
		return summary, true
	}
	if program.model.currentSideFocus() != FocusPullRequestsView {
		return githubcli.PullRequest{}, false
	}

	summary, ok := program.model.SelectedPullRequestSummary()
	if !ok {
		return githubcli.PullRequest{}, false
	}
	if pullRequestDetailKey(summary.Repository, summary.Number) == "" {
		return githubcli.PullRequest{}, false
	}

	return summary, true
}

func (program *Program) pullRequestDetailForSummary(summary githubcli.PullRequest) (pullRequestDetailResult, bool) {
	result, ok := program.pullRequestDetailCache[pullRequestDetailKey(summary.Repository, summary.Number)]
	return result, ok
}

func (program *Program) currentDetailIdentity() string {
	if program.reviewSession.active {
		return program.reviewSessionDetailIdentity()
	}

	switch program.model.currentSideFocus() {
	case FocusPullRequestsView:
		if summary, ok := program.model.SelectedPullRequestSummary(); ok {
			if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
				return fmt.Sprintf("pr:%s:tab:%d", key, program.activeDetailTab)
			}
		}
		return fmt.Sprintf("pr-state:%d:%d", program.model.ActivePullRequestTab(), program.model.SelectedPullRequestIndex(program.model.ActivePullRequestTab()))
	default:
		return fmt.Sprintf("user:%d", program.model.SelectedUserIndex())
	}
}

func (program *Program) invalidatePullRequestDetail(repository string, number int) {
	delete(program.pullRequestDetailCache, strings.TrimSpace(repository)+fmt.Sprintf("#%d", number))
}

func pullRequestDetailKey(repository githubcli.Repository, number int) string {
	repositoryName := strings.TrimSpace(pullRequestRepositoryName(repository))
	if repositoryName == "" || repositoryName == "-" || number <= 0 {
		return ""
	}

	return fmt.Sprintf("%s#%d", repositoryName, number)
}
