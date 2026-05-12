package tui

import (
	persistcache "github.com/l-lin/lazygh/internal/cache"
	"github.com/l-lin/lazygh/internal/githubcli"
)

type notificationDoneStoreAdapter struct {
	store *persistcache.NotificationDoneStore
}

func (adapter notificationDoneStoreAdapter) FilterNotifications(notifications []githubcli.Notification) []githubcli.Notification {
	if adapter.store == nil {
		return append([]githubcli.Notification(nil), notifications...)
	}
	return githubcli.NotificationsFromDomain(adapter.store.FilterNotifications(githubcli.ToDomainNotifications(notifications)))
}

func (adapter notificationDoneStoreAdapter) HideNotifications(notifications []githubcli.Notification) error {
	if adapter.store == nil {
		return nil
	}
	return adapter.store.HideNotifications(githubcli.ToDomainNotifications(notifications))
}
