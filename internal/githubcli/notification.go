package githubcli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const notificationsListAPIPath = "/notifications?all=true&per_page=100"

const (
	NotificationSubjectTypePullRequest = "PullRequest"
	NotificationSubjectTypeIssue       = "Issue"
	NotificationSubjectTypeRelease     = "Release"
)

var (
	ErrInvalidNotificationResponse      = fmt.Errorf("invalid notification response")
	ErrInvalidIssueDetailResponse       = fmt.Errorf("invalid issue detail response")
	ErrInvalidReleaseDetailResponse     = fmt.Errorf("invalid release detail response")
	ErrMissingNotificationSubjectTarget = fmt.Errorf("missing notification subject target")
)

type Notification struct {
	ID              string              `json:"id"`
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
	Draft       bool               `json:"draft"`
	PreRelease  bool               `json:"prerelease"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	PublishedAt string             `json:"published_at"`
	Author      *PullRequestAuthor `json:"author"`
}

func (client *Client) ListNotifications() ([]Notification, error) {
	result, err := client.runGH("gh api notifications", "api", notificationsListAPIPath, "--paginate", "--slurp")
	if err != nil {
		return nil, normalizeNotificationEndpointError(err)
	}

	var pagedNotifications [][]Notification
	if err := json.Unmarshal(result.Stdout, &pagedNotifications); err == nil {
		flattenedNotifications := make([]Notification, 0)
		for _, page := range pagedNotifications {
			flattenedNotifications = append(flattenedNotifications, page...)
		}
		return normalizedNotifications(flattenedNotifications), nil
	}

	var notifications []Notification
	if err := json.Unmarshal(result.Stdout, &notifications); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidNotificationResponse, err)
	}
	return normalizedNotifications(notifications), nil
}

func (client *Client) GetIssueDetail(repository string, number int) (IssueDetail, error) {
	trimmedRepository, err := normalizeNotificationSubjectTarget(repository, number)
	if err != nil {
		return IssueDetail{}, err
	}

	result, err := client.runGH("gh api issue detail", "api", fmt.Sprintf("repos/%s/issues/%d", trimmedRepository, number))
	if err != nil {
		return IssueDetail{}, err
	}

	var detail IssueDetail
	if err := json.Unmarshal(result.Stdout, &detail); err != nil {
		return IssueDetail{}, fmt.Errorf("%w: %v", ErrInvalidIssueDetailResponse, err)
	}
	return detail.normalized(), nil
}

func (client *Client) GetReleaseDetail(repository string, id int) (ReleaseDetail, error) {
	trimmedRepository, err := normalizeNotificationSubjectTarget(repository, id)
	if err != nil {
		return ReleaseDetail{}, err
	}

	result, err := client.runGH("gh api release detail", "api", fmt.Sprintf("repos/%s/releases/%d", trimmedRepository, id))
	if err != nil {
		return ReleaseDetail{}, err
	}

	var detail ReleaseDetail
	if err := json.Unmarshal(result.Stdout, &detail); err != nil {
		return ReleaseDetail{}, fmt.Errorf("%w: %v", ErrInvalidReleaseDetailResponse, err)
	}
	return detail.normalized(), nil
}

func normalizedNotifications(notifications []Notification) []Notification {
	normalized := make([]Notification, 0, len(notifications))
	for _, notification := range notifications {
		normalized = append(normalized, notification.normalized())
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

func (notification Notification) PullRequestSummary() (PullRequest, bool) {
	if notification.Subject.Type != NotificationSubjectTypePullRequest {
		return PullRequest{}, false
	}

	repository := strings.TrimSpace(notification.Repository.NameWithOwner)
	number, ok := subjectTrailingID(notification.Subject.URL, "/pulls/")
	if repository == "" || !ok {
		return PullRequest{}, false
	}

	return PullRequest{
		Title:      notification.Subject.Title,
		Number:     number,
		Repository: notification.Repository,
		URL:        pullRequestHTMLURL(repository, number),
	}, true
}

func (notification Notification) IssueIdentity() (string, int, bool) {
	if notification.Subject.Type != NotificationSubjectTypeIssue {
		return "", 0, false
	}

	repository := strings.TrimSpace(notification.Repository.NameWithOwner)
	number, ok := subjectTrailingID(notification.Subject.URL, "/issues/")
	if repository == "" || !ok {
		return "", 0, false
	}
	return repository, number, true
}

func (notification Notification) ReleaseIdentity() (string, int, bool) {
	if notification.Subject.Type != NotificationSubjectTypeRelease {
		return "", 0, false
	}

	repository := strings.TrimSpace(notification.Repository.NameWithOwner)
	id, ok := subjectTrailingID(notification.Subject.URL, "/releases/")
	if repository == "" || !ok {
		return "", 0, false
	}
	return repository, id, true
}

func subjectTrailingID(rawURL string, expectedPathFragment string) (int, bool) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" || !strings.Contains(trimmedURL, expectedPathFragment) {
		return 0, false
	}

	parsedURL, err := url.Parse(trimmedURL)
	if err != nil {
		return 0, false
	}
	pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathSegments) == 0 {
		return 0, false
	}

	actual, err := strconv.Atoi(strings.TrimSpace(pathSegments[len(pathSegments)-1]))
	if err != nil || actual <= 0 {
		return 0, false
	}
	return actual, true
}

func normalizeNotificationSubjectTarget(repository string, id int) (string, error) {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || trimmedRepository == "-" || id <= 0 {
		return "", ErrMissingNotificationSubjectTarget
	}
	return trimmedRepository, nil
}

func pullRequestHTMLURL(repository string, number int) string {
	trimmedRepository := strings.TrimSpace(repository)
	if trimmedRepository == "" || number <= 0 {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/pull/%d", trimmedRepository, number)
}

func (detail IssueDetail) normalized() IssueDetail {
	detail.Title = strings.TrimSpace(detail.Title)
	detail.URL = strings.TrimSpace(detail.URL)
	detail.Body = strings.TrimSpace(detail.Body)
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
	detail.CreatedAt = strings.TrimSpace(detail.CreatedAt)
	detail.UpdatedAt = strings.TrimSpace(detail.UpdatedAt)
	detail.PublishedAt = strings.TrimSpace(detail.PublishedAt)
	if detail.Author != nil {
		normalizedAuthor := detail.Author.normalized()
		detail.Author = &normalizedAuthor
	}
	return detail
}
