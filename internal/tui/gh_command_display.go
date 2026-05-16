package tui

import (
	"errors"
	"strconv"
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const pullRequestSearchDisplayJSONFields = "title,number,repository,url,body,state,isDraft,updatedAt,id"

func isProviderUnauthenticatedError(err error) bool {
	return errors.Is(err, githubdomain.ErrUnauthenticated)
}

func isProviderUnavailableError(err error) bool {
	return errors.Is(err, githubdomain.ErrUnavailable)
}

func isProviderEmptyConnectedUserError(err error) bool {
	return errors.Is(err, githubdomain.ErrEmptyConnectedUser)
}

func formatPullRequestSearchCommand(commandArguments []string) string {
	return appconfig.FormatGHCommand(pullRequestSearchCommandArguments(commandArguments))
}

func pullRequestSearchCommandArguments(commandArguments []string) []string {
	resolvedCommandArguments := make([]string, 0, len(commandArguments)+2)
	for index := 0; index < len(commandArguments); index++ {
		argument := commandArguments[index]
		switch {
		case argument == "--json":
			index++
			continue
		case strings.HasPrefix(argument, "--json="):
			continue
		default:
			resolvedCommandArguments = append(resolvedCommandArguments, argument)
		}
	}

	resolvedCommandArguments = append(resolvedCommandArguments, "--json", pullRequestSearchDisplayJSONFields)
	return resolvedCommandArguments
}

func formatAssignableUsersCommand(repository string) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return appconfig.FormatGHCommand([]string{"api"})
	}
	return appconfig.FormatGHCommand([]string{"api", "repos/" + trimmedRepository + "/assignees?per_page=100", "--paginate", "--slurp"})
}

func formatAssigneeSearchCommand(repository string, query string) string {
	trimmedRepository := strings.TrimSpace(repository)
	owner, name, ok := strings.Cut(trimmedRepository, "/")
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	if !ok || owner == "" || name == "" {
		return appconfig.FormatGHCommand([]string{"api", "graphql"})
	}

	args := []string{"api", "graphql", "-F", "owner=" + owner, "-F", "name=" + name, "-F", "first=" + strconv.Itoa(assigneePickerSearchResultLimit)}
	if trimmedQuery := strings.TrimSpace(query); trimmedQuery != "" {
		args = append(args, "-F", "search="+trimmedQuery)
	}
	return appconfig.FormatGHCommand(args)
}

func formatPullRequestBuildRunCommand(repository string, check githubdomain.PullRequestStatusCheck) string {
	args, err := pullRequestBuildRunCommandArguments(repository, check)
	if err != nil {
		return appconfig.FormatGHCommand([]string{"run", "view"})
	}
	return appconfig.FormatGHCommand(args)
}

func formatPullRequestBuildRunJobsCommand(repository string, check githubdomain.PullRequestStatusCheck) string {
	args, err := pullRequestBuildRunJobsCommandArguments(repository, check)
	if err != nil {
		return appconfig.FormatGHCommand([]string{"run", "view"})
	}
	return appconfig.FormatGHCommand(args)
}

func pullRequestBuildRunCommandArguments(repository string, check githubdomain.PullRequestStatusCheck) ([]string, error) {
	reference, trimmedRepository, err := pullRequestBuildRunCommandContext(repository, check)
	if err != nil {
		return nil, err
	}

	args := []string{"run", "view", reference.ID, "-R", trimmedRepository}
	if reference.Attempt > 0 {
		args = append(args, "--attempt", strconv.Itoa(reference.Attempt))
	}
	args = append(args, "--verbose")
	return args, nil
}

func pullRequestBuildRunJobsCommandArguments(repository string, check githubdomain.PullRequestStatusCheck) ([]string, error) {
	reference, trimmedRepository, err := pullRequestBuildRunCommandContext(repository, check)
	if err != nil {
		return nil, err
	}

	args := []string{"run", "view", reference.ID, "-R", trimmedRepository}
	if reference.Attempt > 0 {
		args = append(args, "--attempt", strconv.Itoa(reference.Attempt))
	}
	args = append(args, "--json", "jobs")
	return args, nil
}

func pullRequestBuildRunCommandContext(repository string, check githubdomain.PullRequestStatusCheck) (githubdomain.BuildRunReference, string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" {
		return githubdomain.BuildRunReference{}, "", githubdomain.ErrMissingPullRequestIdentity
	}
	actualReference, err := githubdomain.ParseBuildRunReferenceFromURL(check.Link)
	if err != nil {
		return githubdomain.BuildRunReference{}, "", err
	}
	return actualReference, trimmedRepository, nil
}
