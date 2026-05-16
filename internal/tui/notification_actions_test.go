package tui

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	persistcache "github.com/l-lin/lazygh/internal/cache"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestNotificationsView_GivenUnreadNotification_WhenPressingR_ThenItMarksTheNotificationReadAndKeepsTheRowVisible(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'r')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.markNotificationReadIDs, []string{"n-pr"}) {
		t.Fatalf("expected marked notification ids %v, actual %v", []string{"n-pr"}, loader.markNotificationReadIDs)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 2 {
		t.Fatalf("expected notification row count %d, actual %d", 2, len(actualRows))
	}
	if actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the selected notification to stay visible and become read, actual %+v", actualRows[0].Notification)
	}
}

func TestNotificationsView_GivenUnreadNotification_WhenPressingR_ThenItMarksTheRowReadBeforeGitHubConfirmsTheMutation(t *testing.T) {
	notifications := []githubcli.Notification{
		given_unsupportedNotification("n-push-1", true, "Push notification one"),
		given_unsupportedNotification("n-push-2", false, "Push notification two"),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialQueuedRuns := len(asyncRunner.runs)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'r')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != initialQueuedRuns+1 {
		t.Fatalf("expected queued async runs %d, actual %d", initialQueuedRuns+1, len(asyncRunner.runs))
	}
	if len(loader.markNotificationReadIDs) != 0 {
		t.Fatalf("expected no read mutation call before the async work runs, actual %v", loader.markNotificationReadIDs)
	}
	actualRows := subject.model.NotificationRows()
	if actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the selected notification to become read immediately, actual %+v", actualRows[0].Notification)
	}
	then_statusLineContains(t, gui, notificationReadLoadingMessage)
}

func TestNotificationsView_GivenReadNotification_WhenPressingR_ThenItShowsNoopFeedbackWithoutCallingGitHub(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	subject.model.MoveSelectionDown()
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'r')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)

	if len(loader.markNotificationReadIDs) != 0 {
		t.Fatalf("expected no read mutation calls, actual %v", loader.markNotificationReadIDs)
	}
	then_statusLineContains(t, gui, notificationAlreadyReadMessage)
}

func TestNotificationsView_GivenSelectedNotification_WhenPressingD_ThenItMarksTheNotificationDoneAndRemovesTheRow(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'd')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.markNotificationDoneIDs, []string{"n-pr"}) {
		t.Fatalf("expected done notification ids %v, actual %v", []string{"n-pr"}, loader.markNotificationDoneIDs)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 1 {
		t.Fatalf("expected notification row count %d after done, actual %d", 1, len(actualRows))
	}
	if actualRows[0].Notification == nil || actualRows[0].Notification.ID != "n-issue" {
		t.Fatalf("expected the remaining notification id %q, actual %+v", "n-issue", actualRows[0].Notification)
	}
}

