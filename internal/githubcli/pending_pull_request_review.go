package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const pendingPullRequestReviewQuery = `query($owner:String!,$name:String!,$number:Int!){viewer{login}repository(owner:$owner,name:$name){pullRequest(number:$number){id reviews(last:100){nodes{id state author{login}}}}}}`
const addPendingPullRequestReviewMutation = `mutation($pullRequestId:ID!){addPullRequestReview(input:{pullRequestId:$pullRequestId}){pullRequestReview{id}}}`

var ErrInvalidPendingPullRequestReviewResponse = errors.New("invalid pending pull request review response")

func (client *ReviewService) GetPendingPullRequestReviewID(repository string, number int) (string, bool, error) {
	lookup, err := client.pendingPullRequestReviewLookup(repository, number)
	if err != nil {
		return "", false, err
	}
	if lookup.pendingReviewID == "" {
		return "", false, nil
	}

	return lookup.pendingReviewID, true, nil
}

func (client *ReviewService) StartPendingPullRequestReview(repository string, number int) (string, error) {
	lookup, err := client.pendingPullRequestReviewLookup(repository, number)
	if err != nil {
		return "", err
	}
	if lookup.pendingReviewID != "" {
		return lookup.pendingReviewID, nil
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: addPendingPullRequestReviewMutation, Variables: []GraphQLVariable{typedGraphQLVariable("pullRequestId", lookup.pullRequestID)}})
	if err != nil {
		return "", err
	}

	reviewID, err := parsePendingPullRequestReviewID(result.Stdout)
	if err != nil {
		return "", err
	}

	return reviewID, nil
}

func (client *ReviewService) pendingPullRequestReviewLookup(repository string, number int) (pendingPullRequestReviewLookupResult, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return pendingPullRequestReviewLookupResult{}, err
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return pendingPullRequestReviewLookupResult{}, err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: pendingPullRequestReviewQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", owner), typedGraphQLVariable("name", name), typedGraphQLVariable("number", number)}})
	if err != nil {
		return pendingPullRequestReviewLookupResult{}, err
	}

	return parsePendingPullRequestReviewLookup(result.Stdout)
}

func splitRepositoryOwnerAndName(repository string) (string, string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	owner, name, ok := strings.Cut(trimmedRepository, "/")
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	if !ok || owner == "" || name == "" {
		return "", "", ErrMissingPullRequestIdentity
	}

	return owner, name, nil
}

type pendingPullRequestReviewLookupResult struct {
	pullRequestID   string
	pendingReviewID string
}

type pendingPullRequestReviewLookupResponse struct {
	Data struct {
		Viewer struct {
			Login string `json:"login"`
		} `json:"viewer"`
		Repository *struct {
			PullRequest *struct {
				ID      string `json:"id"`
				Reviews struct {
					Nodes []struct {
						ID     string `json:"id"`
						State  string `json:"state"`
						Author *struct {
							Login string `json:"login"`
						} `json:"author"`
					} `json:"nodes"`
				} `json:"reviews"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

func parsePendingPullRequestReviewLookup(stdout []byte) (pendingPullRequestReviewLookupResult, error) {
	var response pendingPullRequestReviewLookupResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return pendingPullRequestReviewLookupResult{}, fmt.Errorf("%w: %v", ErrInvalidPendingPullRequestReviewResponse, err)
	}
	if response.Data.Repository == nil || response.Data.Repository.PullRequest == nil {
		return pendingPullRequestReviewLookupResult{}, ErrInvalidPendingPullRequestReviewResponse
	}

	viewerLogin := strings.TrimSpace(response.Data.Viewer.Login)
	pullRequestID := strings.TrimSpace(response.Data.Repository.PullRequest.ID)
	if viewerLogin == "" || pullRequestID == "" {
		return pendingPullRequestReviewLookupResult{}, ErrInvalidPendingPullRequestReviewResponse
	}

	lookup := pendingPullRequestReviewLookupResult{pullRequestID: pullRequestID}
	for index := len(response.Data.Repository.PullRequest.Reviews.Nodes) - 1; index >= 0; index-- {
		review := response.Data.Repository.PullRequest.Reviews.Nodes[index]
		if !strings.EqualFold(strings.TrimSpace(review.State), "PENDING") || review.Author == nil || !strings.EqualFold(strings.TrimSpace(review.Author.Login), viewerLogin) {
			continue
		}

		lookup.pendingReviewID = strings.TrimSpace(review.ID)
		if lookup.pendingReviewID != "" {
			return lookup, nil
		}
	}

	return lookup, nil
}

type addPendingPullRequestReviewResponse struct {
	Data struct {
		AddPullRequestReview *struct {
			PullRequestReview *struct {
				ID string `json:"id"`
			} `json:"pullRequestReview"`
		} `json:"addPullRequestReview"`
	} `json:"data"`
}

func parsePendingPullRequestReviewID(stdout []byte) (string, error) {
	var response addPendingPullRequestReviewResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPendingPullRequestReviewResponse, err)
	}
	if response.Data.AddPullRequestReview == nil || response.Data.AddPullRequestReview.PullRequestReview == nil {
		return "", ErrInvalidPendingPullRequestReviewResponse
	}

	reviewID := strings.TrimSpace(response.Data.AddPullRequestReview.PullRequestReview.ID)
	if reviewID == "" {
		return "", ErrInvalidPendingPullRequestReviewResponse
	}

	return reviewID, nil
}
