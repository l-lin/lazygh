package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) maybeLoadNotifications(gui *gocui.Gui) {
	program.executeWorkflowCommands(gui, program.notificationStore.planLoad(program, gui))
}

func (program *Program) notificationsPendingLoad() bool {
	rows := program.model.NotificationRows()
	if len(rows) == 0 {
		return true
	}
	return len(rows) == 1 && program.isNotificationLoadingItem(rows[0].Item)
}

func (program *Program) loadNotifications(gui *gocui.Gui) {
	notifications, err := program.notificationQueries.ListNotifications()
	if err == nil {
		notifications = program.filterDoneNotifications(notifications)
		program.cacheNotifications(notifications)
	}

	program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
		program.notificationsLoading = false
		program.notificationsLoadingDetailMessage = ""
		if err == nil {
			program.model.SetNotificationRows(notificationRows(notifications))
			return program.refreshViews(gui)
		}
		if !program.shouldPreserveNotificationRowsOnRefreshError() {
			program.model.SetNotificationRows(notificationsStateRows(nil, err))
		}
		return program.refreshViews(gui)
	})
}

func notificationRows(notifications []githubdomain.Notification) []NotificationRow {
	return notificationsStateRows(notifications, nil)
}
