package githubcli

import (
	"errors"
	"strconv"
)

var ErrEmptyPullRequestReviewBody = errors.New("empty pull request review body")

func (client *Client) ApprovePullRequest(repository string, number int) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	result, err := client.runner.Run(ghBinaryName, "pr", "review", strconv.Itoa(number), "-R", trimmedRepository, "--approve")
	if err != nil {
		return classifyCommandError("gh pr review", err, result.Stderr)
	}

	return nil
}

func (client *Client) ReviewPullRequestWithComment(repository string, number int, body string) error {
	return client.submitPullRequestReviewWithBody(repository, number, body, "--comment")
}

func (client *Client) RequestChangesOnPullRequest(repository string, number int, body string) error {
	return client.submitPullRequestReviewWithBody(repository, number, body, "--request-changes")
}

func (client *Client) submitPullRequestReviewWithBody(repository string, number int, body string, reviewArgument string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestReviewBody); err != nil {
		return err
	}

	result, err := client.runner.RunWithInput(ghBinaryName, []byte(body), "pr", "review", strconv.Itoa(number), "-R", trimmedRepository, reviewArgument, "--body-file", "-")
	if err != nil {
		return classifyCommandError("gh pr review", err, result.Stderr)
	}

	return nil
}
