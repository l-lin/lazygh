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
	OutOfDateWithBase    bool                       `json:"outOfDateWithBase,omitempty"`
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

func (author PullRequestCommentAuthor) normalized() PullRequestCommentAuthor {
	author.Login = strings.TrimSpace(author.Login)
	return author
}
