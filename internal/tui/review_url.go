package tui

import (
	"errors"

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

	pendingReviewID, err := startPendingPullRequestReview(program, summary)
	if err != nil {
		return err
	}

	program.startReviewSession(summary, pendingReviewID)
	return nil
}
