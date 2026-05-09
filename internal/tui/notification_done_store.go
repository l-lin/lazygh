package tui

import "codeberg.org/l-lin/lazygh/internal/githubcli"

type notificationDoneStore interface {
	FilterNotifications([]githubcli.Notification) []githubcli.Notification
	HideNotifications([]githubcli.Notification) error
}

type noopNotificationDoneStore struct{}

func (noopNotificationDoneStore) FilterNotifications(notifications []githubcli.Notification) []githubcli.Notification {
	return append([]githubcli.Notification(nil), notifications...)
}

func (noopNotificationDoneStore) HideNotifications([]githubcli.Notification) error {
	return nil
}

func (program *Program) filterDoneNotifications(notifications []githubcli.Notification) []githubcli.Notification {
	if program == nil || program.notificationDoneStore == nil {
		return append([]githubcli.Notification(nil), notifications...)
	}
	return program.notificationDoneStore.FilterNotifications(notifications)
}

func (program *Program) hideDoneNotificationsBestEffort(notifications []githubcli.Notification) {
	if program == nil || program.notificationDoneStore == nil {
		return
	}
	_ = program.notificationDoneStore.HideNotifications(notifications)
}
