package githubcli

import (
	"errors"
	"strings"
)

const pullRequestMergeQueueQuery = `query($owner:String!,$name:String!,$number:Int!){repository(owner:$owner,name:$name){pullRequest(number:$number){isMergeQueueEnabled isInMergeQueue viewerCanEnableAutoMerge mergeQueueEntry{id state position estimatedTimeToMerge}}}}`
const enqueuePullRequestMutation = `mutation($pullRequestId:ID!){enqueuePullRequest(input:{pullRequestId:$pullRequestId}){mergeQueueEntry{id state position}}}`
const dequeuePullRequestMutation = `mutation($pullRequestId:ID!){dequeuePullRequest(input:{pullRequestId:$pullRequestId}){mergeQueueEntry{id state position}}}`

var (
	ErrInvalidPullRequestMergeQueueMetadataResponse = errors.New("invalid pull request merge queue response")
	ErrInvalidPullRequestMergeQueueMutationResponse = errors.New("invalid pull request merge queue mutation response")
	ErrMissingPullRequestID                         = errors.New("missing pull request id")
)

type PullRequestMergeQueueEntry struct {
	ID                   string `json:"id,omitempty"`
	State                string `json:"state,omitempty"`
	Position             int    `json:"position,omitempty"`
	EstimatedTimeToMerge int    `json:"estimatedTimeToMerge,omitempty"`
}

type pullRequestMergeQueueMetadata struct {
	IsMergeQueueEnabled      bool
	IsInMergeQueue           bool
	ViewerCanEnableAutoMerge bool
	MergeQueueEntry          *PullRequestMergeQueueEntry
}

type pullRequestMergeQueueMutationPayload struct {
	MergeQueueEntry *PullRequestMergeQueueEntry `json:"mergeQueueEntry"`
}

func (entry PullRequestMergeQueueEntry) normalized() PullRequestMergeQueueEntry {
	entry.ID = strings.TrimSpace(entry.ID)
	entry.State = strings.TrimSpace(entry.State)
	return entry
}

func normalizePullRequestMergeQueueEntry(entry *PullRequestMergeQueueEntry) *PullRequestMergeQueueEntry {
	if entry == nil {
		return nil
	}
	actual := entry.normalized()
	return &actual
}

func (metadata pullRequestMergeQueueMetadata) normalized() pullRequestMergeQueueMetadata {
	metadata.MergeQueueEntry = normalizePullRequestMergeQueueEntry(metadata.MergeQueueEntry)
	return metadata
}

func applyPullRequestMergeQueueMetadata(detail PullRequestDetail, metadata pullRequestMergeQueueMetadata) PullRequestDetail {
	normalizedMetadata := metadata.normalized()
	detail.IsMergeQueueEnabled = normalizedMetadata.IsMergeQueueEnabled
	detail.IsInMergeQueue = normalizedMetadata.IsInMergeQueue
	detail.ViewerCanEnableAutoMerge = normalizedMetadata.ViewerCanEnableAutoMerge
	detail.MergeQueueEntry = normalizedMetadata.MergeQueueEntry
	return detail
}

func (client *PullRequestDetailService) loadPullRequestMergeQueueMetadata(repository string, number int) (pullRequestMergeQueueMetadata, error) {
	trimmedRepository, err := normalizePullRequestIdentity(repository, number)
	if err != nil {
		return pullRequestMergeQueueMetadata{}, err
	}

	owner, name, err := splitRepositoryOwnerAndName(trimmedRepository)
	if err != nil {
		return pullRequestMergeQueueMetadata{}, err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: pullRequestMergeQueueQuery, Variables: []GraphQLVariable{typedGraphQLVariable("owner", owner), typedGraphQLVariable("name", name), typedGraphQLVariable("number", number)}})
	if err != nil {
		return pullRequestMergeQueueMetadata{}, err
	}

	return parsePullRequestMergeQueueMetadata(result.Stdout)
}

func parsePullRequestMergeQueueMetadata(stdout []byte) (pullRequestMergeQueueMetadata, error) {
	var response struct {
		Repository *struct {
			PullRequest *struct {
				IsMergeQueueEnabled      bool                        `json:"isMergeQueueEnabled"`
				IsInMergeQueue           bool                        `json:"isInMergeQueue"`
				ViewerCanEnableAutoMerge bool                        `json:"viewerCanEnableAutoMerge"`
				MergeQueueEntry          *PullRequestMergeQueueEntry `json:"mergeQueueEntry"`
			} `json:"pullRequest"`
		} `json:"repository"`
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidPullRequestMergeQueueMetadataResponse); err != nil {
		return pullRequestMergeQueueMetadata{}, err
	}
	if response.Repository == nil || response.Repository.PullRequest == nil {
		return pullRequestMergeQueueMetadata{}, ErrInvalidPullRequestMergeQueueMetadataResponse
	}

	return pullRequestMergeQueueMetadata{
		IsMergeQueueEnabled:      response.Repository.PullRequest.IsMergeQueueEnabled,
		IsInMergeQueue:           response.Repository.PullRequest.IsInMergeQueue,
		ViewerCanEnableAutoMerge: response.Repository.PullRequest.ViewerCanEnableAutoMerge,
		MergeQueueEntry:          normalizePullRequestMergeQueueEntry(response.Repository.PullRequest.MergeQueueEntry),
	}, nil
}

func (client *PullRequestMutationService) EnqueuePullRequest(pullRequestID string) error {
	return client.runPullRequestMergeQueueMutation(enqueuePullRequestMutation, "enqueuePullRequest", pullRequestID)
}

func (client *PullRequestMutationService) DequeuePullRequest(pullRequestID string) error {
	return client.runPullRequestMergeQueueMutation(dequeuePullRequestMutation, "dequeuePullRequest", pullRequestID)
}

func (client *PullRequestMutationService) runPullRequestMergeQueueMutation(mutation string, mutationField string, pullRequestID string) error {
	trimmedPullRequestID, err := validateNonEmptyPullRequestField(pullRequestID, ErrMissingPullRequestID)
	if err != nil {
		return err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: mutation, Variables: []GraphQLVariable{typedGraphQLVariable("pullRequestId", trimmedPullRequestID)}})
	if err != nil {
		return err
	}

	return parsePullRequestMergeQueueMutation(result.Stdout, mutationField)
}

func parsePullRequestMergeQueueMutation(stdout []byte, mutationField string) error {
	var response struct {
		EnqueuePullRequest *pullRequestMergeQueueMutationPayload `json:"enqueuePullRequest"`
		DequeuePullRequest *pullRequestMergeQueueMutationPayload `json:"dequeuePullRequest"`
	}
	if err := decodeEndpointGraphQLResponse(stdout, &response, ErrInvalidPullRequestMergeQueueMutationResponse); err != nil {
		return err
	}

	var payload *pullRequestMergeQueueMutationPayload
	switch strings.TrimSpace(mutationField) {
	case "enqueuePullRequest":
		payload = response.EnqueuePullRequest
	case "dequeuePullRequest":
		payload = response.DequeuePullRequest
	default:
		return ErrInvalidPullRequestMergeQueueMutationResponse
	}
	if payload == nil || payload.MergeQueueEntry == nil {
		return ErrInvalidPullRequestMergeQueueMutationResponse
	}
	if _, err := validateNonEmptyPullRequestField(payload.MergeQueueEntry.ID, ErrInvalidPullRequestMergeQueueMutationResponse); err != nil {
		return err
	}
	return nil
}
