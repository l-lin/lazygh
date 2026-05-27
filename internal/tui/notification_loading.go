package tui

import (
	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func (program *Program) maybeLoadNotifications(gui *gocui.Gui) {
	program.executeWorkflowPlan(gui, program.notificationLoadPlan())
}

func notificationRows(notifications []githubdomain.Notification) []NotificationRow {
	return notificationsStateRows(notifications, nil)
}
