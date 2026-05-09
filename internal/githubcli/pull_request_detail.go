package githubcli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const pullRequestDetailJSONFields = "title,number,url,body,author,state,isDraft,createdAt,updatedAt,labels,assignees,reviewRequests,baseRefName,headRefName,mergeStateStatus,mergeable,comments,commits,reviews,additions,deletions,changedFiles,statusCheckRollup"

var (
	ErrInvalidPullRequestDetailResponse        = fmt.Errorf("invalid pull request detail response")
	ErrInvalidPullRequestInlineCommentResponse = fmt.Errorf("invalid pull request inline comment response")
)

type PullRequestDetail struct {
	ID                   string                     `json:"id,omitempty"`
	Title                string                     `json:"title"`
	Number               int                        `json:"number"`
	URL                  string                     `json:"url"`
	Body                 string                     `json:"body"`
	BodyHTML             string                     `json:"bodyHTML,omitempty"`
	Author               *PullRequestAuthor         `json:"author"`
	State                string                     `json:"state"`
	IsDraft              bool                       `json:"isDraft"`
	CreatedAt            string                     `json:"createdAt"`
	UpdatedAt            string                     `json:"updatedAt"`
	Labels               []PullRequestLabel         `json:"labels"`
	Assignees            []PullRequestAuthor        `json:"assignees"`
	ReviewRequests       []PullRequestReviewRequest `json:"reviewRequests"`
	BaseRefName          string                     `json:"baseRefName"`
	HeadRefName          string                     `json:"headRefName"`
	MergeStateStatus     string                     `json:"mergeStateStatus"`
	Mergeable            string                     `json:"mergeable"`
	ReactionGroups       []ReactionGroup            `json:"reactionGroups,omitempty"`
	Comments             []PullRequestComment       `json:"comments"`
	Commits              []PullRequestCommit        `json:"commits"`
	Reviews              []PullRequestReview        `json:"reviews"`
	InlineComments       []PullRequestInlineComment `json:"-"`
	InlineCommentThreads []PullRequestReviewThread  `json:"-"`
	Additions            int                        `json:"additions"`
	Deletions            int                        `json:"deletions"`
	ChangedFiles         int                        `json:"changedFiles"`
	StatusCheckRollup    []PullRequestStatusCheck   `json:"statusCheckRollup"`
}

type PullRequestAuthor struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	IsBot bool   `json:"is_bot"`
}

type PullRequestLabel struct {
	Name string `json:"name"`
}

type PullRequestReviewRequest struct {
	RequestedReviewer PullRequestRequestedReviewer `json:"requestedReviewer"`
}

