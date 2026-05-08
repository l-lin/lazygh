package tui

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

const (
	notificationMarkReadActionTitle    = "Mark notification as read"
	notificationMarkDoneActionTitle    = "Mark notification as done"
	notificationMarkAllReadActionTitle = "Mark all notifications as read"
	notificationMarkAllDoneActionTitle = "Mark all notifications as done"
	notificationOpenBrowserActionTitle = "Open notification in browser"

	notificationAlreadyReadMessage             = iconUnavailable + " Notification already read"
	notificationMarkedReadMessage              = iconStatusSuccess + " Notification marked as read"
	notificationMarkedDoneMessage              = iconStatusSuccess + " Notification marked as done"
	notificationMarkedAllReadMessage           = iconStatusSuccess + " All notifications marked as read"
	notificationMarkedAllDoneMessage           = iconStatusSuccess + " All notifications marked as done"
	notificationBulkReadStillProcessingMessage = iconStatusPending + " GitHub is still marking notifications as read"
	notificationOpenBrowserSuccessMessage      = iconLink + " Notification opened"
	notificationNoNotificationsLoadedMessage   = iconUnavailable + " No notifications loaded"

	notificationReadLoadingMessage    = "Marking notification as read..."
	notificationDoneLoadingMessage    = "Marking notification as done..."
	notificationAllReadLoadingMessage = "Marking all notifications as read..."

	notificationBulkReadPollAttempts = 5
)

var notificationBulkReadPollInterval = 200 * time.Millisecond

type notificationActionTarget struct {
	notification githubcli.Notification
	threadID     string
}

func (program *Program) isNotificationContext() bool {
	if program.reviewSession.active {
		return false
	}

	switch program.model.Focus() {
	case FocusNotificationsView:
		return true
	case FocusDetailView:
		return program.model.currentSideFocus() == FocusNotificationsView
	default:
		return false
	}
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

func (program *Program) loadedNotifications() []githubcli.Notification {
	rows := program.model.NotificationRows()
	notifications := make([]githubcli.Notification, 0, len(rows))
	for _, row := range rows {
		if row.Notification == nil {
			continue
		}
		notifications = append(notifications, *row.Notification)
	}
	return notifications
}

func (program *Program) markNotificationReadAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "mark-notification-read",
		title:    notificationMarkReadActionTitle,
		icon:     actionsPopupMarkNotificationReadIcon,
		keywords: []string{"notification", "read", "thread"},
		execute:  program.executeMarkNotificationReadAction,
	}
}

func (program *Program) markNotificationDoneAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "mark-notification-done",
		title:    notificationMarkDoneActionTitle,
		icon:     actionsPopupMarkNotificationDoneIcon,
		keywords: []string{"notification", "done", "thread", "dismiss"},
		execute:  program.executeMarkNotificationDoneAction,
	}
}

func (program *Program) markAllNotificationsReadAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "mark-all-notifications-read",
		title:    notificationMarkAllReadActionTitle,
		icon:     actionsPopupMarkAllNotificationsReadIcon,
		keywords: []string{"notification", "read", "all", "bulk"},
		execute:  program.executeMarkAllNotificationsReadAction,
	}
}

func (program *Program) markAllNotificationsDoneAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "mark-all-notifications-done",
		title:    notificationMarkAllDoneActionTitle,
		icon:     actionsPopupMarkAllNotificationsDoneIcon,
		keywords: []string{"notification", "done", "all", "bulk", "dismiss"},
		execute:  program.executeMarkAllNotificationsDoneAction,
	}
}

func (program *Program) openNotificationInBrowserAction() actionsPopupAction {
	return actionsPopupAction{
		id:       "open-notification-in-browser",
		title:    notificationOpenBrowserActionTitle,
		icon:     actionsPopupOpenNotificationBrowserIcon,
		keywords: []string{"notification", "browser", "open", "link", "url"},
		execute:  program.executeOpenNotificationInBrowserAction,
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

	return program.startNotificationMutation(gui, notificationReadLoadingMessage, func() ([]githubcli.Notification, string, error) {
		if err := program.githubLoader.MarkNotificationRead(target.threadID); err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		notifications, err := program.githubLoader.ListNotifications()
		if err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		return notifications, notificationMarkedReadMessage, nil
	})
}

func (program *Program) markSelectedNotificationDone(gui *gocui.Gui) error {
	target, ok := program.selectedNotificationActionTarget()
	if !ok || target.threadID == "" {
		return errActionsPopupActionUnavailable
	}

	return program.startNotificationMutation(gui, notificationDoneLoadingMessage, func() ([]githubcli.Notification, string, error) {
		if err := program.githubLoader.MarkNotificationDone(target.threadID); err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		notifications, err := program.githubLoader.ListNotifications()
		if err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		return notifications, notificationMarkedDoneMessage, nil
	})
}

func (program *Program) markAllLoadedNotificationsRead(gui *gocui.Gui) error {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		program.setFeedback(program.model.Focus(), notificationNoNotificationsLoadedMessage)
		return program.refreshViewsIfGUI(gui)
	}

	return program.startNotificationMutation(gui, notificationAllReadLoadingMessage, func() ([]githubcli.Notification, string, error) {
		result, err := program.githubLoader.MarkAllNotificationsRead()
		if err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		notifications, feedbackMessage, err := program.notificationsAfterBulkRead(result.Accepted)
		if err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		return notifications, feedbackMessage, nil
	})
}

