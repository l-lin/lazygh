package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const submitPullRequestReviewMutation = `mutation($pullRequestReviewId:ID,$event:PullRequestReviewEvent!,$body:String){submitPullRequestReview(input:{pullRequestReviewId:$pullRequestReviewId,event:$event,body:$body}){pullRequestReview{id}}}`

var ErrInvalidPullRequestReviewSubmission = errors.New("invalid pull request review submission")
var ErrInvalidSubmittedPullRequestReviewResponse = errors.New("invalid submitted pull request review response")

type PullRequestReviewEvent string

const (
	PullRequestReviewEventComment        PullRequestReviewEvent = "COMMENT"
	PullRequestReviewEventApprove        PullRequestReviewEvent = "APPROVE"
	PullRequestReviewEventRequestChanges PullRequestReviewEvent = "REQUEST_CHANGES"
)

func (client *ReviewService) SubmitPullRequestReview(pullRequestReviewID string, event PullRequestReviewEvent, body string) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	if trimmedReviewID == "" {
		return ErrInvalidPullRequestReviewSubmission
	}

	normalizedEvent, err := normalizePullRequestReviewEvent(event)
	if err != nil {
		return err
	}

	request := GraphQLRequest{Query: submitPullRequestReviewMutation, Variables: []GraphQLVariable{literalGraphQLVariable("pullRequestReviewId", trimmedReviewID), literalGraphQLVariable("event", string(normalizedEvent))}}
	if strings.TrimSpace(body) != "" {
		request.Variables = append(request.Variables, literalGraphQLVariable("body", body))
	}

	result, err := client.queryGraphQL(request)
	if err != nil {
		return err
	}

	return parseSubmittedPullRequestReview(result.Stdout)
}

func normalizePullRequestReviewEvent(event PullRequestReviewEvent) (PullRequestReviewEvent, error) {
	normalizedEvent := PullRequestReviewEvent(strings.ToUpper(strings.TrimSpace(string(event))))
	switch normalizedEvent {
	case PullRequestReviewEventComment, PullRequestReviewEventApprove, PullRequestReviewEventRequestChanges:
		return normalizedEvent, nil
	default:
		return "", ErrInvalidPullRequestReviewSubmission
	}
}

type submitPullRequestReviewResponse struct {
	Data struct {
		SubmitPullRequestReview *struct {
			PullRequestReview *struct {
				ID string `json:"id"`
			} `json:"pullRequestReview"`
		} `json:"submitPullRequestReview"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func parseSubmittedPullRequestReview(stdout []byte) error {
	var response submitPullRequestReviewResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSubmittedPullRequestReviewResponse, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}
	if response.Data.SubmitPullRequestReview == nil || response.Data.SubmitPullRequestReview.PullRequestReview == nil {
		return ErrInvalidSubmittedPullRequestReviewResponse
	}
	if strings.TrimSpace(response.Data.SubmitPullRequestReview.PullRequestReview.ID) == "" {
		return ErrInvalidSubmittedPullRequestReviewResponse
	}

	return nil
}
