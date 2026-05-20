package tui

import (
	"errors"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) OpenReviewByURL(rawURL string) error {
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}

	summary, err := githubdomain.ParsePullRequestURL(rawURL)
	if err != nil {
		return err
	}
	return program.openPullRequestReview(summary)
}

func (program *Program) openPullRequestReview(summary githubdomain.PullRequest) error {
	if !program.hasReviewMutations() {
		return errors.New("github loader is unavailable")
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return errors.New("missing pull request identity")
	}

	pendingReviewID, err := program.reviewMutations.StartPendingPullRequestReview(repository, summary.Number)
	if err != nil {
		return newTransientErrorPopupActionError(err)
	}

	program.startReviewSession(summary, pendingReviewID)
	return nil
}
