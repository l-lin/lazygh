package githubcli

import (
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func normalizePullRequestIdentity(repository string, number int) (string, error) {
	trimmedRepository, _, err := githubdomain.NormalizePullRequestIdentity(repository, number)
	if err != nil {
		return "", err
	}
	return trimmedRepository, nil
}

func validateNonEmptyPullRequestField(value string, err error) (string, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return "", err
	}
	return trimmedValue, nil
}
