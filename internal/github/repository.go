package github

import "strings"

func RepositoryShortName(repository RepositoryRef) string {
	if repositoryName := strings.TrimSpace(repository.Name); repositoryName != "" {
		return repositoryName
	}

	segments := strings.Split(strings.TrimSpace(repository.NameWithOwner), "/")
	for index := len(segments) - 1; index >= 0; index-- {
		segment := strings.TrimSpace(segments[index])
		if segment != "" {
			return segment
		}
	}

	return ""
}
