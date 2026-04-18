package tui

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	myPullRequestsLoadingTitle                 = "Loading my pull requests..."
	myPullRequestsLoadingDetail                = "Running `gh search prs --author @me --state open` to load authored pull requests."
	myPullRequestsEmptyTitle                   = "No open pull requests"
	myPullRequestsEmptyDetail                  = "GitHub returned no open pull requests authored by the authenticated user."
	myPullRequestsUnauthenticatedTitle         = "GitHub authentication required"
	myPullRequestsUnauthenticatedDetail        = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	myPullRequestsUnavailableTitle             = "`gh` not found"
	myPullRequestsUnavailableDetail            = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	myPullRequestsGenericErrorTitle            = "Could not load my pull requests"
	myPullRequestsGenericErrorPrefix           = "Failed to run `gh search prs --author @me --state open`."
	requestedPullRequestsLoadingTitle          = "Loading requested pull requests..."
	requestedPullRequestsLoadingDetail         = "Running `gh search prs --review-requested @me --state open` to load review requests."
	requestedPullRequestsEmptyTitle            = "No requested pull requests"
	requestedPullRequestsEmptyDetail           = "GitHub returned no open pull requests requesting review from the authenticated user."
	requestedPullRequestsUnauthenticatedTitle  = "GitHub authentication required"
	requestedPullRequestsUnauthenticatedDetail = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	requestedPullRequestsUnavailableTitle      = "`gh` not found"
	requestedPullRequestsUnavailableDetail     = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	requestedPullRequestsGenericErrorTitle     = "Could not load requested pull requests"
	requestedPullRequestsGenericErrorPrefix    = "Failed to run `gh search prs --review-requested @me --state open`."
)

type pullRequestListState struct {
	loadingTitle          string
	loadingDetail         string
	emptyTitle            string
	emptyDetail           string
	unauthenticatedTitle  string
	unauthenticatedDetail string
	unavailableTitle      string
	unavailableDetail     string
	genericErrorTitle     string
	genericErrorPrefix    string
}

var (
	myPullRequestsState = pullRequestListState{
		loadingTitle:          myPullRequestsLoadingTitle,
		loadingDetail:         myPullRequestsLoadingDetail,
		emptyTitle:            myPullRequestsEmptyTitle,
		emptyDetail:           myPullRequestsEmptyDetail,
		unauthenticatedTitle:  myPullRequestsUnauthenticatedTitle,
		unauthenticatedDetail: myPullRequestsUnauthenticatedDetail,
		unavailableTitle:      myPullRequestsUnavailableTitle,
		unavailableDetail:     myPullRequestsUnavailableDetail,
		genericErrorTitle:     myPullRequestsGenericErrorTitle,
		genericErrorPrefix:    myPullRequestsGenericErrorPrefix,
	}
	requestedPullRequestsState = pullRequestListState{
		loadingTitle:          requestedPullRequestsLoadingTitle,
		loadingDetail:         requestedPullRequestsLoadingDetail,
		emptyTitle:            requestedPullRequestsEmptyTitle,
		emptyDetail:           requestedPullRequestsEmptyDetail,
		unauthenticatedTitle:  requestedPullRequestsUnauthenticatedTitle,
		unauthenticatedDetail: requestedPullRequestsUnauthenticatedDetail,
		unavailableTitle:      requestedPullRequestsUnavailableTitle,
		unavailableDetail:     requestedPullRequestsUnavailableDetail,
		genericErrorTitle:     requestedPullRequestsGenericErrorTitle,
		genericErrorPrefix:    requestedPullRequestsGenericErrorPrefix,
	}
)

func myPullRequestsLoadingItem() Item {
	return pullRequestLoadingItem(myPullRequestsState)
}

func requestedPullRequestsLoadingItem() Item {
	return pullRequestLoadingItem(requestedPullRequestsState)
}

func myPullRequestsStateRows(pullRequests []githubcli.PullRequest, err error) []PullRequestRow {
	return pullRequestStateRows(myPullRequestsState, pullRequests, err)
}

func requestedPullRequestsStateRows(pullRequests []githubcli.PullRequest, err error) []PullRequestRow {
	return pullRequestStateRows(requestedPullRequestsState, pullRequests, err)
}

func myPullRequestsStateItems(pullRequests []githubcli.PullRequest, err error) []Item {
	return pullRequestItems(myPullRequestsStateRows(pullRequests, err))
}

func requestedPullRequestsStateItems(pullRequests []githubcli.PullRequest, err error) []Item {
	return pullRequestItems(requestedPullRequestsStateRows(pullRequests, err))
}

func myPullRequestRow(pullRequest githubcli.PullRequest) PullRequestRow {
	return pullRequestRow(pullRequest)
}

func requestedPullRequestRow(pullRequest githubcli.PullRequest) PullRequestRow {
	return pullRequestRow(pullRequest)
}

func myPullRequestItem(pullRequest githubcli.PullRequest) Item {
	return myPullRequestRow(pullRequest).Item
}

func requestedPullRequestItem(pullRequest githubcli.PullRequest) Item {
	return requestedPullRequestRow(pullRequest).Item
}

func myPullRequestsErrorItem(err error) Item {
	return pullRequestErrorItem(myPullRequestsState, err)
}

func requestedPullRequestsErrorItem(err error) Item {
	return pullRequestErrorItem(requestedPullRequestsState, err)
}

func pullRequestLoadingItem(state pullRequestListState) Item {
	return Item{Title: state.loadingTitle, Detail: state.loadingDetail}
}

func pullRequestStateRows(state pullRequestListState, pullRequests []githubcli.PullRequest, err error) []PullRequestRow {
	if err != nil {
		return []PullRequestRow{{Item: pullRequestErrorItem(state, err)}}
	}
	if len(pullRequests) == 0 {
		return []PullRequestRow{{Item: pullRequestEmptyItem(state)}}
	}

	rows := make([]PullRequestRow, 0, len(pullRequests))
	for _, pullRequest := range pullRequests {
		rows = append(rows, pullRequestRow(pullRequest))
	}
	return rows
}

func pullRequestRow(pullRequest githubcli.PullRequest) PullRequestRow {
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

	summaryCopy := pullRequest
	return PullRequestRow{
		Item: Item{
			Title:  fmt.Sprintf("%s#%d %s", repositoryName, pullRequest.Number, valueOrDash(pullRequest.Title)),
			Detail: strings.Join(detailLines, "\n"),
		},
		Summary: &summaryCopy,
	}
}

func pullRequestErrorItem(state pullRequestListState, err error) Item {
	switch {
	case errors.Is(err, githubcli.ErrUnauthenticated):
		return Item{Title: state.unauthenticatedTitle, Detail: state.unauthenticatedDetail}
	case errors.Is(err, githubcli.ErrUnavailable):
		return Item{Title: state.unavailableTitle, Detail: state.unavailableDetail}
	default:
		return Item{Title: state.genericErrorTitle, Detail: formatPullRequestErrorDetail(state.genericErrorPrefix, err)}
	}
}

func pullRequestEmptyItem(state pullRequestListState) Item {
	return Item{Title: state.emptyTitle, Detail: state.emptyDetail}
}

func formatPullRequestErrorDetail(prefix string, err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return prefix
	}

	return fmt.Sprintf("%s\n\n%s", prefix, message)
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
