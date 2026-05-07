package githubcli

import (
	"encoding/json"
	"fmt"
	"strings"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

var (
	ErrInvalidPullRequestResponse               = fmt.Errorf("invalid pull request response")
	ErrInvalidPullRequestReviewMetadataResponse = fmt.Errorf("invalid pull request review metadata response")
)

const pullRequestSearchJSONFields = "title,number,repository,url,body,state,isDraft,updatedAt,id"
const pullRequestListReviewMetadataQuery = `
query($ids:[ID!]!) {
  nodes(ids:$ids) {
    ... on PullRequest {
      id
      reviewDecision
      reviewRequests(first:100) {
        nodes {
          requestedReviewer {
            __typename
            ... on User {
              login
              name
            }
            ... on Team {
              name
              slug
              organization {
                login
              }
            }
          }
        }
      }
    }
  }
}`

type Repository struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
}

type PullRequest struct {
	ID             string                     `json:"id"`
	Title          string                     `json:"title"`
	Number         int                        `json:"number"`
	Repository     Repository                 `json:"repository"`
	URL            string                     `json:"url"`
	Body           string                     `json:"body"`
	State          string                     `json:"state"`
	IsDraft        bool                       `json:"isDraft"`
	UpdatedAt      string                     `json:"updatedAt"`
	ReviewDecision string                     `json:"reviewDecision"`
	ReviewRequests []PullRequestReviewRequest `json:"reviewRequests"`
}

type pullRequestListReviewMetadata struct {
	ReviewDecision string
	ReviewRequests []PullRequestReviewRequest
}

type pullRequestListReviewMetadataResponse struct {
	Data *struct {
		Nodes []*struct {
			ID             string `json:"id"`
			ReviewDecision string `json:"reviewDecision"`
			ReviewRequests struct {
				Nodes []PullRequestReviewRequest `json:"nodes"`
			} `json:"reviewRequests"`
		} `json:"nodes"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
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

	reviewMetadataByID, err := client.listPullRequestReviewMetadata(pullRequests)
	if err != nil {
		return nil, err
	}

	for index := range pullRequests {
		if reviewMetadata, ok := reviewMetadataByID[pullRequests[index].ID]; ok {
			pullRequests[index].ReviewDecision = reviewMetadata.ReviewDecision
			pullRequests[index].ReviewRequests = append([]PullRequestReviewRequest(nil), reviewMetadata.ReviewRequests...)
		}
		pullRequests[index] = pullRequests[index].normalized()
	}

	return pullRequests, nil
}

func (client *Client) listPullRequestReviewMetadata(pullRequests []PullRequest) (map[string]pullRequestListReviewMetadata, error) {
	ids := uniquePullRequestIDs(pullRequests)
	if len(ids) == 0 {
		return nil, nil
	}

	args := []string{"api", "graphql", "-f", "query=" + pullRequestListReviewMetadataQuery}
	for _, id := range ids {
		args = append(args, "-F", "ids[]="+id)
	}

	result, err := client.runGH("gh api graphql", args...)
	if err != nil {
		return nil, err
	}

	return parsePullRequestListReviewMetadata(result.Stdout)
}

func uniquePullRequestIDs(pullRequests []PullRequest) []string {
	uniqueIDs := make([]string, 0, len(pullRequests))
	seen := map[string]bool{}
	for _, pullRequest := range pullRequests {
		trimmedID := strings.TrimSpace(pullRequest.ID)
		if trimmedID == "" || seen[trimmedID] {
			continue
		}
		seen[trimmedID] = true
		uniqueIDs = append(uniqueIDs, trimmedID)
	}
	return uniqueIDs
}

func parsePullRequestListReviewMetadata(stdout []byte) (map[string]pullRequestListReviewMetadata, error) {
	var response pullRequestListReviewMetadataResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewMetadataResponse, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return nil, fmt.Errorf("%w: %s", ErrInvalidPullRequestReviewMetadataResponse, message)
		}
	}
	if response.Data == nil {
		return nil, ErrInvalidPullRequestReviewMetadataResponse
	}

	reviewMetadataByID := make(map[string]pullRequestListReviewMetadata, len(response.Data.Nodes))
	for _, node := range response.Data.Nodes {
		if node == nil {
			continue
		}
		trimmedID := strings.TrimSpace(node.ID)
		if trimmedID == "" {
			continue
		}
		reviewMetadata := pullRequestListReviewMetadata{ReviewDecision: strings.TrimSpace(node.ReviewDecision)}
		if len(node.ReviewRequests.Nodes) > 0 {
			reviewRequests := make([]PullRequestReviewRequest, 0, len(node.ReviewRequests.Nodes))
			for _, reviewRequest := range node.ReviewRequests.Nodes {
				reviewRequests = append(reviewRequests, reviewRequest.normalized())
			}
			reviewMetadata.ReviewRequests = reviewRequests
		}
		reviewMetadataByID[trimmedID] = reviewMetadata
	}
	return reviewMetadataByID, nil
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
	pullRequest.ID = strings.TrimSpace(pullRequest.ID)
	pullRequest.Title = strings.TrimSpace(pullRequest.Title)
	pullRequest.URL = strings.TrimSpace(pullRequest.URL)
	pullRequest.Body = strings.TrimSpace(pullRequest.Body)
	pullRequest.State = strings.TrimSpace(pullRequest.State)
	pullRequest.UpdatedAt = strings.TrimSpace(pullRequest.UpdatedAt)
	pullRequest.ReviewDecision = strings.TrimSpace(pullRequest.ReviewDecision)
	if len(pullRequest.ReviewRequests) > 0 {
		normalizedReviewRequests := make([]PullRequestReviewRequest, 0, len(pullRequest.ReviewRequests))
		for _, reviewRequest := range pullRequest.ReviewRequests {
			normalizedReviewRequests = append(normalizedReviewRequests, reviewRequest.normalized())
		}
		pullRequest.ReviewRequests = normalizedReviewRequests
	}
	pullRequest.Repository = pullRequest.Repository.normalized()
	return pullRequest
}

func (repository Repository) normalized() Repository {
	repository.Name = strings.TrimSpace(repository.Name)
	repository.NameWithOwner = strings.TrimSpace(repository.NameWithOwner)
	return repository
}
