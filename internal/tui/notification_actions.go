package tui

import (
	"errors"
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

func (program *Program) restoreNotificationMutationSnapshot(snapshot notificationMutationSnapshot) {
	program.model.SetNotificationRows(snapshot.rows)
	program.model.SelectNotificationIndex(snapshot.selectedIndex)
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
	return actionsPopupAction{
		id:      "mark-notification-read",
		title:   notificationMarkReadActionTitle,
		icon:    actionsPopupMarkNotificationReadIcon,
		execute: program.executeMarkNotificationReadAction,
	}
}

func (program *Program) markNotificationDoneAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "mark-notification-done",
		title:   notificationMarkDoneActionTitle,
		icon:    actionsPopupMarkNotificationDoneIcon,
		execute: program.executeMarkNotificationDoneAction,
	}
}

func (program *Program) markAllNotificationsReadAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "mark-all-notifications-read",
		title:   notificationMarkAllReadActionTitle,
		icon:    actionsPopupMarkAllNotificationsReadIcon,
		execute: program.executeMarkAllNotificationsReadAction,
	}
}

func (program *Program) markAllNotificationsDoneAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "mark-all-notifications-done",
		title:   notificationMarkAllDoneActionTitle,
		icon:    actionsPopupMarkAllNotificationsDoneIcon,
		execute: program.executeMarkAllNotificationsDoneAction,
	}
}

func (program *Program) openNotificationInBrowserAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "open-notification-in-browser",
		title:   notificationOpenBrowserActionTitle,
		icon:    actionsPopupOpenNotificationBrowserIcon,
		execute: program.executeOpenNotificationInBrowserAction,
	}
}

