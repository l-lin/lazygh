package tui

import (
	"fmt"
	"strings"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/theme"
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
	reference := notificationListReference(notificationValue)
	title := strings.TrimSpace(notificationValue.Subject.Title)
	titleSegments := notificationTitleSegments(notificationValue, reference, title)
	rowTitle := itemTitleFromSegments(titleSegments)

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
			Title:         rowTitle,
			Detail:        strings.Join(detailLines, "\n"),
			TitleSegments: titleSegments,
		},
		Notification: &notificationCopy,
	}
}

func restyledNotificationRows(rows []NotificationRow) []NotificationRow {
	restyledRows := make([]NotificationRow, 0, len(rows))
	for _, row := range rows {
		if row.Notification == nil {
			restyledRows = append(restyledRows, row)
			continue
		}
		restyledRows = append(restyledRows, notificationRow(*row.Notification))
	}
	return restyledRows
}

func notificationTitleSegments(notification githubdomain.Notification, reference string, title string) []ItemTitleSegment {
	segments := make([]ItemTitleSegment, 0, 4)

	readStateIcon := notificationReadStateIcon(notification.Unread)
	if readStateIcon != "" {
		segments = append(segments, ItemTitleSegment{Text: readStateIcon + " "})
	}

	typeIcon := notificationIcon(notification.Subject.Type)
	if typeIcon != "" {
		typeIconText := typeIcon
		if strings.TrimSpace(reference) != "" || strings.TrimSpace(title) != "" {
			typeIconText += " "
		}
		segments = append(segments, ItemTitleSegment{Text: typeIconText})
	}

	trimmedReference := strings.TrimSpace(reference)
	if trimmedReference != "" {
		segments = append(segments, ItemTitleSegment{Text: trimmedReference, ForegroundHex: theme.PullRequestReferenceHex})
	}

	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle != "" {
		titleText := trimmedTitle
		if trimmedReference != "" {
			titleText = " " + titleText
		}
		segments = append(segments, ItemTitleSegment{Text: titleText})
	}

	return segments
}

func itemTitleFromSegments(segments []ItemTitleSegment) string {
	var builder strings.Builder
	for _, segment := range segments {
		builder.WriteString(segment.Text)
	}
	return builder.String()
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
