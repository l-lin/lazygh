package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgNotificationReadRequested_WhenApplying_ThenItStartsTheOptimisticMutationAndReturnsATypedRequestCommand(t *testing.T) {
	notifications := []githubcli.Notification{
		given_notificationValue(t, given_pullRequestNotificationRow()),
		given_notificationValue(t, given_issueNotificationRow()),
	}
	subject := given_notificationActionProgram(notifications, &fakePullRequestDetailLoader{})
	target, ok := subject.selectedNotificationActionTarget()
	if !ok {
		t.Fatal("expected a selected notification target")
	}

	actual := Update(subject, MsgNotificationReadRequested{Target: target})

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
	actualRows := subject.model.NotificationRows()
	if actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the selected notification row to become read optimistically, actual %+v", actualRows[0].Notification)
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
