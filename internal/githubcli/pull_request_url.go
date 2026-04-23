package githubcli

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var ErrInvalidPullRequestURL = errors.New("invalid GitHub pull request URL")

func ParsePullRequestURL(raw string) (PullRequest, error) {
	trimmedURL := strings.TrimSpace(raw)
	if trimmedURL == "" {
		return PullRequest{}, ErrInvalidPullRequestURL
	}

	if !strings.Contains(trimmedURL, "://") {
		normalizedPrefix := strings.ToLower(trimmedURL)
		if strings.HasPrefix(normalizedPrefix, "github.com/") || strings.HasPrefix(normalizedPrefix, "www.github.com/") {
			trimmedURL = "https://" + trimmedURL
		}
	}

	parsedURL, err := url.Parse(trimmedURL)
	if err != nil {
		return PullRequest{}, ErrInvalidPullRequestURL
	}

	host := strings.ToLower(strings.TrimSpace(parsedURL.Hostname()))
	switch host {
	case "github.com", "www.github.com":
	default:
		return PullRequest{}, ErrInvalidPullRequestURL
	}

	pathSegments := pullRequestURLPathSegments(parsedURL.Path)
	if len(pathSegments) < 4 || pathSegments[2] != "pull" {
		return PullRequest{}, ErrInvalidPullRequestURL
	}

	owner := pathSegments[0]
	repositoryName := pathSegments[1]
	pullRequestNumber, err := strconv.Atoi(pathSegments[3])
	if owner == "" || repositoryName == "" || err != nil || pullRequestNumber <= 0 {
		return PullRequest{}, ErrInvalidPullRequestURL
	}

	repository := strings.TrimSpace(owner) + "/" + strings.TrimSpace(repositoryName)
	return PullRequest{
		Number: pullRequestNumber,
		Repository: Repository{
			Name:          repositoryName,
			NameWithOwner: repository,
		},
		URL: canonicalPullRequestURL(repository, pullRequestNumber),
	}, nil
}

func pullRequestURLPathSegments(path string) []string {
	rawSegments := strings.Split(strings.TrimSpace(path), "/")
	segments := make([]string, 0, len(rawSegments))
	for _, segment := range rawSegments {
		trimmedSegment := strings.TrimSpace(segment)
		if trimmedSegment == "" {
			continue
		}
		segments = append(segments, trimmedSegment)
	}
	return segments
}

func canonicalPullRequestURL(repository string, number int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || number <= 0 {
		return ""
	}
	return "https://github.com/" + trimmedRepository + "/pull/" + strconv.Itoa(number)
}
