package github

import (
	"errors"
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
