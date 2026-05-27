package tui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	notificationMarkReadActionTitle    = "Mark notification as read"
	notificationMarkDoneActionTitle    = "Mark notification as done"
	notificationMarkAllReadActionTitle = "Mark all notifications as read"
	notificationMarkAllDoneActionTitle = "Mark all notifications as done"
	notificationOpenBrowserActionTitle = "Open notification in browser"

	notificationAlreadyReadMessage           = iconUnavailable + " Notification already read"
	notificationMarkedReadMessage            = iconStatusSuccess + " Notification marked as read"
	notificationMarkedDoneMessage            = iconStatusSuccess + " Notification marked as done"
	notificationMarkedAllReadMessage         = iconStatusSuccess + " All notifications marked as read"
	notificationMarkedAllDoneMessage         = iconStatusSuccess + " All notifications marked as done"
	notificationOpenBrowserSuccessMessage    = iconLink + " Notification opened"
	notificationNoNotificationsLoadedMessage = iconUnavailable + " No notifications loaded"

	notificationReadLoadingMessage    = "Marking notification as read..."
	notificationDoneLoadingMessage    = "Marking notification as done..."
	notificationAllReadLoadingMessage = "Marking all notifications as read..."
)

type notificationActionTarget struct {
	notification githubdomain.Notification
	threadID     string
}

type notificationMutationSnapshot struct {
	rows          []NotificationRow
	selectedIndex int
}

func (program *Program) isNotificationContext() bool {
	return program.actionContext().IsNotificationContext()
}

func (program *Program) currentNotificationActionsPopupActions() []actionsPopupAction {
	target, ok := program.selectedNotificationActionTarget()
	if !ok {
		return nil
	}

	actions := actionsPopupGrouped(actionsPopupGroupNotifications,
		program.markNotificationReadAction(),
		program.markNotificationDoneAction(),
		program.markAllNotificationsReadAction(),
		program.markAllNotificationsDoneAction(),
		program.openNotificationInBrowserAction(),
	)
	if target.threadID == "" {
		return nil
	}
	return actions
}

func (program *Program) selectedNotificationActionTarget() (notificationActionTarget, bool) {
	if !program.isNotificationContext() {
		return notificationActionTarget{}, false
	}

	notification, ok := program.model.SelectedNotification()
	if !ok {
		return notificationActionTarget{}, false
	}
	return notificationActionTarget{notification: notification, threadID: strings.TrimSpace(notification.ID)}, true
}

func (program *Program) loadedNotifications() []githubdomain.Notification {
	return notificationValues(program.model.NotificationRows())
}

func notificationValues(rows []NotificationRow) []githubdomain.Notification {
	notifications := make([]githubdomain.Notification, 0, len(rows))
	for _, row := range rows {
		if row.Notification == nil {
			continue
		}
		notifications = append(notifications, *row.Notification)
	}
	return notifications
}

func (program *Program) captureNotificationMutationSnapshot() notificationMutationSnapshot {
	return notificationMutationSnapshot{
		rows:          program.model.NotificationRows(),
		selectedIndex: program.model.SelectedNotificationIndex(),
	}
}

func markNotificationReadState(notifications []githubdomain.Notification, threadID string, unread bool) bool {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return false
	}
	for index := range notifications {
		if strings.TrimSpace(notifications[index].ID) != trimmedThreadID {
			continue
		}
		notifications[index].Unread = unread
		return true
	}
	return false
}

func removeNotificationWithThreadID(notifications []githubdomain.Notification, threadID string) ([]githubdomain.Notification, bool) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return append([]githubdomain.Notification(nil), notifications...), false
	}

	filteredNotifications := make([]githubdomain.Notification, 0, len(notifications))
	removed := false
	for _, notification := range notifications {
		if strings.TrimSpace(notification.ID) == trimmedThreadID {
			removed = true
			continue
		}
		filteredNotifications = append(filteredNotifications, notification)
	}
	return filteredNotifications, removed
}

func markAllNotificationsRead(notifications []githubdomain.Notification) {
	for index := range notifications {
		notifications[index].Unread = false
	}
}

func (program *Program) markNotificationReadAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.selectedNotificationActionTarget(); ok && target.threadID != "" {
		requested = MsgNotificationReadRequested{Target: target}
	}
	return actionsPopupAction{
		id:        "mark-notification-read",
		title:     notificationMarkReadActionTitle,
		icon:      actionsPopupMarkNotificationReadIcon,
		requested: requested,
	}
}

