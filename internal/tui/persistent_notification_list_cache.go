package tui

import "github.com/l-lin/lazygh/internal/githubcli"

func (program *Program) hydrateNotificationsFromCache() bool {
	if program.pullRequestCache == nil || !program.canHydrateNotificationsFromCache() {
		return false
	}

	notifications, ok, actualErr := program.pullRequestCache.Notifications()
	if actualErr != nil || !ok {
		return false
	}

	convertedNotifications := githubcli.NotificationsFromDomain(notifications)
	program.model.SetNotificationRows(notificationRows(program.filterDoneNotifications(convertedNotifications)))
	return true
}

func (program *Program) canHydrateNotificationsFromCache() bool {
	rows := program.model.NotificationRows()
	if len(rows) == 0 {
		return true
	}

	return len(rows) == 1 && program.isNotificationLoadingItem(rows[0].Item)
}

func (program *Program) cacheNotifications(notifications []githubcli.Notification) {
	if program.pullRequestCache == nil {
		return
	}

	_ = program.pullRequestCache.SaveNotifications(githubcli.ToDomainNotifications(program.filterDoneNotifications(notifications)))
}

func (program *Program) shouldPreserveNotificationRowsOnRefreshError() bool {
	rows := program.model.NotificationRows()
	if len(rows) == 0 {
		return false
	}
	if len(rows) > 1 {
		return true
	}

	item := rows[0].Item
	return !program.isNotificationLoadingItem(item) && !program.isNotificationErrorItem(item)
}

func (program *Program) isNotificationErrorItem(item Item) bool {
	switch item.Title {
	case notificationsUnauthenticatedTitle, notificationsUnavailableTitle, notificationsGenericErrorTitle:
		return true
	default:
		return false
	}
}
