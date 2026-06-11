package tui

import (
	"fmt"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func shortRepositoryLabel(repository any) string {
	repositoryRef, ok := toDomainRepository(repository)
	if !ok {
		return "-"
	}
	return valueOrDash(githubdomain.RepositoryShortName(repositoryRef))
}

func pullRequestListReference(repository any, number int) string {
	return fmt.Sprintf("%s#%d", shortRepositoryLabel(repository), number)
}
