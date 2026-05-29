package tui

import "testing"

func TestNotificationStore_GivenWorkflowAndMutationLoadingTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := notificationStore{notificationsLoadingDetailMessage: "stale", notificationDoneStore: noopNotificationDoneStore{}}

	planned := subject.withLoadPlanned()
	mutationStarted := subject.withMutationStarted(" marking as read... ")
	finished := mutationStarted.withLoadingFinished()
	reset := planned.withLoadStateReset()

	if !planned.notificationsLoadStarted {
		t.Fatal("expected workflow planning to mark notification loading as started")
	}
	if !planned.notificationsLoading {
		t.Fatal("expected workflow planning to mark notifications as loading")
	}
	if actual := planned.notificationsLoadingDetailMessage; actual != notificationsLoadingDetail {
		t.Fatalf("expected workflow loading detail %q, actual %q", notificationsLoadingDetail, actual)
	}
	if mutationStarted.notificationsLoadStarted {
		t.Fatal("expected mutation preflight to leave the workflow load-start flag unchanged")
	}
	if !mutationStarted.notificationsLoading {
		t.Fatal("expected mutation preflight to mark notifications as loading")
	}
	if actual := mutationStarted.notificationsLoadingDetailMessage; actual != "marking as read..." {
		t.Fatalf("expected mutation loading detail %q, actual %q", "marking as read...", actual)
	}
	if finished.notificationsLoading {
		t.Fatal("expected loading finish to clear the loading flag")
	}
	if actual := finished.notificationsLoadingDetailMessage; actual != "" {
		t.Fatalf("expected finished loading detail %q, actual %q", "", actual)
	}
	if reset.notificationsLoadStarted {
		t.Fatal("expected reset to clear the workflow load-start flag")
	}
	if reset.notificationsLoading {
		t.Fatal("expected reset to clear the loading flag")
	}
	if actual := reset.notificationsLoadingDetailMessage; actual != "" {
		t.Fatalf("expected reset loading detail %q, actual %q", "", actual)
	}
	if subject.notificationsLoadStarted {
		t.Fatal("expected the original workflow load-start flag to stay false")
	}
	if subject.notificationsLoading {
		t.Fatal("expected the original loading flag to stay false")
	}
	if actual := subject.notificationsLoadingDetailMessage; actual != "stale" {
		t.Fatalf("expected the original loading detail %q, actual %q", "stale", actual)
	}
}