func (reviewRequest *PullRequestReviewRequest) UnmarshalJSON(data []byte) error {
	var wrapped struct {
		RequestedReviewer *PullRequestRequestedReviewer `json:"requestedReviewer"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.RequestedReviewer != nil {
		reviewRequest.RequestedReviewer = wrapped.RequestedReviewer.normalized()
		return nil
	}

	var direct PullRequestRequestedReviewer
	if err := json.Unmarshal(data, &direct); err != nil {
		return err
	}
	reviewRequest.RequestedReviewer = direct.normalized()
	return nil
}

type PullRequestRequestedReviewer struct {
	TypeName     string                                `json:"__typename"`
	Login        string                                `json:"login"`
	Name         string                                `json:"name"`
	Slug         string                                `json:"slug"`
	Organization *PullRequestReviewRequestOrganization `json:"organization"`
}

type PullRequestReviewRequestOrganization struct {
	Login string `json:"login"`
}

type PullRequestComment struct {
	ID              string                    `json:"id"`
	Author          *PullRequestCommentAuthor `json:"author"`
	Body            string                    `json:"body"`
	BodyHTML        string                    `json:"bodyHTML,omitempty"`
	CreatedAt       string                    `json:"createdAt"`
	URL             string                    `json:"url"`
	DiffHunk        string                    `json:"diffHunk"`
	State           string                    `json:"state"`
	ViewerDidAuthor bool                      `json:"viewerDidAuthor"`
	ReactionGroups  []ReactionGroup           `json:"reactionGroups,omitempty"`
}

type PullRequestCommit struct {
	OID             string                    `json:"oid"`
	MessageHeadline string                    `json:"messageHeadline"`
	MessageBody     string                    `json:"messageBody"`
	MessageBodyHTML string                    `json:"messageBodyHTML,omitempty"`
	AuthoredDate    string                    `json:"authoredDate"`
	CommittedDate   string                    `json:"committedDate"`
	Authors         []PullRequestCommitAuthor `json:"authors"`
}

type PullRequestCommitAuthor struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type PullRequestInlineComment struct {
	ID                string                    `json:"node_id"`
	Author            *PullRequestCommentAuthor `json:"user"`
	Body              string                    `json:"body"`
	BodyHTML          string                    `json:"bodyHTML,omitempty"`
	CreatedAt         string                    `json:"created_at"`
	URL               string                    `json:"html_url"`
	Path              string                    `json:"path"`
	DiffHunk          string                    `json:"diff_hunk"`
	Line              int                       `json:"line"`
	OriginalLine      int                       `json:"original_line"`
	StartLine         int                       `json:"start_line"`
	OriginalStartLine int                       `json:"original_start_line"`
	Side              string                    `json:"side"`
	StartSide         string                    `json:"start_side"`
	SubjectType       string                    `json:"subject_type"`
	ReactionGroups    []ReactionGroup           `json:"reactionGroups,omitempty"`
}

type PullRequestCommentAuthor struct {
	Login string `json:"login"`
}

type PullRequestReview struct {
	Author      *PullRequestCommentAuthor `json:"author"`
	State       string                    `json:"state"`
	SubmittedAt string                    `json:"submittedAt"`
}

type PullRequestStatusCheck struct {
	TypeName     string `json:"__typename"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	WorkflowName string `json:"workflowName"`
	Link         string `json:"link,omitempty"`
}

func (client *Client) GetPullRequestDetail(repository string, number int) (PullRequestDetail, error) {
	trimmedRepository := strings.TrimSpace(repository)
	result, err := client.runGH("gh pr view", "pr", "view", strconv.Itoa(number), "-R", trimmedRepository, "--json", pullRequestDetailJSONFields)
	if err != nil {
		return PullRequestDetail{}, err
	}

	var detail PullRequestDetail
	if err := json.Unmarshal(result.Stdout, &detail); err != nil {
		return PullRequestDetail{}, fmt.Errorf("%w: %v", ErrInvalidPullRequestDetailResponse, err)
	}
	if len(detail.StatusCheckRollup) > 0 {
		if buildInfos, buildErr := client.listPullRequestBuildInfos(trimmedRepository, number); buildErr == nil {
			detail.StatusCheckRollup = mergePullRequestStatusCheckLinks(detail.StatusCheckRollup, buildInfos)
		}
	}

	inlineComments, err := client.listPullRequestInlineComments(trimmedRepository, number)
	if err != nil {
		return PullRequestDetail{}, err
	}
	if len(inlineComments) > 0 {
		detail.InlineComments = inlineComments
	}

	inlineThreads, err := client.listPullRequestReviewThreads(trimmedRepository, number)
	if err != nil {
		return PullRequestDetail{}, err
	}
	if len(inlineThreads) > 0 {
		detail.InlineCommentThreads = inlineThreads
	}

	reactionTargets, err := client.listPullRequestReactionTargets(trimmedRepository, number)
	if err != nil {
		return PullRequestDetail{}, err
	}
	if strings.TrimSpace(reactionTargets.PullRequestID) != "" {
		detail.ID = reactionTargets.PullRequestID
	}
	if len(reactionTargets.ReactionGroups) > 0 {
		detail.ReactionGroups = reactionTargets.ReactionGroups
	}
	if len(reactionTargets.Comments) > 0 {
		detail.Comments = reactionTargets.Comments
	}

	inlineCommentReactionGroups, err := client.listPullRequestReviewCommentReactionGroups(pullRequestInlineCommentReactionTargetIDs(detail.InlineComments))
	if err != nil {
		return PullRequestDetail{}, err
	}
	if len(inlineCommentReactionGroups) > 0 {
		detail.InlineComments = mergePullRequestInlineCommentReactionGroups(detail.InlineComments, inlineCommentReactionGroups)
	}

	return detail.normalized(), nil
}

func (client *Client) listPullRequestInlineComments(repository string, number int) ([]PullRequestInlineComment, error) {
	result, err := client.runGH("gh api pull request inline comments", "api", fmt.Sprintf("repos/%s/pulls/%d/comments?per_page=100", strings.TrimSpace(repository), number), "--paginate", "--slurp")
	if err != nil {
		return nil, err
	}

	var pagedComments [][]PullRequestInlineComment
	if err := json.Unmarshal(result.Stdout, &pagedComments); err != nil {
		var comments []PullRequestInlineComment
		if err := json.Unmarshal(result.Stdout, &comments); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPullRequestInlineCommentResponse, err)
		}
		return comments, nil
	}

	flattenedComments := make([]PullRequestInlineComment, 0)
	for _, page := range pagedComments {
		flattenedComments = append(flattenedComments, page...)
	}
	return flattenedComments, nil
}

