package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type manualRefreshFeedbackState struct {
	successMessage    string
	pendingOperations int
	failed            bool
}

func (program *Program) beginManualRefresh(successMessage string, pendingOperations int) {
	if program == nil || pendingOperations <= 0 {
		return
	}

	program.feedbackMessage = ""
	program.manualRefreshFeedback = &manualRefreshFeedbackState{
		successMessage:    strings.TrimSpace(successMessage),
		pendingOperations: pendingOperations,
	}
}

func (program *Program) completeManualRefreshOperation(gui *gocui.Gui, err error) {
	if program == nil || program.manualRefreshFeedback == nil {
		return
	}

	state := program.manualRefreshFeedback
	if err != nil {
		if !state.failed {
			program.feedbackMessage = ""
			program.reportError(gui, strings.TrimSpace(normalizeGHCommandError(err).Error()))
		}
		state.failed = true
	}
	if state.pendingOperations > 0 {
		state.pendingOperations--
	}
	if state.pendingOperations > 0 {
		return
	}
	if !state.failed && state.successMessage != "" {
		program.setFeedback(FocusDetailView, state.successMessage)
	}
	program.manualRefreshFeedback = nil
}

func (program *Program) markManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil {
		return false
	}
	if program.manualPullRequestListRefreshPending == nil {
		program.manualPullRequestListRefreshPending = map[PullRequestTab]bool{}
	}
	program.manualPullRequestListRefreshPending[tab] = true
	return true
}

func (program *Program) consumeManualPullRequestListRefresh(tab PullRequestTab) bool {
	if program == nil || program.manualPullRequestListRefreshPending == nil {
		return false
	}
	pending := program.manualPullRequestListRefreshPending[tab]
	delete(program.manualPullRequestListRefreshPending, tab)
	return pending
}

func (program *Program) markManualPullRequestDetailRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}
	if program.manualPullRequestDetailRefreshPending == nil {
		program.manualPullRequestDetailRefreshPending = map[string]bool{}
	}
	if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
		program.manualPullRequestDetailRefreshPending[key] = true
		return true
	}
	return false
}

func (program *Program) consumeManualPullRequestDetailRefresh(key string) bool {
	if program == nil || program.manualPullRequestDetailRefreshPending == nil || key == "" {
		return false
	}
	pending := program.manualPullRequestDetailRefreshPending[key]
	delete(program.manualPullRequestDetailRefreshPending, key)
	return pending
}

func (program *Program) markManualPullRequestDiffRefresh(summary githubdomain.PullRequest) bool {
	if program == nil {
		return false
	}
	if program.manualPullRequestDiffRefreshPending == nil {
		program.manualPullRequestDiffRefreshPending = map[string]bool{}
	}
	if key := pullRequestDetailKey(summary.Repository, summary.Number); key != "" {
		program.manualPullRequestDiffRefreshPending[key] = true
		return true
	}
	return false
}

func (program *Program) consumeManualPullRequestDiffRefresh(key string) bool {
	if program == nil || program.manualPullRequestDiffRefreshPending == nil || key == "" {
		return false
	}
	pending := program.manualPullRequestDiffRefreshPending[key]
	delete(program.manualPullRequestDiffRefreshPending, key)
	return pending
}

func (program *Program) markPullRequestDetailNeedsRefresh(summary githubdomain.PullRequest) {
	if program == nil {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	result, ok := program.pullRequestDetailCache[key]
	if !ok || result.err != nil {
		result = pullRequestDetailResult{detail: program.optimisticPullRequestDetailSeed(summary)}
	}
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	result.err = nil
	program.pullRequestDetailCache[key] = result
	delete(program.pullRequestDetailLoadInFlight, key)
	program.invalidatePullRequestDetailDocumentCache()
}

func (program *Program) markPullRequestDiffNeedsRefresh(summary githubdomain.PullRequest) {
	if program == nil {
		return
	}

	key := pullRequestDetailKey(summary.Repository, summary.Number)
	if key == "" {
		return
	}

	result, ok := program.pullRequestDiffCache[key]
	if !ok || result.err != nil {
		return
	}
	result.sourceUpdatedAt = ""
	result.needsRefresh = true
	result.err = nil
	program.pullRequestDiffCache[key] = result
	delete(program.pullRequestDiffLoadInFlight, key)
	program.invalidateReviewDiffRenderCache()
	program.invalidatePullRequestDetailDocumentCache()
}
