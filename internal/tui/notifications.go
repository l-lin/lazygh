package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	notificationsLoadingTitle          = "Loading notifications..."
	notificationsLoadingDetail         = "Running `gh api /notifications?all=true&per_page=100 --paginate --slurp` to load notifications."
	notificationsEmptyTitle            = "No notifications"
	notificationsEmptyDetail           = "GitHub returned no unread or read notifications that are still not done."
	notificationsUnauthenticatedTitle  = "GitHub authentication required"
	notificationsUnauthenticatedDetail = "GitHub CLI is not authenticated.\n\nRun `gh auth login`, then restart `lazygh`."
	notificationsUnavailableTitle      = "`gh` not found"
	notificationsUnavailableDetail     = "Install GitHub CLI and make sure `gh` is in your `PATH`, then restart `lazygh`."
	notificationsGenericErrorTitle     = "Could not load notifications"
)

func notificationsLoadingItem() Item {
	return Item{Title: notificationsLoadingTitle, Detail: notificationsLoadingDetail}
}

func notificationsStateRows(notifications []githubdomain.Notification, err error) []NotificationRow {
	if err != nil {
		return []NotificationRow{{Item: notificationsErrorItem(err)}}
	}
	if len(notifications) == 0 {
		return []NotificationRow{{Item: notificationsEmptyItem()}}
	}

	rows := make([]NotificationRow, 0, len(notifications))
	for _, notification := range notifications {
		rows = append(rows, notificationRow(notification))
	}
	return rows
}

func notificationsEmptyItem() Item {
	return Item{Title: notificationsEmptyTitle, Detail: notificationsEmptyDetail}
}

func notificationsErrorItem(err error) Item {
	switch {
	case err == nil:
		return notificationsEmptyItem()
	case isProviderUnauthenticatedError(err):
		return Item{Title: notificationsUnauthenticatedTitle, Detail: notificationsUnauthenticatedDetail}
	case isProviderUnavailableError(err):
		return Item{Title: notificationsUnavailableTitle, Detail: notificationsUnavailableDetail}
	default:
		message := strings.TrimSpace(err.Error())
		if message == "" {
			message = "Unknown error. GitHub misplaced the inbox again."
		}
		return Item{Title: notificationsGenericErrorTitle, Detail: "Failed to load notifications.\n\n" + message}
	}
}

func notificationRow(notification any) NotificationRow {
	notificationValue, ok := toDomainNotification(notification)
	if !ok {
		return NotificationRow{}
	}
	reference := notificationDisplayReference(notificationValue)
	title := strings.TrimSpace(notificationValue.Subject.Title)
	rowTitle := strings.TrimSpace(strings.Join(filterEmptyStrings([]string{notificationReadStateIcon(notificationValue.Unread), notificationIcon(notificationValue.Subject.Type), reference, title}), " "))
	if rowTitle == "" {
		rowTitle = notificationReadStateIcon(notificationValue.Unread)
	}

	detailLines := []string{
		fmt.Sprintf("Repository: %s", valueOrDash(strings.TrimSpace(notificationValue.Repository.NameWithOwner))),
		fmt.Sprintf("Type: %s", notificationTypeLabel(notificationValue.Subject.Type)),
		fmt.Sprintf("Reason: %s", valueOrDash(strings.TrimSpace(notificationValue.Reason))),
		fmt.Sprintf("Unread: %s", yesNo(notificationValue.Unread)),
		fmt.Sprintf("Updated: %s", valueOrDash(strings.TrimSpace(notificationValue.UpdatedAt))),
	}
	if subjectURL := strings.TrimSpace(notificationValue.Subject.URL); subjectURL != "" {
		detailLines = append(detailLines, fmt.Sprintf("API URL: %s", subjectURL))
	}
	if title != "" {
		detailLines = append(detailLines, "", title)
	}

	notificationCopy := notificationValue
	return NotificationRow{
		Item: Item{
			Title:  rowTitle,
			Detail: strings.Join(detailLines, "\n"),
		},
		Notification: &notificationCopy,
	}
}

func notificationDisplayReference(notification githubdomain.Notification) string {
	if summary, ok := notification.PullRequestSummary(); ok {
		return fmt.Sprintf("%s#%d", strings.TrimSpace(summary.Repository.NameWithOwner), summary.Number)
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		return fmt.Sprintf("%s#%d", repository, number)
	}
	if repository, _, ok := notification.ReleaseIdentity(); ok {
		return repository
	}
	return strings.TrimSpace(notification.Repository.NameWithOwner)
}

func notificationReadStateIcon(unread bool) string {
	if unread {
		return iconNotificationUnread
	}
	return iconNotificationRead
}

func notificationTypeLabel(kind string) string {
	switch strings.TrimSpace(kind) {
	case githubdomain.NotificationSubjectTypePullRequest:
		return "Pull request"
	case githubdomain.NotificationSubjectTypeIssue:
		return "Issue"
	case githubdomain.NotificationSubjectTypeRelease:
		return "Release"
	default:
		trimmedKind := strings.TrimSpace(kind)
		if trimmedKind == "" {
			return "Unknown"
		}
		return "Unsupported (" + trimmedKind + ")"
	}
}

func notificationIcon(kind string) string {
	switch strings.TrimSpace(kind) {
	case githubdomain.NotificationSubjectTypePullRequest:
		return iconNotificationPullRequest
	case githubdomain.NotificationSubjectTypeIssue:
		return iconNotificationIssue
	case githubdomain.NotificationSubjectTypeRelease:
		return iconNotificationRelease
	default:
		return iconWarning
	}
}

func (program *Program) isNotificationLoadingItem(item Item) bool {
	return item.Title == notificationsLoadingTitle && item.Detail == notificationsLoadingDetail
}