func pullRequestInlineCommentReactionTargetIDs(comments []PullRequestInlineComment) []string {
	ids := make([]string, 0, len(comments))
	for _, comment := range comments {
		ids = append(ids, strings.TrimSpace(comment.ID))
	}
	return ids
}

func mergePullRequestInlineCommentReactionGroups(comments []PullRequestInlineComment, groupsByID map[string][]ReactionGroup) []PullRequestInlineComment {
	if len(comments) == 0 || len(groupsByID) == 0 {
		return comments
	}

	mergedComments := make([]PullRequestInlineComment, 0, len(comments))
	for _, comment := range comments {
		if reactionGroups, ok := groupsByID[strings.TrimSpace(comment.ID)]; ok {
			comment.ReactionGroups = append([]ReactionGroup(nil), reactionGroups...)
		}
		mergedComments = append(mergedComments, comment)
	}
	return mergedComments
}

func (detail PullRequestDetail) normalized() PullRequestDetail {
	detail.ID = strings.TrimSpace(detail.ID)
	detail.Title = strings.TrimSpace(detail.Title)
	detail.URL = strings.TrimSpace(detail.URL)
	detail.Body = strings.TrimSpace(detail.Body)
	detail.BodyHTML = strings.TrimSpace(detail.BodyHTML)
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
	if len(detail.Assignees) > 0 {
		normalizedAssignees := make([]PullRequestAuthor, 0, len(detail.Assignees))
		for _, assignee := range detail.Assignees {
			normalizedAssignees = append(normalizedAssignees, assignee.normalized())
		}
		detail.Assignees = normalizedAssignees
	}
	if len(detail.ReviewRequests) > 0 {
		normalizedReviewRequests := make([]PullRequestReviewRequest, 0, len(detail.ReviewRequests))
		for _, reviewRequest := range detail.ReviewRequests {
			normalizedReviewRequests = append(normalizedReviewRequests, reviewRequest.normalized())
		}
		detail.ReviewRequests = normalizedReviewRequests
	}
	detail.ReactionGroups = normalizeReactionGroups(detail.ReactionGroups)
	if len(detail.Comments) > 0 {
		normalizedComments := make([]PullRequestComment, 0, len(detail.Comments))
		for _, comment := range detail.Comments {
			normalizedComments = append(normalizedComments, comment.normalized())
		}
		detail.Comments = normalizedComments
	}
	if len(detail.Commits) > 0 {
		normalizedCommits := make([]PullRequestCommit, 0, len(detail.Commits))
		for _, commit := range detail.Commits {
			normalizedCommits = append(normalizedCommits, commit.normalized())
		}
		detail.Commits = normalizedCommits
	}
	if len(detail.Reviews) > 0 {
		normalizedReviews := make([]PullRequestReview, 0, len(detail.Reviews))
		for _, review := range detail.Reviews {
			normalizedReviews = append(normalizedReviews, review.normalized())
		}
		detail.Reviews = normalizedReviews
	}
	if len(detail.InlineComments) > 0 {
		normalizedInlineComments := make([]PullRequestInlineComment, 0, len(detail.InlineComments))
		for _, comment := range detail.InlineComments {
			normalizedInlineComments = append(normalizedInlineComments, comment.normalized())
		}
		detail.InlineComments = normalizedInlineComments
	}
	if len(detail.InlineCommentThreads) > 0 {
		normalizedInlineThreads := make([]PullRequestReviewThread, 0, len(detail.InlineCommentThreads))
		for _, thread := range detail.InlineCommentThreads {
			normalizedInlineThreads = append(normalizedInlineThreads, thread.normalized())
		}
		detail.InlineCommentThreads = normalizedInlineThreads
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

func (reviewRequest PullRequestReviewRequest) normalized() PullRequestReviewRequest {
	reviewRequest.RequestedReviewer = reviewRequest.RequestedReviewer.normalized()
	return reviewRequest
}

func (reviewer PullRequestRequestedReviewer) normalized() PullRequestRequestedReviewer {
	reviewer.TypeName = strings.TrimSpace(reviewer.TypeName)
	reviewer.Login = strings.TrimSpace(reviewer.Login)
	reviewer.Name = strings.TrimSpace(reviewer.Name)
	reviewer.Slug = strings.TrimSpace(reviewer.Slug)
	if reviewer.Organization != nil {
		normalizedOrganization := reviewer.Organization.normalized()
		reviewer.Organization = &normalizedOrganization
	}
	return reviewer
}

func (organization PullRequestReviewRequestOrganization) normalized() PullRequestReviewRequestOrganization {
	organization.Login = strings.TrimSpace(organization.Login)
	return organization
}

func (comment PullRequestComment) normalized() PullRequestComment {
	comment.ID = strings.TrimSpace(comment.ID)
	comment.Body = strings.TrimSpace(comment.Body)
	comment.BodyHTML = strings.TrimSpace(comment.BodyHTML)
	comment.CreatedAt = strings.TrimSpace(comment.CreatedAt)
	comment.URL = strings.TrimSpace(comment.URL)
	comment.DiffHunk = strings.TrimSpace(comment.DiffHunk)
	comment.State = strings.TrimSpace(comment.State)
	comment.ReactionGroups = normalizeReactionGroups(comment.ReactionGroups)
	if comment.Author != nil {
		normalizedAuthor := comment.Author.normalized()
		comment.Author = &normalizedAuthor
	}
	return comment
}

func (commit PullRequestCommit) normalized() PullRequestCommit {
	commit.OID = strings.TrimSpace(commit.OID)
	commit.MessageHeadline = strings.TrimSpace(commit.MessageHeadline)
	commit.MessageBody = strings.TrimSpace(commit.MessageBody)
	commit.MessageBodyHTML = strings.TrimSpace(commit.MessageBodyHTML)
	commit.AuthoredDate = strings.TrimSpace(commit.AuthoredDate)
	commit.CommittedDate = strings.TrimSpace(commit.CommittedDate)
	if len(commit.Authors) > 0 {
		normalizedAuthors := make([]PullRequestCommitAuthor, 0, len(commit.Authors))
		for _, author := range commit.Authors {
			normalizedAuthors = append(normalizedAuthors, author.normalized())
		}
		commit.Authors = normalizedAuthors
	}
	return commit
}

func (author PullRequestCommitAuthor) normalized() PullRequestCommitAuthor {
	author.Login = strings.TrimSpace(author.Login)
	author.Name = strings.TrimSpace(author.Name)
	author.Email = strings.TrimSpace(author.Email)
	return author
}

func (comment PullRequestInlineComment) normalized() PullRequestInlineComment {
	comment.ID = strings.TrimSpace(comment.ID)
	comment.Body = strings.TrimSpace(comment.Body)
	comment.BodyHTML = strings.TrimSpace(comment.BodyHTML)
	comment.CreatedAt = strings.TrimSpace(comment.CreatedAt)
	comment.URL = strings.TrimSpace(comment.URL)
	comment.Path = strings.TrimSpace(comment.Path)
	comment.DiffHunk = strings.TrimSpace(comment.DiffHunk)
	comment.Side = strings.TrimSpace(comment.Side)
	comment.StartSide = strings.TrimSpace(comment.StartSide)
	comment.SubjectType = strings.TrimSpace(comment.SubjectType)
	comment.ReactionGroups = normalizeReactionGroups(comment.ReactionGroups)
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

func (review PullRequestReview) normalized() PullRequestReview {
	review.State = strings.TrimSpace(review.State)
	review.SubmittedAt = strings.TrimSpace(review.SubmittedAt)
	if review.Author != nil {
		normalizedAuthor := review.Author.normalized()
		review.Author = &normalizedAuthor
	}
	return review
}

func (check PullRequestStatusCheck) normalized() PullRequestStatusCheck {
	check.TypeName = strings.TrimSpace(check.TypeName)
	check.Name = strings.TrimSpace(check.Name)
	check.Status = strings.TrimSpace(check.Status)
	check.Conclusion = strings.TrimSpace(check.Conclusion)
	check.WorkflowName = strings.TrimSpace(check.WorkflowName)
	check.Link = strings.TrimSpace(check.Link)
	return check
}
