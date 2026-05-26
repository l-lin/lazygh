package tui

import (
	"strings"
	"testing"
)

func TestPullRequestBuildRunPopupRenderState_GivenPopupView_WhenPreparing_ThenItSyncsPopupViewportAndSearchOutsideRender(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()

	view, actualErr := gui.SetView(viewPullRequestBuildInfoName, 0, 0, 79, 19, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha\nTarget\nOmega"})
	subject.pullRequestBuildRunPopup.searchQuery = "Target"
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: 999, column: 0}

	subject.syncPullRequestBuildRunPopupRenderState(view)

	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual >= 999 {
		t.Fatalf("expected the popup cursor to be clamped during render-state prep, actual line %d", actual)
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.searchCacheQuery; actual != "Target" {
		t.Fatalf("expected popup search cache query %q, actual %q", "Target", actual)
	}
	if actual := len(subject.pullRequestBuildRunPopup.viewState.searchMatches); actual != 1 {
		t.Fatalf("expected one popup search match after render-state prep, actual %d", actual)
	}
}

func TestPullRequestBuildRunPopupLink_GivenPopupCursorNeedsClamping_WhenResolving_ThenItUsesASnapshotWithoutMutatingStoredPopupState(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()

	view, actualErr := gui.SetView(viewPullRequestBuildInfoName, 0, 0, 79, 19, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	expected := "https://github.com/acme/widgets/actions/runs/42"
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: expected})
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: 999, column: 0}

	actual, ok := subject.currentPullRequestBuildRunPopupLink(view)

	if !ok {
		t.Fatal("expected a build-popup link under the clamped cursor")
	}
	if actual != expected {
		t.Fatalf("expected popup link %q, actual %q", expected, actual)
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != 999 {
		t.Fatalf("expected stored popup cursor line to remain %d while resolving links from a snapshot, actual %d", 999, actual)
	}
	if actual := strings.TrimSpace(subject.pullRequestBuildRunPopup.viewState.searchCacheQuery); actual != "" {
		t.Fatalf("expected link lookup to avoid mutating popup search cache, actual %q", actual)
	}
}
