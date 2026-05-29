package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgNotificationReadRequestedWithoutExplicitTarget_WhenApplying_ThenItStartsTheOptimisticMutationAndReturnsATypedRequestCommand(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	subject := given_notificationActionProgram(notifications, &fakePullRequestDetailLoader{})
	subject.feedbackMessage = "stale"
	target, ok := subject.selectedNotificationActionTarget()
	if !ok {
		t.Fatal("expected a selected notification target")
	}

	actual := Update(subject, MsgNotificationReadRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one notification mutation command, actual %d", len(actual))
	}
	command, ok := actual[0].(notificationMutationCmd)
	if !ok {
		t.Fatalf("expected a notificationMutationCmd, actual %T", actual[0])
	}
	request, ok := command.request.(notificationReadMutationRequest)
	if !ok {
		t.Fatalf("expected a notificationReadMutationRequest, actual %T", command.request)
	}
	if actual := request.threadID; actual != target.threadID {
		t.Fatalf("expected request thread id %q, actual %q", target.threadID, actual)
	}
	if actual := command.SuccessFeedbackMessage; actual != notificationMarkedReadMessage {
		t.Fatalf("expected success feedback %q, actual %q", notificationMarkedReadMessage, actual)
	}
	if !subject.notificationsLoading {
		t.Fatal("expected notifications loading to start immediately")
	}
	if actual := subject.notificationsLoadingDetailMessage; actual != notificationReadLoadingMessage {
		t.Fatalf("expected notification loading detail %q, actual %q", notificationReadLoadingMessage, actual)
	}
	if actual := subject.feedbackMessage; actual != "" {
		t.Fatalf("expected feedback message %q, actual %q", "", actual)
	}
	actualRows := subject.model.NotificationRows()
	if actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the selected notification row to become read optimistically, actual %+v", actualRows[0].Notification)
	}
}

func TestUpdate_GivenMsgNotificationDoneRequestedWithoutExplicitTarget_WhenApplying_ThenItStartsTheOptimisticMutationAndReturnsATypedRequestCommand(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	subject := given_notificationActionProgram(notifications, &fakePullRequestDetailLoader{})
	subject.feedbackMessage = "stale"
	target, ok := subject.selectedNotificationActionTarget()
	if !ok {
		t.Fatal("expected a selected notification target")
	}

	actual := Update(subject, MsgNotificationDoneRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one notification mutation command, actual %d", len(actual))
	}
	command, ok := actual[0].(notificationMutationCmd)
	if !ok {
		t.Fatalf("expected a notificationMutationCmd, actual %T", actual[0])
	}
	request, ok := command.request.(notificationDoneMutationRequest)
	if !ok {
		t.Fatalf("expected a notificationDoneMutationRequest, actual %T", command.request)
	}
	if actual := request.threadID; actual != target.threadID {
		t.Fatalf("expected request thread id %q, actual %q", target.threadID, actual)
	}
	if actual := request.notification.ID; actual != target.notification.ID {
		t.Fatalf("expected request notification id %q, actual %q", target.notification.ID, actual)
	}
	if actual := command.SuccessFeedbackMessage; actual != notificationMarkedDoneMessage {
		t.Fatalf("expected success feedback %q, actual %q", notificationMarkedDoneMessage, actual)
	}
	if !subject.notificationsLoading {
		t.Fatal("expected notifications loading to start immediately")
	}
	if actual := subject.notificationsLoadingDetailMessage; actual != notificationDoneLoadingMessage {
		t.Fatalf("expected notification loading detail %q, actual %q", notificationDoneLoadingMessage, actual)
	}
	if actual := subject.feedbackMessage; actual != "" {
		t.Fatalf("expected feedback message %q, actual %q", "", actual)
	}
	actualRows := subject.model.NotificationRows()
	if len(actualRows) != 1 {
		t.Fatalf("expected one remaining notification row, actual %d", len(actualRows))
	}
	if actualRows[0].Notification == nil || actualRows[0].Notification.ID != "n-issue" {
		t.Fatalf("expected the selected notification row to disappear optimistically, actual %+v", actualRows)
	}
}

func TestUpdate_GivenMsgReviewStoryRequested_WhenApplying_ThenItClosesThePopupMarksStoryLoadingAndReturnsATypedPrepareRequest(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	summary, ok := subject.currentPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	subject.model.OpenActionsPopup(1)
	subject.actionsPopupWidget.searchEditor = newLineEditor("story")

	actual := Update(subject, MsgReviewStoryRequested{Summary: summary})

	if len(actual) != 1 {
		t.Fatalf("expected one story review command, actual %d", len(actual))
	}
	command, ok := actual[0].(storyReviewPrepareCmd)
	if !ok {
		t.Fatalf("expected a storyReviewPrepareCmd, actual %T", actual[0])
	}
	request, ok := command.request.(pullRequestStoryReviewPrepareRequest)
	if !ok {
		t.Fatalf("expected a pullRequestStoryReviewPrepareRequest, actual %T", command.request)
	}
	if actual := request.summary.Number; actual != summary.Number {
		t.Fatalf("expected story review summary number %d, actual %d", summary.Number, actual)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close before starting the story review work")
	}
	if !subject.storyReviewLoading {
		t.Fatal("expected story review loading to start immediately")
	}
}
