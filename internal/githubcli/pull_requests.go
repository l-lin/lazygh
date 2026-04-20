package githubcli

import (
	"encoding/json"
	"fmt"
	"strings"
)

const pullRequestJSONFields = "title,number,repository,url,body,state,isDraft,updatedAt"

var ErrInvalidPullRequestResponse = fmt.Errorf("invalid pull request response")

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

func (client *Client) ListMyPullRequests() ([]PullRequest, error) {
	return client.listPullRequests("gh search prs", "--author", "@me")
}

func (client *Client) ListRequestedPullRequests() ([]PullRequest, error) {
	return client.listPullRequests("gh search prs", "--review-requested", "@me")
}

func (client *Client) listPullRequests(commandName string, qualifiers ...string) ([]PullRequest, error) {
	args := []string{"search", "prs"}
	args = append(args, qualifiers...)
	args = append(args, "--state", "open", "--json", pullRequestJSONFields)

	result, err := client.runGH(commandName, args...)
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
