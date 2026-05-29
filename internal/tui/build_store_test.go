package tui

import (
	"strings"
	"testing"
)

func TestBuildStore_GivenLoadAndPopupTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	previousPopup := newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Previous", body: "old"})
	subject := buildStore{
		pullRequestBuildRunLoad:  &pullRequestBuildRunLoadState{command: "gh old"},
		pullRequestBuildRunPopup: newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Current", body: "current", previousPopup: previousPopup}),
	}

	buildRunStarted := subject.withBuildRunLoadStarted(" gh run view ")
	jobLogStarted := subject.withJobLogLoadStarted(" gh run view --json jobs ")
	loadingCleared := buildRunStarted.withLoadCleared()
	popupOpened := loadingCleared.withPopupOpened(pullRequestBuildRunPopupContent{checkTitle: "CI", body: "run body", previousPopup: previousPopup})
	popupClosed := popupOpened.withPopupClosed()

	if buildRunStarted.pullRequestBuildRunPopup != nil {
		t.Fatal("expected build-run load start to clear the popup state")
	}
	if buildRunStarted.pullRequestBuildRunLoad == nil || buildRunStarted.pullRequestBuildRunLoad.command != "gh run view" {
		t.Fatalf("expected build-run load command %q, actual %+v", "gh run view", buildRunStarted.pullRequestBuildRunLoad)
	}
	if jobLogStarted.pullRequestBuildRunPopup == nil {
		t.Fatal("expected job-log load start to preserve the existing popup")
	}
	if jobLogStarted.pullRequestBuildRunLoad == nil || jobLogStarted.pullRequestBuildRunLoad.command != "gh run view --json jobs" {
		t.Fatalf("expected job-log load command %q, actual %+v", "gh run view --json jobs", jobLogStarted.pullRequestBuildRunLoad)
	}
	if loadingCleared.pullRequestBuildRunLoad != nil {
		t.Fatalf("expected loading state to clear, actual %+v", loadingCleared.pullRequestBuildRunLoad)
	}
	if popupOpened.pullRequestBuildRunPopup == nil {
		t.Fatal("expected popup open to create popup state")
	}
	if actual := popupOpened.pullRequestBuildRunPopup.title; !strings.Contains(actual, "CI") {
		t.Fatalf("expected popup title to contain %q, actual %q", "CI", actual)
	}
	if popupClosed.pullRequestBuildRunPopup != previousPopup {
		t.Fatalf("expected popup close to restore the previous popup, actual %+v", popupClosed.pullRequestBuildRunPopup)
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the original popup to stay intact")
	}
	if subject.pullRequestBuildRunLoad == nil || subject.pullRequestBuildRunLoad.command != "gh old" {
		t.Fatalf("expected the original load command %q, actual %+v", "gh old", subject.pullRequestBuildRunLoad)
	}
}