func (program *Program) markNotificationDoneAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	if target, ok := program.selectedNotificationActionTarget(); ok && target.threadID != "" {
		requested = MsgNotificationDoneRequested{Target: target}
	}
	return actionsPopupAction{
		id:        "mark-notification-done",
		title:     notificationMarkDoneActionTitle,
		icon:      actionsPopupMarkNotificationDoneIcon,
		requested: requested,
	}
}

func (program *Program) markAllNotificationsReadAction() actionsPopupAction {
	return actionsPopupAction{
		id:        "mark-all-notifications-read",
		title:     notificationMarkAllReadActionTitle,
		icon:      actionsPopupMarkAllNotificationsReadIcon,
		requested: MsgAllNotificationsReadRequested{},
	}
}

func (program *Program) markAllNotificationsDoneAction() actionsPopupAction {
	return actionsPopupAction{
		id:        "mark-all-notifications-done",
		title:     notificationMarkAllDoneActionTitle,
		icon:      actionsPopupMarkAllNotificationsDoneIcon,
		requested: MsgAllNotificationsDoneRequested{},
	}
}

func (program *Program) openNotificationInBrowserAction() actionsPopupAction {
	requested := actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	switch {
	case program.linkOpener == nil:
		requested = actionsPopupErrorRequested(ErrLinkOpenerUnavailable)
	case !program.isNotificationContext():
	case func() bool {
		_, ok := program.selectedNotificationBrowserURL()
		return ok
	}():
		requested = MsgOpenNotificationInBrowserRequested{}
	}
	return actionsPopupAction{
		id:        "open-notification-in-browser",
		title:     notificationOpenBrowserActionTitle,
		icon:      actionsPopupOpenNotificationBrowserIcon,
		requested: requested,
	}
}

func (program *Program) markSelectedNotificationRead(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgNotificationReadRequested{})
}

func (program *Program) markSelectedNotificationDone(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgNotificationDoneRequested{})
}

func (program *Program) markAllLoadedNotificationsRead(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgAllNotificationsReadRequested{})
}

func (program *Program) markAllLoadedNotificationsDone(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgAllNotificationsDoneRequested{})
}

func (program *Program) openSelectedNotificationInBrowser(gui *gocui.Gui) error {
	return program.dispatch(gui, MsgOpenNotificationInBrowserRequested{})
}

func (program *Program) selectedNotificationBrowserURL() (string, bool) {
	notification, ok := program.model.SelectedNotification()
	if !ok {
		return "", false
	}

	if summary, ok := notification.PullRequestSummary(); ok {
		if result, ok := program.pullRequestDetailForSummary(summary); ok && result.err == nil {
			if actual := strings.TrimSpace(result.detail.URL); actual != "" {
				return actual, true
			}
		}
		actual := strings.TrimSpace(summary.URL)
		return actual, actual != ""
	}
	if repository, number, ok := notification.IssueIdentity(); ok {
		if result, ok := program.issueDetailForNotification(notification); ok && result.err == nil {
			if actual := strings.TrimSpace(result.detail.URL); actual != "" {
				return actual, true
			}
		}
		return fmt.Sprintf("https://github.com/%s/issues/%d", repository, number), true
	}
	if repository, _, ok := notification.ReleaseIdentity(); ok {
		if result, ok := program.releaseDetailForNotification(notification); ok && result.err == nil {
			if actual := strings.TrimSpace(result.detail.URL); actual != "" {
				return actual, true
			}
		}
		tag := strings.TrimSpace(notification.Subject.Title)
		if tag == "" {
			return "", false
		}
		return fmt.Sprintf("https://github.com/%s/releases/tag/%s", repository, url.PathEscape(tag)), true
	}
	return "", false
}

func (program *Program) markNotificationRead(gui *gocui.Gui, _ *gocui.View) error {
	return program.markSelectedNotificationRead(gui)
}

func (program *Program) markNotificationDone(gui *gocui.Gui, _ *gocui.View) error {
	return program.markSelectedNotificationDone(gui)
}

func normalizedNotificationMutationError(err error) error {
	return normalizeGHCommandError(err)
}
