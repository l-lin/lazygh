package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
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
      autoMergeRequest {
        enabledAt
      }
      isMergeQueueEnabled
      isInMergeQueue
      viewerCanEnableAutoMerge
      mergeQueueEntry {
        id
        state
        position
        estimatedTimeToMerge
      }
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
	ID                       string                       `json:"id"`
	Title                    string                       `json:"title"`
	Number                   int                          `json:"number"`
	Repository               Repository                   `json:"repository"`
	URL                      string                       `json:"url"`
	Body                     string                       `json:"body"`
	State                    string                       `json:"state"`
	IsDraft                  bool                         `json:"isDraft"`
	UpdatedAt                string                       `json:"updatedAt"`
	ReviewDecision           string                       `json:"reviewDecision"`
	ReviewRequests           []PullRequestReviewRequest   `json:"reviewRequests"`
	MergeStateStatus         string                       `json:"mergeStateStatus"`
	Mergeable                string                       `json:"mergeable"`
	AutoMergeRequest         *PullRequestAutoMergeRequest `json:"autoMergeRequest,omitempty"`
	IsMergeQueueEnabled      bool                         `json:"isMergeQueueEnabled,omitempty"`
	IsInMergeQueue           bool                         `json:"isInMergeQueue,omitempty"`
	MergeQueueEntry          *PullRequestMergeQueueEntry  `json:"mergeQueueEntry,omitempty"`
	ViewerCanEnableAutoMerge bool                         `json:"viewerCanEnableAutoMerge,omitempty"`
	StatusCheckRollupState   string                       `json:"statusCheckRollupState"`
}

type pullRequestListReviewMetadata struct {
	ReviewDecision           string
	ReviewRequests           []PullRequestReviewRequest
	MergeStateStatus         string
	Mergeable                string
	AutoMergeRequest         *PullRequestAutoMergeRequest
	IsMergeQueueEnabled      bool
	IsInMergeQueue           bool
	MergeQueueEntry          *PullRequestMergeQueueEntry
	ViewerCanEnableAutoMerge bool
	StatusCheckRollupState   string
}

func (client *PullRequestListService) ListPullRequests(commandArguments []string) ([]PullRequest, error) {
	resolvedCommandArguments := pullRequestSearchCommandArguments(commandArguments)
	result, err := client.execute(rawCommand(resolvedCommandArguments...))
	if err != nil {
		return nil, err
	}

	pullRequests, err := parsePullRequestSearchResultsResponse(result.Stdout)
	if err != nil {
		return nil, err
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
			pullRequests[index].IsMergeQueueEnabled = reviewMetadata.IsMergeQueueEnabled
			pullRequests[index].IsInMergeQueue = reviewMetadata.IsInMergeQueue
			pullRequests[index].MergeQueueEntry = normalizePullRequestMergeQueueEntry(reviewMetadata.MergeQueueEntry)
			pullRequests[index].ViewerCanEnableAutoMerge = reviewMetadata.ViewerCanEnableAutoMerge
			if reviewMetadata.AutoMergeRequest != nil {
				normalizedRequest := reviewMetadata.AutoMergeRequest.normalized()
				pullRequests[index].AutoMergeRequest = &normalizedRequest
			} else {
				pullRequests[index].AutoMergeRequest = nil
			}
			pullRequests[index].StatusCheckRollupState = reviewMetadata.StatusCheckRollupState
		}
		pullRequests[index] = pullRequests[index].normalized()
	}

	return pullRequests, nil
}

func (client *PullRequestListService) listPullRequestReviewMetadata(pullRequests []PullRequest) (map[string]pullRequestListReviewMetadata, error) {
	ids := uniquePullRequestIDs(pullRequests)
	if len(ids) == 0 {
		return nil, nil
	}

	metadataByID := map[string]pullRequestListReviewMetadata{}
	for batchStart := 0; batchStart < len(ids); batchStart += pullRequestListReviewMetadataBatchSize {
		batchEnd := min(batchStart+pullRequestListReviewMetadataBatchSize, len(ids))

		batchMetadataByID, err := client.listPullRequestReviewMetadataBatch(ids[batchStart:batchEnd])
		maps.Copy(metadataByID, batchMetadataByID)
		if err != nil {
			return metadataByID, err
		}
	}

	if len(metadataByID) == 0 {
		return nil, nil
	}
	return metadataByID, nil
}

