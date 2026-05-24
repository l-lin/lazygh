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
	program.dispatchAsync(gui, MsgNotificationsLoaded{Notifications: notifications, Err: err})
}

func notificationRows(notifications []githubdomain.Notification) []NotificationRow {
	return notificationsStateRows(notifications, nil)
}
