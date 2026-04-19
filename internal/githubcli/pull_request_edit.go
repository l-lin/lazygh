package githubcli

import (
	"errors"
	"strconv"
)

var ErrEmptyPullRequestTitle = errors.New("empty pull request title")

func (client *Client) EditPullRequestTitle(repository string, number int, title string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}
	trimmedTitle, err := validateNonEmptyPullRequestField(title, ErrEmptyPullRequestTitle)
	if err != nil {
		return err
	}

	result, err := client.runner.Run(ghBinaryName, "pr", "edit", strconv.Itoa(number), "-R", trimmedRepository, "--title", trimmedTitle)
	if err != nil {
		return classifyCommandError("gh pr edit", err, result.Stderr)
	}

	return nil
}

func (client *Client) EditPullRequestDescription(repository string, number int, body string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	result, err := client.runner.RunWithInput(ghBinaryName, []byte(body), "pr", "edit", strconv.Itoa(number), "-R", trimmedRepository, "--body-file", "-")
	if err != nil {
		return classifyCommandError("gh pr edit", err, result.Stderr)
	}

	return nil
}
