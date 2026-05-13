package tui

import (
	"errors"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	transportgithub "github.com/l-lin/lazygh/internal/githubcli"
)

func isProviderUnauthenticatedError(err error) bool {
	return errors.Is(err, transportgithub.ErrUnauthenticated)
}

func isProviderUnavailableError(err error) bool {
	return errors.Is(err, transportgithub.ErrUnavailable)
}

func isProviderEmptyConnectedUserError(err error) bool {
	return errors.Is(err, transportgithub.ErrEmptyConnectedUser)
}

func formatPullRequestSearchCommand(commandArguments []string) string {
	return transportgithub.FormatPullRequestSearchCommand(commandArguments)
}

func formatAssignableUsersCommand(repository string) string {
	return transportgithub.FormatAssignableUsersCommand(repository)
}

func formatPullRequestBuildRunCommand(repository string, check githubdomain.PullRequestStatusCheck) string {
	return transportgithub.FormatPullRequestBuildRunCommand(repository, transportgithub.PullRequestStatusCheckFromDomain(check))
}

func formatPullRequestBuildRunJobsCommand(repository string, check githubdomain.PullRequestStatusCheck) string {
	return transportgithub.FormatPullRequestBuildRunJobsCommand(repository, transportgithub.PullRequestStatusCheckFromDomain(check))
}

func formatPullRequestBuildRunJobLogCommand(repository string, jobDatabaseID int) string {
	return transportgithub.FormatPullRequestBuildRunJobLogCommand(repository, jobDatabaseID)
}
