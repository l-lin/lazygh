package githubcli

import (
	"errors"
	"strconv"
)

var ErrEmptyPullRequestTitle = errors.New("empty pull request title")

func (client *PullRequestMutationService) EditPullRequestTitle(repository string, number int, title string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}
	trimmedTitle, err := validateNonEmptyPullRequestField(title, ErrEmptyPullRequestTitle)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommand("pr", "edit", strconv.Itoa(number), "-R", trimmedRepository, "--title", trimmedTitle)); err != nil {
		return err
	}

	return nil
}

func (client *PullRequestMutationService) EditPullRequestDescription(repository string, number int, body string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	if _, err := client.execute(rawCommandWithInput([]byte(body), "pr", "edit", strconv.Itoa(number), "-R", trimmedRepository, "--body-file", "-")); err != nil {
		return err
	}

	return nil
}
