package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

func notificationRows(notifications []githubdomain.Notification) []NotificationRow {
	return notificationsStateRows(notifications, nil)
}
