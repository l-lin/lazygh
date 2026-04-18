package tui

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	myPullRequestsLoadingTitle          = "Loading my pull requests..."
	myPullRequestsLoadingDetail         = "Running `gh search prs --author @me --state open` to load authored pull requests."
	myPullRequestsEmptyTitle            = "No open pull requests"
	myPullRequestsEmptyDetail           = "GitHub returned no open pull requests authored by the authenticated user."
	myPullRequestsUnauthenticatedTitle  = "GitHub authentication required"
	myPullRequestsUnauthenticatedDetail = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	myPullRequestsUnavailableTitle      = "`gh` not found"
	myPullRequestsUnavailableDetail     = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	myPullRequestsGenericErrorTitle     = "Could not load my pull requests"
	myPullRequestsGenericErrorPrefix    = "Failed to run `gh search prs --author @me --state open`."
)

func myPullRequestsLoadingItem() Item {
	return Item{
		Title:  myPullRequestsLoadingTitle,
		Detail: myPullRequestsLoadingDetail,
	}
}

func myPullRequestsStateItems(pullRequests []githubcli.PullRequest, err error) []Item {
	if err != nil {
		return []Item{myPullRequestsErrorItem(err)}
	}
	if len(pullRequests) == 0 {
		return []Item{myPullRequestsEmptyItem()}
	}

	items := make([]Item, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		items = append(items, myPullRequestItem(pullRequest))
	}
	return items
}

func myPullRequestItem(pullRequest githubcli.PullRequest) Item {
	repositoryName := pullRequestRepositoryName(pullRequest.Repository)
	body := strings.TrimSpace(pullRequest.Body)
	if body == "" {
		body = "No description available."
	}

	detailLines := []string{
		fmt.Sprintf("Repository: %s", repositoryName),
		fmt.Sprintf("Number: #%d", pullRequest.Number),
		fmt.Sprintf("State: %s", valueOrDash(pullRequest.State)),
		fmt.Sprintf("Draft: %s", yesNo(pullRequest.IsDraft)),
		fmt.Sprintf("Updated: %s", valueOrDash(pullRequest.UpdatedAt)),
		fmt.Sprintf("URL: %s", valueOrDash(pullRequest.URL)),
		"",
		body,
	}

	return Item{
		Title:  fmt.Sprintf("%s#%d %s", repositoryName, pullRequest.Number, valueOrDash(pullRequest.Title)),
		Detail: strings.Join(detailLines, "\n"),
	}
}

func myPullRequestsErrorItem(err error) Item {
	switch {
	case errors.Is(err, githubcli.ErrUnauthenticated):
		return Item{Title: myPullRequestsUnauthenticatedTitle, Detail: myPullRequestsUnauthenticatedDetail}
	case errors.Is(err, githubcli.ErrUnavailable):
		return Item{Title: myPullRequestsUnavailableTitle, Detail: myPullRequestsUnavailableDetail}
	default:
		return Item{Title: myPullRequestsGenericErrorTitle, Detail: formatPullRequestErrorDetail(err)}
	}
}

func myPullRequestsEmptyItem() Item {
	return Item{Title: myPullRequestsEmptyTitle, Detail: myPullRequestsEmptyDetail}
}

func formatPullRequestErrorDetail(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return myPullRequestsGenericErrorPrefix
	}

	return fmt.Sprintf("%s\n\n%s", myPullRequestsGenericErrorPrefix, message)
}

func pullRequestRepositoryName(repository githubcli.Repository) string {
	if repository.NameWithOwner != "" {
		return repository.NameWithOwner
	}
	if repository.Name != "" {
		return repository.Name
	}
	return "-"
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
