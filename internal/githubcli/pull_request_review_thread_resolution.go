package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const resolveReviewThreadMutation = `mutation($threadId:ID!){resolveReviewThread(input:{threadId:$threadId}){thread{id isResolved}}}`
const unresolveReviewThreadMutation = `mutation($threadId:ID!){unresolveReviewThread(input:{threadId:$threadId}){thread{id isResolved}}}`

var ErrInvalidPullRequestReviewThreadMutation = errors.New("invalid pull request review thread mutation")

func (client *Client) ResolvePullRequestReviewThread(threadID string) error {
	return client.updatePullRequestReviewThreadResolution(threadID, true)
}

func (client *Client) UnresolvePullRequestReviewThread(threadID string) error {
	return client.updatePullRequestReviewThreadResolution(threadID, false)
}

func (client *Client) updatePullRequestReviewThreadResolution(threadID string, resolved bool) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return ErrInvalidPullRequestReviewThreadMutation
	}

	mutation := resolveReviewThreadMutation
	responseField := "resolveReviewThread"
	if !resolved {
		mutation = unresolveReviewThreadMutation
		responseField = "unresolveReviewThread"
	}

	result, err := client.runGH(
		"gh api graphql",
		"api",
		"graphql",
		"-f",
		"query="+mutation,
		"-f",
		"threadId="+trimmedThreadID,
	)
	if err != nil {
		return err
	}

	return parseUpdatedPullRequestReviewThreadResolution(result.Stdout, responseField)
}

type updatePullRequestReviewThreadResolutionResponse struct {
	Data struct {
		ResolveReviewThread *struct {
			Thread *struct {
				ID string `json:"id"`
			} `json:"thread"`
		} `json:"resolveReviewThread"`
		UnresolveReviewThread *struct {
			Thread *struct {
				ID string `json:"id"`
			} `json:"thread"`
		} `json:"unresolveReviewThread"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func parseUpdatedPullRequestReviewThreadResolution(stdout []byte, responseField string) error {
	var response updatePullRequestReviewThreadResolutionResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewThreadMutation, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}

	var threadID string
	switch responseField {
	case "resolveReviewThread":
		if response.Data.ResolveReviewThread != nil && response.Data.ResolveReviewThread.Thread != nil {
			threadID = strings.TrimSpace(response.Data.ResolveReviewThread.Thread.ID)
		}
	case "unresolveReviewThread":
		if response.Data.UnresolveReviewThread != nil && response.Data.UnresolveReviewThread.Thread != nil {
			threadID = strings.TrimSpace(response.Data.UnresolveReviewThread.Thread.ID)
		}
	default:
		return ErrInvalidPullRequestReviewThreadMutation
	}
	if threadID == "" {
		return ErrInvalidPullRequestReviewThreadMutation
	}

	return nil
}
