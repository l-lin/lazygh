package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const addPullRequestReviewThreadMutation = `mutation($pullRequestReviewId:ID!,$path:String!,$line:Int!,$side:DiffSide!,$body:String!,$startLine:Int,$startSide:DiffSide,$subjectType:PullRequestReviewThreadSubjectType!){addPullRequestReviewThread(input:{pullRequestReviewId:$pullRequestReviewId,path:$path,line:$line,side:$side,body:$body,startLine:$startLine,startSide:$startSide,subjectType:$subjectType}){thread{id}}}`

var ErrInvalidPullRequestReviewThreadResponse = errors.New("invalid pull request review thread response")
var ErrInvalidPullRequestReviewThreadTarget = errors.New("invalid pull request review thread target")

type PullRequestReviewThreadTarget struct {
	Path        string
	Line        int
	Side        string
	StartLine   int
	StartSide   string
	SubjectType string
}

func (client *ReviewService) AddPullRequestReviewThread(pullRequestReviewID string, body string, target PullRequestReviewThreadTarget) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	if trimmedReviewID == "" {
		return ErrInvalidPullRequestReviewThreadTarget
	}
	if _, err := validateNonEmptyPullRequestField(body, ErrEmptyPullRequestReviewBody); err != nil {
		return err
	}

	normalizedTarget, err := normalizePullRequestReviewThreadTarget(target)
	if err != nil {
		return err
	}

	request := GraphQLRequest{Query: addPullRequestReviewThreadMutation, Variables: []GraphQLVariable{
		literalGraphQLVariable("pullRequestReviewId", trimmedReviewID),
		literalGraphQLVariable("path", normalizedTarget.Path),
		typedGraphQLVariable("line", normalizedTarget.Line),
		literalGraphQLVariable("side", normalizedTarget.Side),
		literalGraphQLVariable("body", body),
	}}
	if normalizedTarget.StartLine > 0 {
		request.Variables = append(request.Variables,
			typedGraphQLVariable("startLine", normalizedTarget.StartLine),
			literalGraphQLVariable("startSide", normalizedTarget.StartSide),
		)
	}
	request.Variables = append(request.Variables, literalGraphQLVariable("subjectType", normalizedTarget.SubjectType))

	result, err := client.queryGraphQL(request)
	if err != nil {
		return err
	}

	return parseAddedPullRequestReviewThread(result.Stdout)
}

func normalizePullRequestReviewThreadTarget(target PullRequestReviewThreadTarget) (PullRequestReviewThreadTarget, error) {
	normalized := PullRequestReviewThreadTarget{
		Path:        strings.TrimSpace(target.Path),
		Line:        target.Line,
		Side:        strings.ToUpper(strings.TrimSpace(target.Side)),
		StartLine:   target.StartLine,
		StartSide:   strings.ToUpper(strings.TrimSpace(target.StartSide)),
		SubjectType: strings.ToUpper(strings.TrimSpace(target.SubjectType)),
	}
	if normalized.Path == "" || normalized.Line <= 0 || normalized.SubjectType == "" {
		return PullRequestReviewThreadTarget{}, ErrInvalidPullRequestReviewThreadTarget
	}
	if normalized.Side != "LEFT" && normalized.Side != "RIGHT" {
		return PullRequestReviewThreadTarget{}, ErrInvalidPullRequestReviewThreadTarget
	}
	if normalized.SubjectType != "LINE" && normalized.SubjectType != "FILE" {
		return PullRequestReviewThreadTarget{}, ErrInvalidPullRequestReviewThreadTarget
	}
	if normalized.StartLine < 0 {
		return PullRequestReviewThreadTarget{}, ErrInvalidPullRequestReviewThreadTarget
	}
	if normalized.StartLine == 0 {
		normalized.StartSide = ""
		return normalized, nil
	}
	if normalized.StartSide != "LEFT" && normalized.StartSide != "RIGHT" {
		return PullRequestReviewThreadTarget{}, ErrInvalidPullRequestReviewThreadTarget
	}
	if normalized.StartLine > normalized.Line {
		return PullRequestReviewThreadTarget{}, ErrInvalidPullRequestReviewThreadTarget
	}

	return normalized, nil
}

type addPullRequestReviewThreadResponse struct {
	Data struct {
		AddPullRequestReviewThread *struct {
			Thread *struct {
				ID string `json:"id"`
			} `json:"thread"`
		} `json:"addPullRequestReviewThread"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func parseAddedPullRequestReviewThread(stdout []byte) error {
	var response addPullRequestReviewThreadResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPullRequestReviewThreadResponse, err)
	}
	for _, graphqlErr := range response.Errors {
		message := strings.TrimSpace(graphqlErr.Message)
		if message != "" {
			return errors.New(message)
		}
	}
	if response.Data.AddPullRequestReviewThread == nil || response.Data.AddPullRequestReviewThread.Thread == nil {
		return ErrInvalidPullRequestReviewThreadResponse
	}
	if strings.TrimSpace(response.Data.AddPullRequestReviewThread.Thread.ID) == "" {
		return ErrInvalidPullRequestReviewThreadResponse
	}

	return nil
}
