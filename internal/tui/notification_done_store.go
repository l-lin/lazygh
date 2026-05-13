package tui

import githubdomain "github.com/l-lin/lazygh/internal/github"

type notificationDoneStore interface {
	FilterNotifications([]githubdomain.Notification) []githubdomain.Notification
	HideNotifications([]githubdomain.Notification) error
}

type noopNotificationDoneStore struct{}

func (noopNotificationDoneStore) FilterNotifications(notifications []githubdomain.Notification) []githubdomain.Notification {
	return append([]githubdomain.Notification(nil), notifications...)
}

func (noopNotificationDoneStore) HideNotifications([]githubdomain.Notification) error {
	return nil
}

func (program *Program) filterDoneNotifications(notifications []githubdomain.Notification) []githubdomain.Notification {
	if program == nil || program.notificationDoneStore == nil {
		return append([]githubdomain.Notification(nil), notifications...)
	}
	return program.notificationDoneStore.FilterNotifications(notifications)
}

func (program *Program) hideDoneNotificationsBestEffort(notifications []githubdomain.Notification) {
	if program == nil || program.notificationDoneStore == nil {
		return
	}
	_ = program.notificationDoneStore.HideNotifications(notifications)
}
