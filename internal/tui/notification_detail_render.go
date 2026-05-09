package tui

import (
	"fmt"
	"strings"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func renderIssueDetail(repository string, detail githubcli.IssueDetail, renderer MarkdownRenderer, width int) string {
	header := strings.Join(filterEmptyStrings([]string{
		stylePullRequestTitleText(iconNotificationIssue + " " + firstNonEmpty(detail.Title, fmt.Sprintf("Issue #%d", detail.Number))),
		stylePullRequestReferenceText(fmt.Sprintf("%s#%d", strings.TrimSpace(repository), detail.Number)),
		renderIssueMetaLine(detail),
		renderPullRequestAssigneesLine(detail.Assignees),
		renderPullRequestLabelsLine(detail.Labels),
	}), "\n")
	body := renderMarkdownWithFallback(prepareMarkdownForImageRendering(detail.Body, detail.BodyHTML), renderer, width, "No description available.")
	return renderPullRequestDetailContent(header, body)
}

func renderReleaseDetail(repository string, detail githubcli.ReleaseDetail, renderer MarkdownRenderer, width int) string {
	header := strings.Join(filterEmptyStrings([]string{
		stylePullRequestTitleText(iconNotificationRelease + " " + firstNonEmpty(detail.Name, detail.TagName, "Release")),
		stylePullRequestReferenceText(strings.TrimSpace(repository)),
		renderReleaseMetaLine(detail),
	}), "\n")
	body := renderMarkdownWithFallback(prepareMarkdownForImageRendering(detail.Body, detail.BodyHTML), renderer, width, "No release notes available.")
	return renderPullRequestDetailContent(header, body)
}

func renderNotificationDetailLoading(notification githubcli.Notification, repository string, spinner string) string {
	return renderPullRequestDetailContent(renderNotificationHeader(notification, repository), strings.TrimSpace(spinner))
}

func renderIssueDetailError(notification githubcli.Notification, repository string, err error) string {
	return renderNotificationDetailError(notification, repository, "Could not load issue detail.", err)
}

func renderReleaseDetailError(notification githubcli.Notification, repository string, err error) string {
	return renderNotificationDetailError(notification, repository, "Could not load release detail.", err)
}

func renderNotificationDetailError(notification githubcli.Notification, repository string, title string, err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		message = "Unknown error. GitHub found another way to be unhelpful."
	}
	fallback := strings.TrimSpace(notificationRow(notification).Item.Detail)
	if fallback == "" {
		fallback = "No fallback detail available."
	}
	return renderPullRequestDetailContent(renderNotificationHeader(notification, repository), fmt.Sprintf("%s\n\n%s\n\n%s", strings.TrimSpace(title), message, fallback))
}

func renderUnsupportedNotificationDetail(notification githubcli.Notification) string {
	repository := strings.TrimSpace(notification.Repository.NameWithOwner)
	body := fmt.Sprintf("Unsupported notification type: %s\n\n%s", notificationTypeLabel(notification.Subject.Type), strings.TrimSpace(notificationRow(notification).Item.Detail))
	return renderPullRequestDetailContent(renderNotificationHeader(notification, repository), body)
}

func renderNotificationHeader(notification githubcli.Notification, repository string) string {
	reference := notificationDisplayReference(notification)
	if reference == "" {
		reference = strings.TrimSpace(repository)
	}
	title := strings.TrimSpace(notification.Subject.Title)
	if title == "" {
		title = notificationTypeLabel(notification.Subject.Type)
	}
	return strings.Join(filterEmptyStrings([]string{
		stylePullRequestTitleText(notificationIcon(notification.Subject.Type) + " " + title),
		stylePullRequestReferenceText(reference),
	}), "\n")
}

func renderIssueMetaLine(detail githubcli.IssueDetail) string {
	parts := []string{}
	if state := strings.ToUpper(strings.TrimSpace(detail.State)); state != "" {
		parts = append(parts, state)
	}
	if detail.Comments > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", detail.Comments, pluralize(detail.Comments, "comment", "comments")))
	}
	if detail.Author != nil {
		parts = append(parts, "Opened by "+formatLogin(detail.Author.Login))
	}
	if createdAt := formattedOptionalTimestamp(detail.CreatedAt); createdAt != "" {
		parts = append(parts, "Created "+createdAt)
	}
	if updatedAt := formattedOptionalTimestamp(detail.UpdatedAt); updatedAt != "" {
		parts = append(parts, "Updated "+updatedAt)
	}
	if len(parts) == 0 {
		return ""
	}
	return stylePullRequestMutedText(strings.Join(parts, "  "))
}

func renderReleaseMetaLine(detail githubcli.ReleaseDetail) string {
	parts := []string{}
	if tag := strings.TrimSpace(detail.TagName); tag != "" {
		parts = append(parts, "Tag "+tag)
	}
	if detail.Draft {
		parts = append(parts, "Draft")
	} else if detail.PreRelease {
		parts = append(parts, "Pre-release")
	} else {
		parts = append(parts, "Published")
	}
	if detail.Author != nil {
		parts = append(parts, "By "+formatLogin(detail.Author.Login))
	}
	if publishedAt := formattedOptionalTimestamp(firstNonEmpty(detail.PublishedAt, detail.UpdatedAt, detail.CreatedAt)); publishedAt != "" {
		parts = append(parts, "Updated "+publishedAt)
	}
	if len(parts) == 0 {
		return ""
	}
	return stylePullRequestMutedText(strings.Join(parts, "  "))
}