func TestNotificationsView_GivenSelectedNotification_WhenPressingD_ThenItRemovesTheRowBeforeGitHubConfirmsTheMutation(t *testing.T) {
	notifications := []githubcli.Notification{
		given_unsupportedNotification("n-push-1", true, "Push notification one"),
		given_unsupportedNotification("n-push-2", false, "Push notification two"),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialQueuedRuns := len(asyncRunner.runs)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'd')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != initialQueuedRuns+1 {
		t.Fatalf("expected queued async runs %d, actual %d", initialQueuedRuns+1, len(asyncRunner.runs))
	}
	if len(loader.markNotificationDoneIDs) != 0 {
		t.Fatalf("expected no done mutation call before the async work runs, actual %v", loader.markNotificationDoneIDs)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 1 {
		t.Fatalf("expected the selected notification to disappear immediately, actual %+v", actualRows)
	}
	if actualRows[0].Notification == nil || actualRows[0].Notification.ID != "n-push-2" {
		t.Fatalf("expected the remaining notification id %q, actual %+v", "n-push-2", actualRows[0].Notification)
	}
	then_statusLineContains(t, gui, notificationDoneLoadingMessage)
}

func TestNotificationsView_GivenNotificationMarkedDone_WhenClearingCacheAndReloading_ThenItKeepsTheThreadHidden(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	cache := &fakePersistentPullRequestCache{}
	subject := given_notificationActionProgram(loader.notifications, loader)
	subject.pullRequestCache = cache
	doneStore, actualErr := persistcache.OpenNotificationDoneStore(filepath.Join(t.TempDir(), "notification-done.json"))
	then_noError(t, actualErr)
	subject.notificationDoneStore = notificationDoneStoreAdapter{store: doneStore}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	handler := given_handlerForBinding(t, subject.keybindingSpecs(), viewNotificationsName, 'd')
	actualErr = handler(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.clearCachedData()
	then_noError(t, actualErr)
	subject.loadNotifications(gui)

	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 1 {
		t.Fatalf("expected notification row count %d after reload, actual %d", 1, len(actualRows))
	}
	if actualRows[0].Notification == nil || actualRows[0].Notification.ID != "n-issue" {
		t.Fatalf("expected the remaining notification id %q, actual %+v", "n-issue", actualRows[0].Notification)
	}
}

func TestActionsPopup_GivenNotificationContext_WhenOpening_ThenItShowsOnlyNotificationActions(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{
		"Mark notification as read",
		"Mark notification as done",
		"Mark all notifications as read",
		"Mark all notifications as done",
		"Open notification in browser",
	} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	for _, hidden := range []string{"Open PR in browser", "Refresh current PR information", "Start review"} {
		if strings.Contains(popupView.Buffer(), hidden) {
			t.Fatalf("expected popup buffer to hide %q in notification context, actual %q", hidden, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenPullRequestNotificationDetailFocus_WhenOpening_ThenItShowsTheUsualPullRequestActions(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{
		"Start review",
		"Start review as story",
		"Yank URL to clipboard",
		"Open PR in browser",
		"Refresh current PR information",
	} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	for _, hidden := range []string{"Mark notification as read", "Mark notification as done", "Mark all notifications as read", "Mark all notifications as done"} {
		if strings.Contains(popupView.Buffer(), hidden) {
			t.Fatalf("expected popup buffer to hide %q in pull request notification detail, actual %q", hidden, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenPullRequestNotificationDetailFocusAndPendingReview_WhenOpening_ThenItAlsoShowsCancelPendingReview(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{
		notifications:        append([]githubcli.Notification(nil), notifications...),
		reviewKeyByPendingID: map[string]string{"PRR_pending": "acme/widgets#42"},
	}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{
		"Start review",
		"Cancel pending review",
		"Start review as story",
		"Yank URL to clipboard",
		"Open PR in browser",
		"Refresh current PR information",
	} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
}

func TestActionsPopup_GivenPullRequestNotificationDetailFocus_WhenExecutingStartReview_ThenItStartsReviewMode(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{
		notifications: append([]githubcli.Notification(nil), notifications...),
		startReviewID: "PRR_pending",
	}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	if !subject.reviewSession.active {
		t.Fatal("expected review mode to become active")
	}
	if subject.reviewSession.pendingReviewID != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", subject.reviewSession.pendingReviewID)
	}
	if subject.reviewSession.summary.Number != 42 {
		t.Fatalf("expected review summary number %d, actual %d", 42, subject.reviewSession.summary.Number)
	}
}

func TestActionsPopup_GivenNotificationReadAction_WhenExecuting_ThenItRefreshesTheNotificationRowState(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark notification as read", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark notification as read"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	actualRows := subject.model.NotificationRows()
	if actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the refreshed notification row to be marked read, actual %+v", actualRows[0].Notification)
	}
}

func TestActionsPopup_GivenBulkNotificationReadAction_WhenExecuting_ThenItMarksAllLoadedRowsRead(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{
		notifications:                    append([]githubcli.Notification(nil), notifications...),
		markAllNotificationsReadAccepted: true,
	}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark all notifications as read", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark all notifications as read"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if loader.markAllNotificationsReadCalls != 1 {
		t.Fatalf("expected bulk read calls %d, actual %d", 1, loader.markAllNotificationsReadCalls)
	}
	actualRows := subject.model.NotificationRows()
	for _, row := range actualRows {
		if row.Notification != nil && row.Notification.Unread {
			t.Fatalf("expected all notification rows to be marked read, actual %+v", actualRows)
		}
	}
	then_statusLineContains(t, gui, notificationMarkedAllReadMessage)
}

func TestActionsPopup_GivenBulkNotificationReadAction_WhenExecuting_ThenItMarksAllLoadedRowsReadBeforeGitHubConfirmsTheMutation(t *testing.T) {
	notifications := []githubcli.Notification{
		given_unsupportedNotification("n-push-1", true, "Push notification one"),
		given_unsupportedNotification("n-push-2", false, "Push notification two"),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialQueuedRuns := len(asyncRunner.runs)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark all notifications as read", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark all notifications as read"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != initialQueuedRuns+1 {
		t.Fatalf("expected queued async runs %d, actual %d", initialQueuedRuns+1, len(asyncRunner.runs))
	}
	if loader.markAllNotificationsReadCalls != 0 {
		t.Fatalf("expected no bulk read mutation call before the async work runs, actual %d", loader.markAllNotificationsReadCalls)
	}
	actualRows := subject.model.NotificationRows()
	for _, row := range actualRows {
		if row.Notification != nil && row.Notification.Unread {
			t.Fatalf("expected all notification rows to become read immediately, actual %+v", actualRows)
		}
	}
	then_statusLineContains(t, gui, notificationAllReadLoadingMessage)
}

func TestActionsPopup_GivenBulkNotificationDoneAction_WhenExecuting_ThenItRemovesAllLoadedNotificationRowsBeforeGitHubConfirmsTheMutation(t *testing.T) {
	notifications := []githubcli.Notification{
		given_unsupportedNotification("n-push-1", true, "Push notification one"),
		given_unsupportedNotification("n-push-2", false, "Push notification two"),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	initialQueuedRuns := len(asyncRunner.runs)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark all notifications as done", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark all notifications as done"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != initialQueuedRuns+1 {
		t.Fatalf("expected queued async runs %d, actual %d", initialQueuedRuns+1, len(asyncRunner.runs))
	}
	if len(loader.markAllNotificationsDoneIDs) != 0 {
		t.Fatalf("expected no bulk done mutation call before the async work runs, actual %v", loader.markAllNotificationsDoneIDs)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 1 || actualRows[0].Item.Title != notificationsEmptyTitle || actualRows[0].Notification != nil {
		t.Fatalf("expected the notifications list to become empty immediately, actual %+v", actualRows)
	}
	then_statusLineContains(t, gui, "Marking 2 notifications as done...")
}

func TestActionsPopup_GivenBulkNotificationDoneAction_WhenExecuting_ThenItRemovesAllLoadedNotificationRows(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("mark all notifications as done", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark all notifications as done"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.markAllNotificationsDoneIDs, [][]string{{"n-pr", "n-issue"}}) {
		t.Fatalf("expected bulk done ids %v, actual %v", [][]string{{"n-pr", "n-issue"}}, loader.markAllNotificationsDoneIDs)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 1 || actualRows[0].Item.Title != notificationsEmptyTitle || actualRows[0].Notification != nil {
		t.Fatalf("expected the notifications list to refresh to the empty state, actual %+v", actualRows)
	}
}

func TestActionsPopup_GivenNotificationOpenBrowserAction_WhenExecuting_ThenItUsesTheConfiguredLinkOpener(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_issueNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	opener := &fakeLinkOpener{}
	subject.linkOpener = opener
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("open notification in browser", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "open notification in browser"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(opener.urls, []string{"https://github.com/acme/opencode/issues/3235"}) {
		t.Fatalf("expected opened notification urls %v, actual %v", []string{"https://github.com/acme/opencode/issues/3235"}, opener.urls)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
}

func TestStatusLineKeyHints_GivenNotificationsFocus_WhenRendering_ThenItShowsReadDoneAndActionHints(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, r: read, d: done, a: action")
}

func TestHelpPopup_GivenNotificationsFocus_WhenTogglingHelp_ThenItShowsTheNotificationReadAndDoneKeys(t *testing.T) {
	notifications := []githubcli.Notification{given_notificationValue(t, given_pullRequestNotificationRow())}
	loader := &fakePullRequestDetailLoader{notifications: append([]githubcli.Notification(nil), notifications...)}
	subject := given_notificationActionProgram(loader.notifications, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Mark notification as read", "r")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Mark notification as done", "d")
}

func given_notificationActionProgram(notifications []githubcli.Notification, loader *fakePullRequestDetailLoader) *Program {
	model := NewModel(DefaultSeedData())
	model.FocusNotificationsView()
	rows := make([]NotificationRow, 0, len(notifications))
	for _, notification := range notifications {
		rows = append(rows, notificationRow(notification))
	}
	model.SetNotificationRows(rows)

	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	return subject
}

func given_notificationValue(t *testing.T, row NotificationRow) githubcli.Notification {
	t.Helper()

	if row.Notification == nil {
		t.Fatal("expected a notification row with a notification value")
	}
	return githubcli.NotificationFromDomain(*row.Notification)
}

func given_unsupportedNotification(threadID string, unread bool, title string) githubcli.Notification {
	return githubcli.Notification{
		ID:        threadID,
		Unread:    unread,
		Reason:    "manual",
		UpdatedAt: "2026-05-08T16:53:11Z",
		Repository: githubcli.Repository{
			NameWithOwner: "acme/widgets",
		},
		Subject: githubcli.NotificationSubject{
			Type:  "Push",
			Title: title,
		},
	}
}
