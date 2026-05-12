package githubcli

import (
	"errors"
	"strconv"
)

var ErrEmptyPullRequestReviewBody = errors.New("empty pull request review body")

func (client *ReviewService) ApprovePullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "review", strconv.Itoa(number), "-R", trimmedRepository, "--approve")); err != nil {
		return err
	}

	return nil
}

func (client *ReviewService) ReviewPullRequestWithComment(repository string, number int, body string) error {
	return client.submitPullRequestReviewWithBody(repository, number, body, "--comment")
}

func (client *ReviewService) RequestChangesOnPullRequest(repository string, number int, body string) error {
	return client.submitPullRequestReviewWithBody(repository, number, body, "--request-changes")
}

func (client *ReviewService) submitPullRequestReviewWithBody(repository string, number int, body string, reviewArgument string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestReviewBody); err != nil {
		return err
	}

	if _, err := client.execute(rawCommandWithInput([]byte(body), "pr", "review", strconv.Itoa(number), "-R", trimmedRepository, reviewArgument, "--body-file", "-")); err != nil {
		return err
	}

	return nil
}
