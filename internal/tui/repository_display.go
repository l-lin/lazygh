package tui

import (
	"fmt"
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func repositoryListLabel(style string, repository any) string {
	repositoryRef, ok := toDomainRepository(repository)
	if !ok {
		return "-"
	}
	return repositoryListLabelFromRef(style, repositoryRef)
}

func repositoryListLabelFromRef(style string, repository githubdomain.RepositoryRef) string {
	switch appconfig.ResolveDisplayConfig(appconfig.DisplayConfig{RepositoryStyle: style}).RepositoryStyle {
	case appconfig.RepositoryStyleName:
		return valueOrDash(githubdomain.RepositoryShortName(repository))
	default:
		return valueOrDash(fullRepositoryLabel(repository))
	}
}

func fullRepositoryLabel(repository githubdomain.RepositoryRef) string {
	if nameWithOwner := strings.TrimSpace(repository.NameWithOwner); nameWithOwner != "" {
		return nameWithOwner
	}
	if name := strings.TrimSpace(repository.Name); name != "" {
		return name
	}
	return ""
}

func pullRequestListReference(style string, repository any, number int) string {
	return fmt.Sprintf("%s#%d", repositoryListLabel(style, repository), number)
}

func repositoryListLabelFromNameWithOwner(style string, repository string) string {
	repositoryRef := githubdomain.RepositoryRef{NameWithOwner: strings.TrimSpace(repository)}
	return repositoryListLabelFromRef(style, repositoryRef)
}

func notificationListReference(style string, notification githubdomain.Notification) string {
	if summary, ok := notification.PullRequestSummary(); ok {
		return pullRequestListReference(style, summary.Repository, summary.Number)
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		return fmt.Sprintf("%s#%d", repositoryListLabelFromNameWithOwner(style, repository), number)
	}
	if repository, _, ok := notification.ReleaseIdentity(); ok {
		return repositoryListLabelFromNameWithOwner(style, repository)
	}
	return repositoryListLabel(style, notification.Repository)
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