func (client *PullRequestListService) listPullRequestReviewMetadataBatch(ids []string) (map[string]pullRequestListReviewMetadata, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	request := GraphQLRequest{Query: pullRequestListReviewMetadataQuery}
	for _, id := range ids {
		request.Variables = append(request.Variables, typedGraphQLVariable("ids[]", id))
	}

	result, err := client.queryGraphQL(request)
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
	var response struct {
		Nodes []*struct {
			ID                       string                       `json:"id"`
			ReviewDecision           string                       `json:"reviewDecision"`
			MergeStateStatus         string                       `json:"mergeStateStatus"`
			Mergeable                string                       `json:"mergeable"`
			AutoMergeRequest         *PullRequestAutoMergeRequest `json:"autoMergeRequest"`
			IsMergeQueueEnabled      bool                         `json:"isMergeQueueEnabled"`
			IsInMergeQueue           bool                         `json:"isInMergeQueue"`
			ViewerCanEnableAutoMerge bool                         `json:"viewerCanEnableAutoMerge"`
			MergeQueueEntry          *PullRequestMergeQueueEntry  `json:"mergeQueueEntry"`
			ReviewRequests           struct {
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
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidPullRequestReviewMetadataResponse); err != nil {
		return nil, err
	}
	return mapPullRequestListReviewMetadataResponse(response.Nodes), nil
}

func mapPullRequestListReviewMetadataResponse(nodes []*struct {
	ID                       string                       `json:"id"`
	ReviewDecision           string                       `json:"reviewDecision"`
	MergeStateStatus         string                       `json:"mergeStateStatus"`
	Mergeable                string                       `json:"mergeable"`
	AutoMergeRequest         *PullRequestAutoMergeRequest `json:"autoMergeRequest"`
	IsMergeQueueEnabled      bool                         `json:"isMergeQueueEnabled"`
	IsInMergeQueue           bool                         `json:"isInMergeQueue"`
	ViewerCanEnableAutoMerge bool                         `json:"viewerCanEnableAutoMerge"`
	MergeQueueEntry          *PullRequestMergeQueueEntry  `json:"mergeQueueEntry"`
	ReviewRequests           struct {
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
}) map[string]pullRequestListReviewMetadata {
	if len(nodes) == 0 {
		return nil
	}

	reviewMetadataByID := make(map[string]pullRequestListReviewMetadata, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		trimmedID := strings.TrimSpace(node.ID)
		if trimmedID == "" {
			continue
		}
		reviewMetadata := pullRequestListReviewMetadata{
			ReviewDecision:           strings.TrimSpace(node.ReviewDecision),
			MergeStateStatus:         strings.TrimSpace(node.MergeStateStatus),
			Mergeable:                strings.TrimSpace(node.Mergeable),
			IsMergeQueueEnabled:      node.IsMergeQueueEnabled,
			IsInMergeQueue:           node.IsInMergeQueue,
			ViewerCanEnableAutoMerge: node.ViewerCanEnableAutoMerge,
			MergeQueueEntry:          normalizePullRequestMergeQueueEntry(node.MergeQueueEntry),
			StatusCheckRollupState:   pullRequestListStatusCheckRollupState(node.HeadRefStatusCheckRollup.Nodes),
		}
		if node.AutoMergeRequest != nil {
			normalizedRequest := node.AutoMergeRequest.normalized()
			reviewMetadata.AutoMergeRequest = &normalizedRequest
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
	if len(reviewMetadataByID) == 0 {
		return nil
	}
	return reviewMetadataByID
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
	return formatCommandArguments(pullRequestSearchCommandArguments(commandArguments))
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
	if pullRequest.AutoMergeRequest != nil {
		normalizedRequest := pullRequest.AutoMergeRequest.normalized()
		pullRequest.AutoMergeRequest = &normalizedRequest
	}
	pullRequest.MergeQueueEntry = normalizePullRequestMergeQueueEntry(pullRequest.MergeQueueEntry)
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
