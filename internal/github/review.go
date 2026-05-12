package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidReviewEvent        = errors.New("invalid pull request review submission")
	ErrInvalidReviewThreadTarget = errors.New("invalid pull request review thread target")
)

type ReviewThread struct {
	ID                 string               `json:"id"`
	IsResolved         bool                 `json:"isResolved"`
	IsOutdated         bool                 `json:"isOutdated"`
	ViewerCanResolve   bool                 `json:"viewerCanResolve"`
	ViewerCanUnresolve bool                 `json:"viewerCanUnresolve"`
	Path               string               `json:"path"`
	Line               int                  `json:"line"`
	OriginalLine       int                  `json:"originalLine"`
	StartLine          int                  `json:"startLine"`
	OriginalStartLine  int                  `json:"originalStartLine"`
	DiffSide           string               `json:"diffSide"`
	StartDiffSide      string               `json:"startDiffSide"`
	Comments           []PullRequestComment `json:"-"`
}

type PullRequestReviewThread = ReviewThread

type ReviewThreadTarget struct {
	Path        string
	Line        int
	Side        string
	StartLine   int
	StartSide   string
	SubjectType string
}

type PullRequestReviewThreadTarget = ReviewThreadTarget

type ReviewEvent string

type PullRequestReviewEvent = ReviewEvent

const (
	ReviewEventComment        ReviewEvent = "COMMENT"
	ReviewEventApprove        ReviewEvent = "APPROVE"
	ReviewEventRequestChanges ReviewEvent = "REQUEST_CHANGES"

	PullRequestReviewEventComment        PullRequestReviewEvent = ReviewEventComment
	PullRequestReviewEventApprove        PullRequestReviewEvent = ReviewEventApprove
	PullRequestReviewEventRequestChanges PullRequestReviewEvent = ReviewEventRequestChanges
)

func NormalizeReviewEvent(event ReviewEvent) (ReviewEvent, error) {
	normalizedEvent := ReviewEvent(strings.ToUpper(strings.TrimSpace(string(event))))
	switch normalizedEvent {
	case ReviewEventComment, ReviewEventApprove, ReviewEventRequestChanges:
		return normalizedEvent, nil
	default:
		return "", ErrInvalidReviewEvent
	}
}

func NormalizeReviewThreadTarget(target ReviewThreadTarget) (ReviewThreadTarget, error) {
	normalized := ReviewThreadTarget{
		Path:        strings.TrimSpace(target.Path),
		Line:        target.Line,
		Side:        strings.ToUpper(strings.TrimSpace(target.Side)),
		StartLine:   target.StartLine,
		StartSide:   strings.ToUpper(strings.TrimSpace(target.StartSide)),
		SubjectType: strings.ToUpper(strings.TrimSpace(target.SubjectType)),
	}
	if normalized.Path == "" || normalized.Line <= 0 || normalized.SubjectType == "" {
		return ReviewThreadTarget{}, ErrInvalidReviewThreadTarget
	}
	if normalized.Side != "LEFT" && normalized.Side != "RIGHT" {
		return ReviewThreadTarget{}, ErrInvalidReviewThreadTarget
	}
	if normalized.SubjectType != "LINE" && normalized.SubjectType != "FILE" {
		return ReviewThreadTarget{}, ErrInvalidReviewThreadTarget
	}
	if normalized.StartLine < 0 {
		return ReviewThreadTarget{}, ErrInvalidReviewThreadTarget
	}
	if normalized.StartLine == 0 {
		normalized.StartSide = ""
		return normalized, nil
	}
	if normalized.StartSide != "LEFT" && normalized.StartSide != "RIGHT" {
		return ReviewThreadTarget{}, ErrInvalidReviewThreadTarget
	}
	if normalized.StartLine > normalized.Line {
		return ReviewThreadTarget{}, ErrInvalidReviewThreadTarget
	}
	return normalized, nil
}

func (thread ReviewThread) normalized() ReviewThread {
	thread.ID = strings.TrimSpace(thread.ID)
	thread.Path = strings.TrimSpace(thread.Path)
	thread.DiffSide = strings.ToUpper(strings.TrimSpace(thread.DiffSide))
	thread.StartDiffSide = strings.ToUpper(strings.TrimSpace(thread.StartDiffSide))
	thread.Comments = normalizePullRequestComments(thread.Comments)
	return thread
}

func normalizePullRequestComments(comments []PullRequestComment) []PullRequestComment {
	normalizedComments := make([]PullRequestComment, 0, len(comments))
	for _, comment := range comments {
		normalizedComments = append(normalizedComments, comment.normalized())
	}
	return normalizedComments
}

