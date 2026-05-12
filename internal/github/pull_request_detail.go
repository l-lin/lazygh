package github

import "strings"

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
	InlineCommentThreads []ReviewThread             `json:"-"`
	Additions            int                        `json:"additions"`
	Deletions            int                        `json:"deletions"`
	ChangedFiles         int                        `json:"changedFiles"`
	StatusCheckRollup    []BuildInfo                `json:"statusCheckRollup"`
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
		normalizedInlineThreads := make([]ReviewThread, 0, len(detail.InlineCommentThreads))
		for _, thread := range detail.InlineCommentThreads {
			normalizedInlineThreads = append(normalizedInlineThreads, thread.normalized())
		}
		detail.InlineCommentThreads = normalizedInlineThreads
	}
	if len(detail.StatusCheckRollup) > 0 {
		normalizedChecks := make([]BuildInfo, 0, len(detail.StatusCheckRollup))
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
