package tui

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type pullRequestActionTarget struct {
	repository string
	number     int
	title      string
	body       string
}

func (program *Program) currentPullRequestSummary() (githubdomain.PullRequest, bool) {
	actionContext := program.actionContext()
	if actionContext.IsReviewContext() {
		summary, ok := toDomainPullRequestSummary(program.reviewSession.summary)
		if !ok {
			return githubdomain.PullRequest{}, false
		}
		return summary, true
	}

	switch actionContext.ActiveView.Focus {
	case FocusPullRequestsView:
		return program.model.SelectedPullRequestSummary()
	case FocusDetailView:
		if actionContext.MainView.ContentKind != MainContentKindPullRequestDetail {
			return githubdomain.PullRequest{}, false
		}
		return program.selectedPullRequestSummaryForDetail()
	default:
		return githubdomain.PullRequest{}, false
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
