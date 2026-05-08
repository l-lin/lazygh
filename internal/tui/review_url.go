package tui

import (
	"errors"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func (program *Program) OpenReviewByURL(rawURL string) error {
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	summary, err := githubcli.ParsePullRequestURL(rawURL)
	if err != nil {
		return err
	}
	return program.openPullRequestReview(summary)
}

func (program *Program) openPullRequestReview(summary githubcli.PullRequest) error {
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	repository := strings.TrimSpace(pullRequestRepositoryName(summary.Repository))
	if repository == "" || repository == "-" || summary.Number <= 0 {
		return errors.New("missing pull request identity")
	}

	pendingReviewID, err := program.githubLoader.StartPendingPullRequestReview(repository, summary.Number)
	if err != nil {
		return err
	}

	program.startReviewSession(summary, pendingReviewID)
	return nil
}