func (program *Program) markAllLoadedNotificationsDone(gui *gocui.Gui) error {
	loadedNotifications := program.loadedNotifications()
	if len(loadedNotifications) == 0 {
		program.setFeedback(program.model.Focus(), notificationNoNotificationsLoadedMessage)
		return program.refreshViewsIfGUI(gui)
	}

	loadingMessage := fmt.Sprintf("Marking %d notifications as done...", len(loadedNotifications))
	return program.startNotificationMutation(gui, loadingMessage, func() ([]githubcli.Notification, string, error) {
		if _, err := program.githubLoader.MarkAllNotificationsDone(loadedNotifications); err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		notifications, err := program.githubLoader.ListNotifications()
		if err != nil {
			return nil, "", normalizedNotificationMutationError(err)
		}
		return notifications, notificationMarkedAllDoneMessage, nil
	})
}

func (program *Program) notificationsAfterBulkRead(accepted bool) ([]githubcli.Notification, string, error) {
	notifications, err := program.githubLoader.ListNotifications()
	if err != nil {
		return nil, "", err
	}
	if !accepted {
		return notifications, notificationMarkedAllReadMessage, nil
	}

	for attempt := 0; attempt < notificationBulkReadPollAttempts && notificationsHaveUnreadThreads(notifications); attempt++ {
		if notificationBulkReadPollInterval > 0 {
			time.Sleep(notificationBulkReadPollInterval)
		}
		notifications, err = program.githubLoader.ListNotifications()
		if err != nil {
			return nil, "", err
		}
	}
	if notificationsHaveUnreadThreads(notifications) {
		return notifications, notificationBulkReadStillProcessingMessage, nil
	}
	return notifications, notificationMarkedAllReadMessage, nil
}

func notificationsHaveUnreadThreads(notifications []githubcli.Notification) bool {
	for _, notification := range notifications {
		if notification.Unread {
			return true
		}
	}
	return false
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

func (program *Program) startNotificationMutation(gui *gocui.Gui, loadingMessage string, work func() ([]githubcli.Notification, string, error)) error {
	if program.githubLoader == nil {
		return errors.New("github loader is unavailable")
	}

	if gui == nil {
		notifications, feedbackMessage, err := work()
		return program.finishNotificationMutation(gui, notifications, feedbackMessage, err)
	}

	program.feedbackMessage = ""
	program.notificationsLoading = true
	program.notificationsLoadingDetailMessage = strings.TrimSpace(loadingMessage)
	program.asyncRunner.Go(func() {
		notifications, feedbackMessage, err := work()
		program.uiUpdater.Apply(gui, func(gui *gocui.Gui) error {
			return program.finishNotificationMutation(gui, notifications, feedbackMessage, err)
		})
	})
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) finishNotificationMutation(gui *gocui.Gui, notifications []githubcli.Notification, feedbackMessage string, err error) error {
	program.notificationsLoading = false
	program.notificationsLoadingDetailMessage = ""
	if err != nil {
		program.setFeedback(program.model.Focus(), strings.TrimSpace(err.Error()))
		return program.refreshViewsIfGUI(gui)
	}

	program.cacheNotifications(notifications)
	program.model.SetNotificationRows(notificationRows(notifications))
	program.invalidateNotificationMutationDetailState()
	program.setFeedback(program.model.Focus(), feedbackMessage)
	return program.refreshViewsIfGUI(gui)
}

func (program *Program) invalidateNotificationMutationDetailState() {
	program.issueDetailCache = map[string]issueDetailResult{}
	program.issueDetailLoadInFlight = map[string]bool{}
	program.releaseDetailCache = map[string]releaseDetailResult{}
	program.releaseDetailLoadInFlight = map[string]bool{}
	program.invalidatePullRequestDetailDocumentCache()
}

func normalizedNotificationMutationError(err error) error {
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(err.Error())
	if message == "" {
		return err
	}
	if strings.HasPrefix(message, "run `") {
		if separatorIndex := strings.Index(message, ":"); separatorIndex >= 0 {
			message = strings.TrimSpace(message[separatorIndex+1:])
		}
	}
	if strings.HasPrefix(message, "exit status ") {
		if separatorIndex := strings.Index(message, ":"); separatorIndex >= 0 {
			message = strings.TrimSpace(message[separatorIndex+1:])
		}
	}
	if message == "" {
		return err
	}
	return errors.New(message)
}