// Transport adapters still decode GraphQL thread payloads through shared models.
type reviewThreadsPage struct {
	Threads     []reviewThreadPageNode
	HasNextPage bool
	EndCursor   string
}

type reviewThreadPageNode struct {
	Thread              ReviewThread
	CommentsHasNextPage bool
	CommentsEndCursor   string
}

type reviewThreadsResponse struct {
	Data struct {
		Repository *struct {
			PullRequest *struct {
				ReviewThreads struct {
					PageInfo struct {
						HasNextPage bool   `json:"hasNextPage"`
						EndCursor   string `json:"endCursor"`
					} `json:"pageInfo"`
					Nodes []struct {
						ID                 string `json:"id"`
						IsResolved         bool   `json:"isResolved"`
						IsOutdated         bool   `json:"isOutdated"`
						ViewerCanResolve   bool   `json:"viewerCanResolve"`
						ViewerCanUnresolve bool   `json:"viewerCanUnresolve"`
						Path               string `json:"path"`
						Line               int    `json:"line"`
						OriginalLine       int    `json:"originalLine"`
						StartLine          int    `json:"startLine"`
						OriginalStartLine  int    `json:"originalStartLine"`
						DiffSide           string `json:"diffSide"`
						StartDiffSide      string `json:"startDiffSide"`
						Comments           struct {
							PageInfo struct {
								HasNextPage bool   `json:"hasNextPage"`
								EndCursor   string `json:"endCursor"`
							} `json:"pageInfo"`
							Nodes []PullRequestComment `json:"nodes"`
						} `json:"comments"`
					} `json:"nodes"`
				} `json:"reviewThreads"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

func ParseReviewThreadsPage(stdout []byte) (reviewThreadsPage, error) {
	var response reviewThreadsResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return reviewThreadsPage{}, fmt.Errorf("decode review threads page: %w", err)
	}
	if response.Data.Repository == nil || response.Data.Repository.PullRequest == nil {
		return reviewThreadsPage{}, errors.New("missing review threads payload")
	}

	reviewThreads := response.Data.Repository.PullRequest.ReviewThreads
	page := reviewThreadsPage{
		Threads:     make([]reviewThreadPageNode, 0, len(reviewThreads.Nodes)),
		HasNextPage: reviewThreads.PageInfo.HasNextPage,
		EndCursor:   strings.TrimSpace(reviewThreads.PageInfo.EndCursor),
	}
	for _, node := range reviewThreads.Nodes {
		page.Threads = append(page.Threads, reviewThreadPageNode{
			Thread: ReviewThread{
				ID:                 node.ID,
				IsResolved:         node.IsResolved,
				IsOutdated:         node.IsOutdated,
				ViewerCanResolve:   node.ViewerCanResolve,
				ViewerCanUnresolve: node.ViewerCanUnresolve,
				Path:               node.Path,
				Line:               node.Line,
				OriginalLine:       node.OriginalLine,
				StartLine:          node.StartLine,
				OriginalStartLine:  node.OriginalStartLine,
				DiffSide:           node.DiffSide,
				StartDiffSide:      node.StartDiffSide,
				Comments:           normalizePullRequestComments(node.Comments.Nodes),
			},
			CommentsHasNextPage: node.Comments.PageInfo.HasNextPage,
			CommentsEndCursor:   strings.TrimSpace(node.Comments.PageInfo.EndCursor),
		})
	}
	return page, nil
}

type reviewThreadCommentsPage struct {
	Comments    []PullRequestComment
	HasNextPage bool
	EndCursor   string
}

type reviewThreadCommentsResponse struct {
	Data *struct {
		Node *struct {
			Comments struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []PullRequestComment `json:"nodes"`
			} `json:"comments"`
		} `json:"node"`
	} `json:"data"`
}

func ParseReviewThreadCommentsPage(stdout []byte) (reviewThreadCommentsPage, error) {
	var response reviewThreadCommentsResponse
	if err := json.Unmarshal(stdout, &response); err != nil {
		return reviewThreadCommentsPage{}, fmt.Errorf("decode review thread comments page: %w", err)
	}
	if response.Data == nil || response.Data.Node == nil {
		return reviewThreadCommentsPage{}, errors.New("missing review thread comments payload")
	}

	comments := response.Data.Node.Comments
	return reviewThreadCommentsPage{
		Comments:    normalizePullRequestComments(comments.Nodes),
		HasNextPage: comments.PageInfo.HasNextPage,
		EndCursor:   strings.TrimSpace(comments.PageInfo.EndCursor),
	}, nil
}

func ParseLineNumber(raw string) (int, error) {
	actual, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, err
	}
	return actual, nil
}
