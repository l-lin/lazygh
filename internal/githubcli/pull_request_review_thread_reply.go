package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const addPullRequestReviewThreadReplyMutation = `mutation($pullRequestReviewThreadId:ID!,$body:String!,$pullRequestReviewId:ID){addPullRequestReviewThreadReply(input:{pullRequestReviewThreadId:$pullRequestReviewThreadId,body:$body,pullRequestReviewId:$pullRequestReviewId}){comment{id}}}`

var ErrInvalidPullRequestReviewThreadReply = errors.New("invalid pull request review thread reply")

type addPullRequestReviewThreadReplyResponse struct {
	Data struct {
		AddPullRequestReviewThreadReply *struct {
			Comment *struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"addPullRequestReviewThreadReply"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (client *Client) AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error {
	trimmedThreadID := strings.TrimSpace(pullRequestReviewThreadID)
	if trimmedThreadID == "" {
		return ErrInvalidPullRequestReviewThreadReply
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestReviewBody); err != nil {
		return err
	}

	args := []string{
		"api",
		"graphql",
		"-f",
		"query=" + addPullRequestReviewThreadReplyMutation,
		"-f",
		"pullRequestReviewThreadId=" + trimmedThreadID,
		"-f",
		"body=" + body,
	}
	if trimmedReviewID := strings.TrimSpace(pullRequestReviewID); trimmedReviewID != "" {
		args = append(args,
			"-f",
			"pullRequestReviewId="+trimmedReviewID,
		)
	}

	result, err := client.runGH("gh api graphql", args...)
	if err != nil {
		return err
	}

	return parseAddedPullRequestReviewThreadReply(result.Stdout)
}

func parseAddedPullRequestReviewThreadReply(stdout []byte) error {
	var response addPullRequestReviewThreadReplyResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewThreadReply, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}
	if response.Data.AddPullRequestReviewThreadReply == nil || response.Data.AddPullRequestReviewThreadReply.Comment == nil {
		return ErrInvalidPullRequestReviewThreadReply
	}
	if strings.TrimSpace(response.Data.AddPullRequestReviewThreadReply.Comment.ID) == "" {
		return ErrInvalidPullRequestReviewThreadReply
	}

	return nil
}
