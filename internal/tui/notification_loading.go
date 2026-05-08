package tui

import (
	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func (program *Program) maybeLoadNotifications(gui *gocui.Gui) {
	if gui == nil || program.githubLoader == nil || program.reviewSession.active || program.notificationsLoadStarted || !program.notificationsPendingLoad() {
		return
	}

	program.notificationsLoadStarted = true
	program.notificationsLoading = true
	program.asyncRunner.Go(func() {
		program.loadNotifications(gui)
	})
}

func (program *Program) notificationsPendingLoad() bool {
	rows := program.model.NotificationRows()
	if len(rows) == 0 {
		return true
	}
	return len(rows) == 1 && program.isNotificationLoadingItem(rows[0].Item)
}

func (program *Program) loadNotifications(gui *gocui.Gui) {
	notifications, err := program.githubLoader.ListNotifications()

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.notificationsLoading = false
		if err == nil {
			program.model.SetNotificationRows(notificationRows(notifications))
			return program.refreshViews(gui)
		}

		program.model.SetNotificationRows(notificationsStateRows(nil, err))
		return program.refreshViews(gui)
	})
}

func notificationRows(notifications []githubcli.Notification) []NotificationRow {
	return notificationsStateRows(notifications, nil)
}
