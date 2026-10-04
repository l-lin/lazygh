package githubcli

import (
	"strconv"
	"strings"
)

func (client *PullRequestDetailService) GetPullRequestSummary(repository string, number int) (PullRequest, error) {
	repository = strings.TrimSpace(repository)
	result, err := client.execute(rawCommand("pr", "view", strconv.Itoa(number), "-R", repository, "--json", "number,title,url,body,state,isDraft,updatedAt"))
	if err != nil {
		return PullRequest{}, err
	}
	detail, err := parsePullRequestDetailResponse(result.Stdout)
	if err != nil {
		return PullRequest{}, err
	}
	if detail.Number != number || strings.TrimSpace(detail.UpdatedAt) == "" {
		return PullRequest{}, ErrInvalidPullRequestDetailResponse
	}
	return PullRequest{Number: detail.Number, Repository: Repository{NameWithOwner: repository}, Title: detail.Title, URL: detail.URL, Body: detail.Body, State: detail.State, IsDraft: detail.IsDraft, UpdatedAt: detail.UpdatedAt}, nil
}
