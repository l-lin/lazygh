package githubcli

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrMissingPullRequestIdentity = errors.New("missing pull request identity")
	ErrEmptyPullRequestComment    = errors.New("empty pull request comment")
)

func (client *Client) CommentOnPullRequest(repository string, number int, body string) error {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" || number <= 0 {
		return ErrMissingPullRequestIdentity
	}
	if strings.TrimSpace(body) == "" {
		return ErrEmptyPullRequestComment
	}

	result, err := client.runner.RunWithInput(ghBinaryName, []byte(body), "pr", "comment", strconv.Itoa(number), "-R", trimmedRepository, "--body-file", "-")
	if err != nil {
		return classifyCommandError("gh pr comment", err, result.Stderr)
	}

	return nil
}
