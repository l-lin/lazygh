package githubcli

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const notificationsListAPIPath = "/notifications?all=true&per_page=100"

const (
	NotificationSubjectTypePullRequest = githubdomain.NotificationSubjectTypePullRequest
	NotificationSubjectTypeIssue       = githubdomain.NotificationSubjectTypeIssue
	NotificationSubjectTypeRelease     = githubdomain.NotificationSubjectTypeRelease
)

var (
	ErrInvalidNotificationResponse      = fmt.Errorf("invalid notification response")
	ErrInvalidIssueDetailResponse       = fmt.Errorf("invalid issue detail response")
	ErrInvalidReleaseDetailResponse     = fmt.Errorf("invalid release detail response")
	ErrMissingNotificationSubjectTarget = githubdomain.ErrMissingNotificationSubjectTarget
)

type Notification struct {
	ID              string              `json:"id"`
	Done            bool                `json:"done"`
	Unread          bool                `json:"unread"`
	Reason          string              `json:"reason"`
	UpdatedAt       string              `json:"updated_at"`
	LastReadAt      string              `json:"last_read_at"`
	URL             string              `json:"url"`
	SubscriptionURL string              `json:"subscription_url"`
	Repository      Repository          `json:"repository"`
	Subject         NotificationSubject `json:"subject"`
}

type NotificationSubject struct {
	Title            string `json:"title"`
	Type             string `json:"type"`
	URL              string `json:"url"`
	LatestCommentURL string `json:"latest_comment_url"`
}

type IssueDetail struct {
	Title     string              `json:"title"`
	Number    int                 `json:"number"`
	URL       string              `json:"html_url"`
	Body      string              `json:"body"`
	BodyHTML  string              `json:"bodyHTML,omitempty"`
	Author    *PullRequestAuthor  `json:"user"`
	State     string              `json:"state"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
	Labels    []PullRequestLabel  `json:"labels"`
	Assignees []PullRequestAuthor `json:"assignees"`
	Comments  int                 `json:"comments"`
}

type ReleaseDetail struct {
	Name        string             `json:"name"`
	TagName     string             `json:"tag_name"`
	URL         string             `json:"html_url"`
	Body        string             `json:"body"`
	BodyHTML    string             `json:"bodyHTML,omitempty"`
	Draft       bool               `json:"draft"`
	PreRelease  bool               `json:"prerelease"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	PublishedAt string             `json:"published_at"`
	Author      *PullRequestAuthor `json:"author"`
}

func (client *NotificationService) ListNotifications() ([]Notification, error) {
	result, err := client.doREST(RESTRequest{Path: notificationsListAPIPath, Paginate: true, Slurp: true})
	if err != nil {
		return nil, normalizeNotificationEndpointError(err)
	}

	return NotificationAssembler{}.ParseList(result.Stdout)
}

func (client *NotificationService) GetIssueDetail(repository string, number int) (IssueDetail, error) {
	trimmedRepository, err := normalizeNotificationSubjectTarget(repository, number)
	if err != nil {
		return IssueDetail{}, err
	}

	result, err := client.doREST(RESTRequest{Path: fmt.Sprintf("repos/%s/issues/%d", trimmedRepository, number)})
	if err != nil {
		return IssueDetail{}, err
	}
	return NotificationAssembler{}.ParseIssueDetail(result.Stdout)
}

func (client *NotificationService) GetReleaseDetail(repository string, id int) (ReleaseDetail, error) {
	trimmedRepository, err := normalizeNotificationSubjectTarget(repository, id)
	if err != nil {
		return ReleaseDetail{}, err
	}

	result, err := client.doREST(RESTRequest{Path: fmt.Sprintf("repos/%s/releases/%d", trimmedRepository, id)})
	if err != nil {
		return ReleaseDetail{}, err
	}
	return NotificationAssembler{}.ParseReleaseDetail(result.Stdout)
}

func normalizedNotifications(notifications []Notification) []Notification {
	normalized := make([]Notification, 0, len(notifications))
	for _, notification := range notifications {
		normalizedNotification := notification.normalized()
		if normalizedNotification.Done {
			continue
		}
		normalized = append(normalized, normalizedNotification)
	}
	return normalized
}

func (notification Notification) normalized() Notification {
	notification.ID = strings.TrimSpace(notification.ID)
	notification.Reason = strings.TrimSpace(notification.Reason)
	notification.UpdatedAt = strings.TrimSpace(notification.UpdatedAt)
	notification.LastReadAt = strings.TrimSpace(notification.LastReadAt)
	notification.URL = strings.TrimSpace(notification.URL)
	notification.SubscriptionURL = strings.TrimSpace(notification.SubscriptionURL)
	notification.Repository = notification.Repository.normalized()
	notification.Subject = notification.Subject.normalized()
	return notification
}

func (subject NotificationSubject) normalized() NotificationSubject {
	subject.Title = strings.TrimSpace(subject.Title)
	subject.Type = strings.TrimSpace(subject.Type)
	subject.URL = strings.TrimSpace(subject.URL)
	subject.LatestCommentURL = strings.TrimSpace(subject.LatestCommentURL)
	return subject
}

func normalizeNotificationSubjectTarget(repository string, id int) (string, error) {
	return githubdomain.NormalizeNotificationSubjectTarget(repository, id)
}

func (detail IssueDetail) normalized() IssueDetail {
	detail.Title = strings.TrimSpace(detail.Title)
	detail.URL = strings.TrimSpace(detail.URL)
	detail.Body = strings.TrimSpace(detail.Body)
	detail.BodyHTML = strings.TrimSpace(detail.BodyHTML)
	detail.State = strings.TrimSpace(detail.State)
	detail.CreatedAt = strings.TrimSpace(detail.CreatedAt)
	detail.UpdatedAt = strings.TrimSpace(detail.UpdatedAt)
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
	return detail
}

func (detail ReleaseDetail) normalized() ReleaseDetail {
	detail.Name = strings.TrimSpace(detail.Name)
	detail.TagName = strings.TrimSpace(detail.TagName)
	detail.URL = strings.TrimSpace(detail.URL)
	detail.Body = strings.TrimSpace(detail.Body)
	detail.BodyHTML = strings.TrimSpace(detail.BodyHTML)
	detail.CreatedAt = strings.TrimSpace(detail.CreatedAt)
	detail.UpdatedAt = strings.TrimSpace(detail.UpdatedAt)
	detail.PublishedAt = strings.TrimSpace(detail.PublishedAt)
	if detail.Author != nil {
		normalizedAuthor := detail.Author.normalized()
		detail.Author = &normalizedAuthor
	}
	return detail
}
