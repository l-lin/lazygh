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

	if _, err := client.execute(rawCommandWithInput([]byte(body), "pr", "comment", strconv.Itoa(number), "-R", trimmedRepository, "--body-file", "-")); err != nil {
		return err
	}

	return nil
}
