package tui

import (
	persistcache "github.com/l-lin/lazygh/internal/cache"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type notificationDoneStoreAdapter struct {
	store *persistcache.NotificationDoneStore
}

func (adapter notificationDoneStoreAdapter) FilterNotifications(notifications []githubdomain.Notification) []githubdomain.Notification {
	if adapter.store == nil {
		return append([]githubdomain.Notification(nil), notifications...)
	}
	return adapter.store.FilterNotifications(notifications)
}

func (adapter notificationDoneStoreAdapter) HideNotifications(notifications []githubdomain.Notification) error {
	if adapter.store == nil {
		return nil
	}
	return adapter.store.HideNotifications(notifications)
}
