package githubcli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const pullRequestDetailJSONFields = "title,number,url,body,author,state,isDraft,createdAt,updatedAt,labels,baseRefName,headRefName,mergeStateStatus,mergeable,comments,additions,deletions,changedFiles,statusCheckRollup"

var ErrInvalidPullRequestDetailResponse = fmt.Errorf("invalid pull request detail response")

type PullRequestDetail struct {
	Title             string                   `json:"title"`
	Number            int                      `json:"number"`
	URL               string                   `json:"url"`
	Body              string                   `json:"body"`
	Author            *PullRequestAuthor       `json:"author"`
	State             string                   `json:"state"`
	IsDraft           bool                     `json:"isDraft"`
	CreatedAt         string                   `json:"createdAt"`
	UpdatedAt         string                   `json:"updatedAt"`
	Labels            []PullRequestLabel       `json:"labels"`
	BaseRefName       string                   `json:"baseRefName"`
	HeadRefName       string                   `json:"headRefName"`
	MergeStateStatus  string                   `json:"mergeStateStatus"`
	Mergeable         string                   `json:"mergeable"`
	Comments          []PullRequestComment     `json:"comments"`
	Additions         int                      `json:"additions"`
	Deletions         int                      `json:"deletions"`
	ChangedFiles      int                      `json:"changedFiles"`
	StatusCheckRollup []PullRequestStatusCheck `json:"statusCheckRollup"`
}

type PullRequestAuthor struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	IsBot bool   `json:"is_bot"`
}

type PullRequestLabel struct {
	Name string `json:"name"`
}

type PullRequestComment struct {
	Author    *PullRequestCommentAuthor `json:"author"`
	Body      string                    `json:"body"`
	CreatedAt string                    `json:"createdAt"`
	URL       string                    `json:"url"`
}

type PullRequestCommentAuthor struct {
	Login string `json:"login"`
}

type PullRequestStatusCheck struct {
	TypeName     string `json:"__typename"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName"`
}

func (client *Client) GetPullRequestDetail(repository string, number int) (PullRequestDetail, error) {
	result, err := client.runGH("gh pr view", "pr", "view", strconv.Itoa(number), "-R", strings.TrimSpace(repository), "--json", pullRequestDetailJSONFields)
	if err != nil {
		return PullRequestDetail{}, err
	}

	var detail PullRequestDetail
	if err := json.Unmarshal(result.Stdout, &detail); err != nil {
		return PullRequestDetail{}, fmt.Errorf("%w: %v", ErrInvalidPullRequestDetailResponse, err)
	}

	return detail.normalized(), nil
}

func (detail PullRequestDetail) normalized() PullRequestDetail {
	detail.Title = strings.TrimSpace(detail.Title)
	detail.URL = strings.TrimSpace(detail.URL)
	detail.Body = strings.TrimSpace(detail.Body)
	detail.State = strings.TrimSpace(detail.State)
	detail.CreatedAt = strings.TrimSpace(detail.CreatedAt)
	detail.UpdatedAt = strings.TrimSpace(detail.UpdatedAt)
	detail.BaseRefName = strings.TrimSpace(detail.BaseRefName)
	detail.HeadRefName = strings.TrimSpace(detail.HeadRefName)
	detail.MergeStateStatus = strings.TrimSpace(detail.MergeStateStatus)
	detail.Mergeable = strings.TrimSpace(detail.Mergeable)
	if detail.Author != nil {
		normalizedAuthor := detail.Author.normalized()
		detail.Author = &normalizedAuthor
	}
	if len(detail.Labels) > 0 {
		normalizedLabels := make([]PullRequestLabel, 0, len(detail.Labels))
		for _, label := range detail.Labels {
			normalizedLabels = append(normalizedLabels, label.normalized())
		}
		detail.Labels = normalizedLabels
	}
	if len(detail.Comments) > 0 {
		normalizedComments := make([]PullRequestComment, 0, len(detail.Comments))
		for _, comment := range detail.Comments {
			normalizedComments = append(normalizedComments, comment.normalized())
		}
		detail.Comments = normalizedComments
	}
	if len(detail.StatusCheckRollup) > 0 {
		normalizedChecks := make([]PullRequestStatusCheck, 0, len(detail.StatusCheckRollup))
		for _, check := range detail.StatusCheckRollup {
			normalizedChecks = append(normalizedChecks, check.normalized())
		}
		detail.StatusCheckRollup = normalizedChecks
	}
	return detail
}

func (author PullRequestAuthor) normalized() PullRequestAuthor {
	author.Login = strings.TrimSpace(author.Login)
	author.Name = strings.TrimSpace(author.Name)
	return author
}

func (label PullRequestLabel) normalized() PullRequestLabel {
	label.Name = strings.TrimSpace(label.Name)
	return label
}

func (comment PullRequestComment) normalized() PullRequestComment {
	comment.Body = strings.TrimSpace(comment.Body)
	comment.CreatedAt = strings.TrimSpace(comment.CreatedAt)
	comment.URL = strings.TrimSpace(comment.URL)
	if comment.Author != nil {
		normalizedAuthor := comment.Author.normalized()
		comment.Author = &normalizedAuthor
	}
	return comment
}

func (author PullRequestCommentAuthor) normalized() PullRequestCommentAuthor {
	author.Login = strings.TrimSpace(author.Login)
	return author
}

func (check PullRequestStatusCheck) normalized() PullRequestStatusCheck {
	check.TypeName = strings.TrimSpace(check.TypeName)
	check.Name = strings.TrimSpace(check.Name)
	check.Status = strings.TrimSpace(check.Status)
	check.Conclusion = strings.TrimSpace(check.Conclusion)
	check.WorkflowName = strings.TrimSpace(check.WorkflowName)
	return check
}
