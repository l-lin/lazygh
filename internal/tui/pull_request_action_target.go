package tui

import (
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

type pullRequestActionTarget struct {
	repository string
	number     int
	title      string
	body       string
}

func (program *Program) currentPullRequestSummary() (githubcli.PullRequest, bool) {
	if program.reviewSession.active {
		return program.reviewSession.summary, true
	}

	switch program.model.Focus() {
	case FocusPullRequestsView:
		return program.model.SelectedPullRequestSummary()
	case FocusDetailView:
		return program.selectedPullRequestSummaryForDetail()
	default:
		return githubcli.PullRequest{}, false
	}
}

func (program *Program) selectedPullRequestActionTarget() (pullRequestActionTarget, bool) {
	summary, ok := program.currentPullRequestSummary()
	if !ok {
		return pullRequestActionTarget{}, false
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return pullRequestActionTarget{}, false
	}

	target := pullRequestActionTarget{
		repository: repository,
		number:     summary.Number,
		title:      strings.TrimSpace(summary.Title),
		body:       strings.TrimSpace(summary.Body),
	}
	if result, ok := program.pullRequestDetailForSummary(summary); ok {
		target.title = firstNonEmpty(result.detail.Title, target.title)
		target.body = firstNonEmpty(result.detail.Body, target.body)
	}

	return target, true
}
