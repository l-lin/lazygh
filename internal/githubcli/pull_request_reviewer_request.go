package githubcli

import (
	"errors"
	"strconv"
	"strings"
)

var ErrMissingPullRequestReviewer = errors.New("missing pull request reviewer")

func (client *Client) RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return err
	}

	trimmedReviewerLogin := strings.TrimSpace(reviewerLogin)
	if trimmedReviewerLogin == "" {
		return ErrMissingPullRequestReviewer
	}

	if _, err := client.runGH(
		"gh pr edit",
		"pr",
		"edit",
		strconv.Itoa(number),
		"-R",
		trimmedRepository,
		"--add-reviewer",
		trimmedReviewerLogin,
	); err != nil {
		return err
	}

	return nil
}
