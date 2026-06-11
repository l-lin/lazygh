package tui

import (
	"fmt"
	"strings"

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

func shortRepositoryLabelFromNameWithOwner(repository string) string {
	return valueOrDash(githubdomain.RepositoryShortName(githubdomain.RepositoryRef{NameWithOwner: repository}))
}

func notificationListReference(notification githubdomain.Notification) string {
	if summary, ok := notification.PullRequestSummary(); ok {
		return pullRequestListReference(summary.Repository, summary.Number)
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		return fmt.Sprintf("%s#%d", shortRepositoryLabelFromNameWithOwner(repository), number)
	}
	if repository, _, ok := notification.ReleaseIdentity(); ok {
		return shortRepositoryLabelFromNameWithOwner(repository)
	}
	return shortRepositoryLabel(notification.Repository)
}

func notificationDetailReference(notification githubdomain.Notification) string {
	if summary, ok := notification.PullRequestSummary(); ok {
		return fmt.Sprintf("%s#%d", strings.TrimSpace(summary.Repository.NameWithOwner), summary.Number)
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		return fmt.Sprintf("%s#%d", strings.TrimSpace(repository), number)
	}
	if repository, _, ok := notification.ReleaseIdentity(); ok {
		return strings.TrimSpace(repository)
	}
	return strings.TrimSpace(notification.Repository.NameWithOwner)
}
