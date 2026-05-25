package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgNotificationReadRequested_WhenApplying_ThenItStartsTheOptimisticMutationAndReturnsAnUpdateOwnedCommand(t *testing.T) {
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
	if _, ok := actual[0].(notificationMutationCmd); !ok {
		t.Fatalf("expected a notificationMutationCmd, actual %T", actual[0])
	}
	if !subject.notificationsLoading {
		t.Fatal("expected notifications loading to start immediately")
	}
	actualRows := subject.model.NotificationRows()
	if actualRows[0].Notification == nil || actualRows[0].Notification.Unread {
		t.Fatalf("expected the selected notification row to become read optimistically, actual %+v", actualRows[0].Notification)
	}
}

func TestUpdate_GivenMsgReviewStoryRequested_WhenApplying_ThenItClosesThePopupMarksStoryLoadingAndReturnsACommand(t *testing.T) {
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
	if _, ok := actual[0].(storyReviewPrepareCmd); !ok {
		t.Fatalf("expected a storyReviewPrepareCmd, actual %T", actual[0])
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close before starting the story review work")
	}
	if !subject.storyReviewLoading {
		t.Fatal("expected story review loading to start immediately")
	}
}
