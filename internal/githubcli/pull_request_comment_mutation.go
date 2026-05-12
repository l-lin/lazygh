package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const updatePullRequestCommentMutation = `mutation($id:ID!,$body:String!){updateIssueComment(input:{id:$id,body:$body}){issueComment{id}}}`
const deletePullRequestCommentMutation = `mutation($id:ID!){deleteIssueComment(input:{id:$id}){clientMutationId}}`

var ErrInvalidPullRequestCommentMutation = errors.New("invalid pull request comment mutation")

type pullRequestCommentMutationResponse struct {
	Data struct {
		UpdateIssueComment *struct {
			IssueComment *struct {
				ID string `json:"id"`
			} `json:"issueComment"`
		} `json:"updateIssueComment"`
		DeleteIssueComment *struct {
			ClientMutationID string `json:"clientMutationId"`
		} `json:"deleteIssueComment"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *PullRequestMutationService) UpdatePullRequestComment(commentID string, body string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return ErrInvalidPullRequestCommentMutation
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestComment); err != nil {
		return err
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: updatePullRequestCommentMutation, Variables: []GraphQLVariable{literalGraphQLVariable("id", trimmedCommentID), literalGraphQLVariable("body", body)}})
	if err != nil {
		return err
	}

	return parsePullRequestCommentMutation(result.Stdout, true)
}

func (client *PullRequestMutationService) DeletePullRequestComment(commentID string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return ErrInvalidPullRequestCommentMutation
	}

	result, err := client.queryGraphQL(GraphQLRequest{Query: deletePullRequestCommentMutation, Variables: []GraphQLVariable{literalGraphQLVariable("id", trimmedCommentID)}})
	if err != nil {
		return err
	}

	return parsePullRequestCommentMutation(result.Stdout, false)
}

func parsePullRequestCommentMutation(stdout []byte, updated bool) error {
	var response pullRequestCommentMutationResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPullRequestCommentMutation, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}

	if updated {
		if response.Data.UpdateIssueComment == nil || response.Data.UpdateIssueComment.IssueComment == nil || strings.TrimSpace(response.Data.UpdateIssueComment.IssueComment.ID) == "" {
			return ErrInvalidPullRequestCommentMutation
		}
		return nil
	}
	if response.Data.DeleteIssueComment == nil {
		return ErrInvalidPullRequestCommentMutation
	}
	return nil
}
