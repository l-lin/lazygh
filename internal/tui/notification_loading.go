package tui

import (
	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func notificationRows(notifications []githubdomain.Notification) []NotificationRow {
	return notificationRowsWithRepositoryStyle(appconfig.RepositoryStyleOwnerName, notifications)
}
