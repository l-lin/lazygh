package githubcli

import (
	"encoding/json"
	"fmt"
	"strings"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

var ErrInvalidPullRequestResponse = fmt.Errorf("invalid pull request response")

const pullRequestSearchJSONFields = "title,number,repository,url,body,state,isDraft,updatedAt"

type Repository struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
}

type PullRequest struct {
	Title      string     `json:"title"`
	Number     int        `json:"number"`
	Repository Repository `json:"repository"`
	URL        string     `json:"url"`
	Body       string     `json:"body"`
	State      string     `json:"state"`
	IsDraft    bool       `json:"isDraft"`
	UpdatedAt  string     `json:"updatedAt"`
}

func (client *Client) ListPullRequests(commandArguments []string) ([]PullRequest, error) {
	resolvedCommandArguments := pullRequestSearchCommandArguments(commandArguments)
	result, err := client.runGH(FormatPullRequestSearchCommand(commandArguments), resolvedCommandArguments...)
	if err != nil {
		return nil, err
	}

	var pullRequests []PullRequest
	if err := json.Unmarshal(result.Stdout, &pullRequests); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestResponse, err)
	}

	for index := range pullRequests {
		pullRequests[index] = pullRequests[index].normalized()
	}

	return pullRequests, nil
}

func FormatPullRequestSearchCommand(commandArguments []string) string {
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

	resolvedCommandArguments = append(resolvedCommandArguments, "--json", pullRequestSearchJSONFields)
	return resolvedCommandArguments
}

func (pullRequest PullRequest) normalized() PullRequest {
	pullRequest.Title = strings.TrimSpace(pullRequest.Title)
	pullRequest.URL = strings.TrimSpace(pullRequest.URL)
	pullRequest.Body = strings.TrimSpace(pullRequest.Body)
	pullRequest.State = strings.TrimSpace(pullRequest.State)
	pullRequest.UpdatedAt = strings.TrimSpace(pullRequest.UpdatedAt)
	pullRequest.Repository = pullRequest.Repository.normalized()
	return pullRequest
}

func (repository Repository) normalized() Repository {
	repository.Name = strings.TrimSpace(repository.Name)
	repository.NameWithOwner = strings.TrimSpace(repository.NameWithOwner)
	return repository
}
