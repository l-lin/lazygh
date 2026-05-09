package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

var (
	ErrInvalidPullRequestResponse               = fmt.Errorf("invalid pull request response")
	ErrInvalidPullRequestReviewMetadataResponse = fmt.Errorf("invalid pull request review metadata response")
)

const (
	pullRequestSearchJSONFields            = "title,number,repository,url,body,state,isDraft,updatedAt,id"
	pullRequestListReviewMetadataBatchSize = 20
)

const pullRequestListReviewMetadataQuery = `
query($ids:[ID!]!) {
  nodes(ids:$ids) {
    ... on PullRequest {
      id
      reviewDecision
      mergeable
      mergeStateStatus
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
      headRefStatusCheckRollup: commits(last:1) {
        nodes {
          commit {
            statusCheckRollup {
              state
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

func (repository *Repository) UnmarshalJSON(data []byte) error {
	var payload struct {
		Name          string `json:"name"`
		NameWithOwner string `json:"nameWithOwner"`
		FullName      string `json:"full_name"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	repository.Name = payload.Name
	repository.NameWithOwner = payload.NameWithOwner
	if strings.TrimSpace(repository.NameWithOwner) == "" {
		repository.NameWithOwner = payload.FullName
	}
	return nil
}

type PullRequest struct {
	ID                     string                     `json:"id"`
	Title                  string                     `json:"title"`
	Number                 int                        `json:"number"`
	Repository             Repository                 `json:"repository"`
	URL                    string                     `json:"url"`
	Body                   string                     `json:"body"`
	State                  string                     `json:"state"`
	IsDraft                bool                       `json:"isDraft"`
	UpdatedAt              string                     `json:"updatedAt"`
	ReviewDecision         string                     `json:"reviewDecision"`
	ReviewRequests         []PullRequestReviewRequest `json:"reviewRequests"`
	MergeStateStatus       string                     `json:"mergeStateStatus"`
	Mergeable              string                     `json:"mergeable"`
	StatusCheckRollupState string                     `json:"statusCheckRollupState"`
}

type pullRequestListReviewMetadata struct {
	ReviewDecision         string
	ReviewRequests         []PullRequestReviewRequest
	MergeStateStatus       string
	Mergeable              string
	StatusCheckRollupState string
}

type pullRequestListReviewMetadataResponse struct {
	Data *struct {
		Nodes []*struct {
			ID               string `json:"id"`
			ReviewDecision   string `json:"reviewDecision"`
			MergeStateStatus string `json:"mergeStateStatus"`
			Mergeable        string `json:"mergeable"`
			ReviewRequests   struct {
				Nodes []PullRequestReviewRequest `json:"nodes"`
			} `json:"reviewRequests"`
			HeadRefStatusCheckRollup struct {
				Nodes []struct {
					Commit struct {
						StatusCheckRollup *struct {
							State string `json:"state"`
						} `json:"statusCheckRollup"`
					} `json:"commit"`
				} `json:"nodes"`
			} `json:"headRefStatusCheckRollup"`
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
	if err != nil && !shouldIgnorePullRequestReviewMetadataError(err) {
		return nil, err
	}

	for index := range pullRequests {
		if reviewMetadata, ok := reviewMetadataByID[pullRequests[index].ID]; ok {
			pullRequests[index].ReviewDecision = reviewMetadata.ReviewDecision
			pullRequests[index].ReviewRequests = append([]PullRequestReviewRequest(nil), reviewMetadata.ReviewRequests...)
			pullRequests[index].MergeStateStatus = reviewMetadata.MergeStateStatus
			pullRequests[index].Mergeable = reviewMetadata.Mergeable
			pullRequests[index].StatusCheckRollupState = reviewMetadata.StatusCheckRollupState
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

	metadataByID := map[string]pullRequestListReviewMetadata{}
	for batchStart := 0; batchStart < len(ids); batchStart += pullRequestListReviewMetadataBatchSize {
		batchEnd := batchStart + pullRequestListReviewMetadataBatchSize
		if batchEnd > len(ids) {
			batchEnd = len(ids)
		}

		batchMetadataByID, err := client.listPullRequestReviewMetadataBatch(ids[batchStart:batchEnd])
		for id, metadata := range batchMetadataByID {
			metadataByID[id] = metadata
		}
		if err != nil {
			return metadataByID, err
		}
	}

	if len(metadataByID) == 0 {
		return nil, nil
	}
	return metadataByID, nil
}

func (client *Client) listPullRequestReviewMetadataBatch(ids []string) (map[string]pullRequestListReviewMetadata, error) {
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

func shouldIgnorePullRequestReviewMetadataError(err error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, ErrInvalidPullRequestReviewMetadataResponse)
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
		reviewMetadata := pullRequestListReviewMetadata{
			ReviewDecision:         strings.TrimSpace(node.ReviewDecision),
			MergeStateStatus:       strings.TrimSpace(node.MergeStateStatus),
			Mergeable:              strings.TrimSpace(node.Mergeable),
			StatusCheckRollupState: pullRequestListStatusCheckRollupState(node.HeadRefStatusCheckRollup.Nodes),
		}
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

func pullRequestListStatusCheckRollupState(nodes []struct {
	Commit struct {
		StatusCheckRollup *struct {
			State string `json:"state"`
		} `json:"statusCheckRollup"`
	} `json:"commit"`
}) string {
	for _, node := range nodes {
		if node.Commit.StatusCheckRollup == nil {
			continue
		}
		state := strings.TrimSpace(node.Commit.StatusCheckRollup.State)
		if state != "" {
			return state
		}
	}
	return ""
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
	pullRequest.MergeStateStatus = strings.TrimSpace(pullRequest.MergeStateStatus)
	pullRequest.Mergeable = strings.TrimSpace(pullRequest.Mergeable)
	pullRequest.StatusCheckRollupState = strings.TrimSpace(pullRequest.StatusCheckRollupState)
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
