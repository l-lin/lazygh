package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const deletePullRequestReviewMutation = `mutation($pullRequestReviewId:ID!){deletePullRequestReview(input:{pullRequestReviewId:$pullRequestReviewId}){pullRequestReview{id}}}`

var ErrInvalidPullRequestReviewDeletion = errors.New("invalid pull request review deletion")

type deletePullRequestReviewResponse struct {
	Data struct {
		DeletePullRequestReview *struct {
			PullRequestReview *struct {
				ID string `json:"id"`
			} `json:"pullRequestReview"`
		} `json:"deletePullRequestReview"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *Client) DeletePullRequestReview(pullRequestReviewID string) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	if trimmedReviewID == "" {
		return ErrInvalidPullRequestReviewDeletion
	}

	result, err := client.runGH(
		"gh api graphql",
		"api",
		"graphql",
		"-f",
		"query="+deletePullRequestReviewMutation,
		"-f",
		"pullRequestReviewId="+trimmedReviewID,
	)
	if err != nil {
		return err
	}

	return parseDeletedPullRequestReview(result.Stdout)
}

func parseDeletedPullRequestReview(stdout []byte) error {
	var response deletePullRequestReviewResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewDeletion, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}
	if response.Data.DeletePullRequestReview == nil || response.Data.DeletePullRequestReview.PullRequestReview == nil {
		return ErrInvalidPullRequestReviewDeletion
	}
	if strings.TrimSpace(response.Data.DeletePullRequestReview.PullRequestReview.ID) == "" {
		return ErrInvalidPullRequestReviewDeletion
	}

	return nil
}
