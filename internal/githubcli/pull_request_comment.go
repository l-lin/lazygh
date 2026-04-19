package githubcli

import (
	"errors"
	"strconv"
)

var (
	ErrMissingPullRequestIdentity = errors.New("missing pull request identity")
	ErrEmptyPullRequestComment    = errors.New("empty pull request comment")
)

func (client *Client) CommentOnPullRequest(repository string, number int, body string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestComment); err != nil {
		return err
	}

	result, err := client.runner.RunWithInput(ghBinaryName, []byte(body), "pr", "comment", strconv.Itoa(number), "-R", trimmedRepository, "--body-file", "-")
	if err != nil {
		return classifyCommandError("gh pr comment", err, result.Stderr)
	}

	return nil
}
