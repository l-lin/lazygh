package githubcli

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

var (
	ErrMissingPullRequestBuildLink = errors.New("missing pull request build link")
	ErrInvalidPullRequestBuildLink = errors.New("invalid pull request build link")

	pullRequestBuildRunIDPattern      = regexp.MustCompile(`(?i)(?:^|/)actions/runs/(\d+)(?:/|$)`)
	pullRequestBuildRunAttemptPattern = regexp.MustCompile(`(?i)/attempts/(\d+)(?:/|$)`)
)

type pullRequestBuildRunReference struct {
	id      string
	attempt int
}

func (client *Client) GetPullRequestBuildRun(repository string, check PullRequestStatusCheck) (string, error) {
	args, err := pullRequestBuildRunCommandArguments(repository, check)
	if err != nil {
		return "", err
	}

	result, err := client.runGH("gh run view", args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(result.Stdout)), nil
}

func FormatPullRequestBuildRunCommand(repository string, check PullRequestStatusCheck) string {
	args, err := pullRequestBuildRunCommandArguments(repository, check)
	if err != nil {
		return appconfig.FormatGHCommand([]string{"run", "view"})
	}
	return appconfig.FormatGHCommand(args)
}

func pullRequestBuildRunCommandArguments(repository string, check PullRequestStatusCheck) ([]string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return nil, ErrMissingPullRequestIdentity
	}

	reference, err := pullRequestBuildRunReferenceFromLink(check.Link)
	if err != nil {
		return nil, err
	}

	args := []string{"run", "view", reference.id, "-R", trimmedRepository}
	if reference.attempt > 0 {
		args = append(args, "--attempt", strconv.Itoa(reference.attempt))
	}
	args = append(args, "--verbose")
	return args, nil
}

func pullRequestBuildRunReferenceFromLink(raw string) (pullRequestBuildRunReference, error) {
	trimmedLink := strings.TrimSpace(raw)
	if trimmedLink == "" {
		return pullRequestBuildRunReference{}, ErrMissingPullRequestBuildLink
	}

	path := trimmedLink
	if parsedURL, err := url.Parse(trimmedLink); err == nil && strings.TrimSpace(parsedURL.Path) != "" {
		path = parsedURL.Path
	}

	matches := pullRequestBuildRunIDPattern.FindStringSubmatch(path)
	if len(matches) < 2 {
		return pullRequestBuildRunReference{}, ErrInvalidPullRequestBuildLink
	}

	reference := pullRequestBuildRunReference{id: strings.TrimSpace(matches[1])}
	if attemptMatches := pullRequestBuildRunAttemptPattern.FindStringSubmatch(path); len(attemptMatches) >= 2 {
		if attempt, err := strconv.Atoi(strings.TrimSpace(attemptMatches[1])); err == nil && attempt > 0 {
			reference.attempt = attempt
		}
	}
	return reference, nil
}
