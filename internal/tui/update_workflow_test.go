package tui

import "testing"

func TestUpdate_GivenMsgNotificationsLoadPlanned_WhenApplying_ThenItStartsTheNotificationStoreLoadingState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.notificationsLoadingDetailMessage = "stale"

	Update(subject, MsgNotificationsLoadPlanned{})

	if !subject.notificationsLoadStarted {
		t.Fatal("expected notification workflow planning to mark the load as started")
	}
	if !subject.notificationsLoading {
		t.Fatal("expected notification workflow planning to mark notifications as loading")
	}
	if actual := subject.notificationsLoadingDetailMessage; actual != notificationsLoadingDetail {
		t.Fatalf("expected notification loading detail %q, actual %q", notificationsLoadingDetail, actual)
	}
}