func (program *Program) executeMarkNotificationReadAction(gui *gocui.Gui) actionsPopupActionResult {
	if err := program.markSelectedNotificationRead(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executeMarkNotificationDoneAction(gui *gocui.Gui) actionsPopupActionResult {
	if err := program.markSelectedNotificationDone(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executeMarkAllNotificationsReadAction(gui *gocui.Gui) actionsPopupActionResult {
	if err := program.markAllLoadedNotificationsRead(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executeMarkAllNotificationsDoneAction(gui *gocui.Gui) actionsPopupActionResult {
	if err := program.markAllLoadedNotificationsDone(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) executeOpenNotificationInBrowserAction(gui *gocui.Gui) actionsPopupActionResult {
	if err := program.openSelectedNotificationInBrowser(gui); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) markSelectedNotificationRead(gui *gocui.Gui) error {
	target, ok := program.selectedNotificationActionTarget()
	if !ok || target.threadID == "" {
		return errActionsPopupActionUnavailable
	}
	if !target.notification.Unread {
		program.setFeedback(program.model.Focus(), notificationAlreadyReadMessage)
		return program.refreshViewsIfGUI(gui)
	}

	optimisticNotifications := program.loadedNotifications()
	if !markNotificationReadState(optimisticNotifications, target.threadID, false) {
		return errActionsPopupActionUnavailable
	}

	return program.startNotificationMutation(
		gui,
		notificationReadLoadingMessage,
		notificationMarkedReadMessage,
		notificationRows(optimisticNotifications),
		func() error {
			return normalizedNotificationMutationError(program.notificationMutations.MarkNotificationRead(target.threadID))
		},
	)
}

func (program *Program) markSelectedNotificationDone(gui *gocui.Gui) error {
	target, ok := program.selectedNotificationActionTarget()
	if !ok || target.threadID == "" {
		return errActionsPopupActionUnavailable
	}

	optimisticNotifications, removed := removeNotificationWithThreadID(program.loadedNotifications(), target.threadID)
	if !removed {
		return errActionsPopupActionUnavailable
	}

	return program.startNotificationMutation(
		gui,
		notificationDoneLoadingMessage,
		notificationMarkedDoneMessage,
		notificationRows(optimisticNotifications),
		func() error {
			if err := normalizedNotificationMutationError(program.notificationMutations.MarkNotificationDone(target.threadID)); err != nil {
				return err
			}
			program.hideDoneNotificationsBestEffort([]githubdomain.Notification{target.notification})
			return nil
		},
	)
}

func (program *Program) markAllLoadedNotificationsRead(gui *gocui.Gui) error {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		program.setFeedback(program.model.Focus(), notificationNoNotificationsLoadedMessage)
		return program.refreshViewsIfGUI(gui)
	}

	optimisticNotifications := append([]githubdomain.Notification(nil), loadedNotifications...)
	markAllNotificationsRead(optimisticNotifications)
	return program.startNotificationMutation(
		gui,
		notificationAllReadLoadingMessage,
		notificationMarkedAllReadMessage,
		notificationRows(optimisticNotifications),
		func() error {
			_, err := program.notificationMutations.MarkAllNotificationsRead()
			return normalizedNotificationMutationError(err)
		},
	)
}

func (program *Program) markAllLoadedNotificationsDone(gui *gocui.Gui) error {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		program.setFeedback(program.model.Focus(), notificationNoNotificationsLoadedMessage)
		return program.refreshViewsIfGUI(gui)
	}

	loadingMessage := fmt.Sprintf("Marking %d notifications as done...", len(loadedNotifications))
	return program.startNotificationMutation(
		gui,
		loadingMessage,
		notificationMarkedAllDoneMessage,
		notificationRows(nil),
		func() error {
			_, err := program.notificationMutations.MarkAllNotificationsDone(loadedNotifications)
			if err = normalizedNotificationMutationError(err); err != nil {
				return err
			}
			program.hideDoneNotificationsBestEffort(loadedNotifications)
			return nil
		},
	)
}

func (program *Program) openSelectedNotificationInBrowser(gui *gocui.Gui) error {
	if program.linkOpener == nil {
		return ErrLinkOpenerUnavailable
	}

	browserURL, ok := program.selectedNotificationBrowserURL()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	if err := program.linkOpener.Open(browserURL); err != nil {
		return err
	}
	program.setFeedback(program.model.Focus(), notificationOpenBrowserSuccessMessage)
	return program.refreshViewsIfGUI(gui)
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
	return program.handleNotificationKeyAction(gui, program.markSelectedNotificationRead)
}

func (program *Program) markNotificationDone(gui *gocui.Gui, _ *gocui.View) error {
	return program.handleNotificationKeyAction(gui, program.markSelectedNotificationDone)
}

func (program *Program) handleNotificationKeyAction(gui *gocui.Gui, action func(*gocui.Gui) error) error {
	if err := action(gui); err != nil {
		program.setFeedback(program.model.Focus(), err.Error())
		return program.refreshViewsIfGUI(gui)
	}
	return nil
}

func (program *Program) startNotificationMutation(gui *gocui.Gui, loadingMessage string, successFeedbackMessage string, optimisticRows []NotificationRow, work func() error) error {
	if !program.hasNotificationMutations() {
		return errors.New("github loader is unavailable")
	}

	snapshot := program.captureNotificationMutationSnapshot()
	if actualErr := program.dispatch(gui, MsgNotificationMutationStarted{OptimisticRows: optimisticRows, LoadingMessage: loadingMessage}); actualErr != nil {
		return actualErr
	}
	if gui == nil {
		err := work()
		return program.dispatch(gui, MsgNotificationMutationFinished{Snapshot: snapshot, SuccessFeedbackMessage: successFeedbackMessage, Err: err})
	}

	program.runAsync(func() {
		err := work()
		program.dispatchAsync(gui, MsgNotificationMutationFinished{Snapshot: snapshot, SuccessFeedbackMessage: successFeedbackMessage, Err: err})
	})
	return nil
}

func normalizedNotificationMutationError(err error) error {
	return normalizeGHCommandError(err)
}
