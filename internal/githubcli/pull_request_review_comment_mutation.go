package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const updatePullRequestReviewCommentMutation = `mutation($pullRequestReviewCommentId:ID!,$body:String!){updatePullRequestReviewComment(input:{pullRequestReviewCommentId:$pullRequestReviewCommentId,body:$body}){pullRequestReviewComment{id}}}`
const deletePullRequestReviewCommentMutation = `mutation($id:ID!){deletePullRequestReviewComment(input:{id:$id}){pullRequestReviewComment{id}}}`

var ErrInvalidPullRequestReviewCommentMutation = errors.New("invalid pull request review comment mutation")

type updatePullRequestReviewCommentResponse struct {
	Data struct {
		UpdatePullRequestReviewComment *struct {
			PullRequestReviewComment *struct {
				ID string `json:"id"`
			} `json:"pullRequestReviewComment"`
		} `json:"updatePullRequestReviewComment"`
		DeletePullRequestReviewComment *struct {
			PullRequestReviewComment *struct {
				ID string `json:"id"`
			} `json:"pullRequestReviewComment"`
		} `json:"deletePullRequestReviewComment"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *Client) UpdatePullRequestReviewComment(commentID string, body string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return ErrInvalidPullRequestReviewCommentMutation
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestReviewBody); err != nil {
		return err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: updatePullRequestReviewCommentMutation, Variables: []GraphQLVariable{literalGraphQLVariable("pullRequestReviewCommentId", trimmedCommentID), literalGraphQLVariable("body", body)}})
	if err != nil {
		return err
	}

	return parseUpdatedPullRequestReviewComment(result.Stdout, true)
}

func (client *Client) DeletePullRequestReviewComment(commentID string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return ErrInvalidPullRequestReviewCommentMutation
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: deletePullRequestReviewCommentMutation, Variables: []GraphQLVariable{literalGraphQLVariable("id", trimmedCommentID)}})
	if err != nil {
		return err
	}

	return parseUpdatedPullRequestReviewComment(result.Stdout, false)
}

func parseUpdatedPullRequestReviewComment(stdout []byte, updated bool) error {
	var response updatePullRequestReviewCommentResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewCommentMutation, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}

	var commentID string
	if updated {
		if response.Data.UpdatePullRequestReviewComment != nil && response.Data.UpdatePullRequestReviewComment.PullRequestReviewComment != nil {
			commentID = strings.TrimSpace(response.Data.UpdatePullRequestReviewComment.PullRequestReviewComment.ID)
		}
	} else {
		if response.Data.DeletePullRequestReviewComment != nil && response.Data.DeletePullRequestReviewComment.PullRequestReviewComment != nil {
			commentID = strings.TrimSpace(response.Data.DeletePullRequestReviewComment.PullRequestReviewComment.ID)
		}
	}
	if commentID == "" {
		return ErrInvalidPullRequestReviewCommentMutation
	}

	return nil
}
