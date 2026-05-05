package tui

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

type pullRequestDiffResult struct {
	data reviewDiffData
	err  error
}

func (program *Program) maybeLoadSelectedPullRequestDiff(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil || !program.reviewSession.active {
		return
	}

	summary := program.reviewSession.summary
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" || program.pullRequestDiffLoadInFlight[key] {
		return
	}
	if _, ok := program.pullRequestDiffCache[key]; ok {
		return
	}

	program.pullRequestDiffLoadInFlight[key] = true
	program.asyncRunner.Go(func() {
		program.loadPullRequestDiff(gui, summary)
	})
}

func (program *Program) loadPullRequestDiff(gui *gocui.Gui, summary githubcli.PullRequest) {
	repository := pullRequestRepositoryName(summary.Repository)
	rawDiff, err := program.githubLoader.GetPullRequestDiff(repository, summary.Number)
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	result := pullRequestDiffResult{err: err}
	if err == nil {
		result.data = buildReviewDiffData(rawDiff)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		delete(program.pullRequestDiffLoadInFlight, key)
		program.pullRequestDiffCache[key] = result
		program.invalidateReviewDiffRenderCache()
		program.clampReviewSessionSelection()
		return program.refreshViews(gui)
	})
}

func (program *Program) pullRequestDiffForSummary(summary githubcli.PullRequest) (pullRequestDiffResult, bool) {
	result, ok := program.pullRequestDiffCache[pullRequestDetailKey(summary.Repository, summary.Number)]
	return result, ok
}

func (program *Program) invalidatePullRequestDiff(repository string, number int) {
	delete(program.pullRequestDiffCache, strings.TrimSpace(repository)+fmt.Sprintf("#%d", number))
	program.invalidateReviewDiffRenderCache()
}

func (program *Program) selectedPullRequestDiffLoading() bool {
	if !program.reviewSession.active {
		return false
	}
	return program.pullRequestDiffLoadInFlight[pullRequestDetailKey(program.reviewSession.summary.Repository, program.reviewSession.summary.Number)]
}
