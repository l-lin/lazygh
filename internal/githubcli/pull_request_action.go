package githubcli

import "strings"

func normalizePullRequestIdentity(repository string, number int) (string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" || number <= 0 {
		return "", ErrMissingPullRequestIdentity
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
