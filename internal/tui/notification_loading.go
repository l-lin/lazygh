package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) maybeLoadNotifications(gui *gocui.Gui) {
	program.executeCmds(gui, program.notificationStore.planLoad(program, gui))
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
		manualRefresh := program.consumeManualNotificationRefresh()
		if err == nil {
			program.model.SetNotificationRows(notificationRows(notifications))
			if manualRefresh {
				program.completeManualRefreshOperation(gui, nil)
			}
			return program.afterStateChange(gui)
		}
		if manualRefresh {
			program.completeManualRefreshOperation(gui, err)
		}
		if !program.shouldPreserveNotificationRowsOnRefreshError() {
			program.model.SetNotificationRows(notificationsStateRows(nil, err))
		}
		return program.afterStateChange(gui)
	})
}

func notificationRows(notifications []githubdomain.Notification) []NotificationRow {
	return notificationsStateRows(notifications, nil)
}
