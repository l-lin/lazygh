package tui

import "testing"

func TestSessionStore_GivenWorkflowAndConnectedUserTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := sessionStore{connectedUserLogin: "stale-login", connectedUserName: "stale name"}

	planned := subject.withLoadPlanned()
	loaded := planned.withConnectedUser(" octocat ", " The Octocat ")

	if !planned.connectedUserLoadStarted {
		t.Fatal("expected workflow planning to mark the connected-user load as started")
	}
	if actual := planned.connectedUserLogin; actual != "stale-login" {
		t.Fatalf("expected planned login %q, actual %q", "stale-login", actual)
	}
	if actual := planned.connectedUserName; actual != "stale name" {
		t.Fatalf("expected planned name %q, actual %q", "stale name", actual)
	}
	if !loaded.connectedUserLoadStarted {
		t.Fatal("expected the loaded store to preserve the load-started flag")
	}
	if actual := loaded.connectedUserLogin; actual != "octocat" {
		t.Fatalf("expected loaded login %q, actual %q", "octocat", actual)
	}
	if actual := loaded.connectedUserName; actual != "The Octocat" {
		t.Fatalf("expected loaded name %q, actual %q", "The Octocat", actual)
	}
	if subject.connectedUserLoadStarted {
		t.Fatal("expected the original load-started flag to stay false")
	}
	if actual := subject.connectedUserLogin; actual != "stale-login" {
		t.Fatalf("expected the original login %q, actual %q", "stale-login", actual)
	}
	if actual := subject.connectedUserName; actual != "stale name" {
		t.Fatalf("expected the original name %q, actual %q", "stale name", actual)
	}
}
