package tui

import (
	"testing"
	"time"
)

func TestPullRequestBuildRunPopupState_GivenViewAndSearchTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	document := newDetailDocument("Alpha\nTarget\nOmega", 40)
	subject := pullRequestBuildRunPopupState{searchQuery: "Target", viewState: detailViewState{
		cursor:                 detailPosition{line: 1, column: 0},
		currentSearchMatch:     -1,
		pendingKeySequence:     keySequenceState{pendingTarget: keySequenceTarget{viewName: viewPullRequestBuildInfoName}},
		pendingCharacterMotion: detailPendingCharacterMotion{active: true, direction: detailCharacterMotionDirectionForward, mode: detailCharacterMotionMatch},
		pendingYank:            true,
		mode:                   detailVisualMode,
		visualAnchor:           detailPosition{line: 0, column: 0},
	}}

	synced := subject.withRenderStateSynced(document, 2)
	cleared := subject.withPendingPrefixCleared()
	exited := subject.withVisualModeExited()
	opened := subject.withSearchOpened()
	submitted := opened.withSearchSubmitted("Next target")
	cancelled := opened.withSearchCancelled()
	replaced := subject.withViewState(detailViewState{cursor: detailPosition{line: 2, column: 0}, originRow: 1, currentSearchMatch: 7})

	if actual := synced.viewState.currentSearchMatch; actual != 0 {
		t.Fatalf("expected synced current search match %d, actual %d", 0, actual)
	}
	if actual := synced.viewState.searchCacheQuery; actual != "Target" {
		t.Fatalf("expected synced search cache query %q, actual %q", "Target", actual)
	}
	if cleared.viewState.pendingKeySequence != (keySequenceState{}) {
		t.Fatalf("expected cleared pending key sequence to reset, actual %+v", cleared.viewState.pendingKeySequence)
	}
	if cleared.viewState.pendingCharacterMotion != (detailPendingCharacterMotion{}) {
		t.Fatalf("expected cleared pending character motion to reset, actual %+v", cleared.viewState.pendingCharacterMotion)
	}
	if cleared.viewState.pendingYank {
		t.Fatal("expected cleared pending yank flag to reset")
	}
	if actual := exited.viewState.mode; actual != detailNormalMode {
		t.Fatalf("expected exited mode %v, actual %v", detailNormalMode, actual)
	}
	if !opened.searchActive {
		t.Fatal("expected opened popup search to become active")
	}
	if opened.viewState.pendingYank {
		t.Fatal("expected opening popup search to clear the pending prefix")
	}
	if submitted.searchActive {
		t.Fatal("expected submitted popup search to become inactive")
	}
	if actual := submitted.searchQuery; actual != "Next target" {
		t.Fatalf("expected submitted popup search query %q, actual %q", "Next target", actual)
	}
	if cancelled.searchActive {
		t.Fatal("expected cancelled popup search to become inactive")
	}
	if actual := replaced.viewState.cursor.line; actual != 2 {
		t.Fatalf("expected replaced cursor line %d, actual %d", 2, actual)
	}
	if actual := replaced.viewState.originRow; actual != 1 {
		t.Fatalf("expected replaced origin row %d, actual %d", 1, actual)
	}
	if actual := replaced.viewState.currentSearchMatch; actual != 7 {
		t.Fatalf("expected replaced current search match %d, actual %d", 7, actual)
	}
	if actual := subject.viewState.currentSearchMatch; actual != -1 {
		t.Fatalf("expected original current search match %d, actual %d", -1, actual)
	}
	if actual := subject.viewState.mode; actual != detailVisualMode {
		t.Fatalf("expected original mode %v, actual %v", detailVisualMode, actual)
	}
	if !subject.viewState.pendingYank {
		t.Fatal("expected the original pending yank flag to stay true")
	}
	if subject.searchActive {
		t.Fatal("expected the original popup search flag to stay inactive")
	}
}

func TestPullRequestBuildRunPopupState_GivenVisualSelection_WhenPreparingClipboard_ThenItReturnsUpdatedCopyAndSelectedTextWithoutMutatingTheOriginal(t *testing.T) {
	subject := newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"})
	document := newDetailDocumentWithWrap(subject.body, 40, false)
	subject.viewState.cursor = detailPosition{line: 0, column: 0}
	subject.viewState.enterVisualMode()
	subject.viewState.cursor = detailPosition{line: 0, column: 4}

	actual := subject.preparedClipboard(document, 3)

	if !actual.hasVisualSelection {
		t.Fatal("expected clipboard preparation to keep the visual selection result")
	}
	if actual.text != "Alpha" {
		t.Fatalf("expected clipboard text %q, actual %q", "Alpha", actual.text)
	}
	if actual.state.viewState.mode != detailNormalMode {
		t.Fatalf("expected prepared popup mode %v, actual %v", detailNormalMode, actual.state.viewState.mode)
	}
	if subject.viewState.mode != detailVisualMode {
		t.Fatalf("expected original popup mode %v, actual %v", detailVisualMode, subject.viewState.mode)
	}
}

func TestPullRequestBuildRunPopupState_GivenPendingKeySequenceTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	originalTarget := keySequenceTargetFor(viewPullRequestBuildInfoName, keymapScopeSelection, "move_selection_to_top")
	replacementTarget := keySequenceTargetFor(viewPullRequestBuildInfoName, keymapScopeSelection, "recenter_selection")
	subject := pullRequestBuildRunPopupState{viewState: detailViewState{pendingKeySequence: keySequenceState{pendingTarget: originalTarget}}}

	armed := subject.withPendingKeySequenceArmed(replacementTarget)
	cleared := subject.withPendingKeySequenceCleared()

	if actual := armed.pendingKeySequenceTarget(); actual != replacementTarget {
		t.Fatalf("expected armed pending key sequence target %+v, actual %+v", replacementTarget, actual)
	}
	if actual := cleared.pendingKeySequenceTarget(); actual != (keySequenceTarget{}) {
		t.Fatalf("expected cleared pending key sequence target %+v, actual %+v", keySequenceTarget{}, actual)
	}
	if actual := subject.pendingKeySequenceTarget(); actual != originalTarget {
		t.Fatalf("expected the original pending key sequence target %+v, actual %+v", originalTarget, actual)
	}
}

func TestPullRequestBuildRunPopupState_GivenYankHighlightLifecycleTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	now := time.Date(2026, time.May, 29, 14, 0, 0, 0, time.UTC)
	selection := detailSelectionRange{start: detailPosition{line: 0, column: 0}, end: detailPosition{line: 0, column: 4}}
	subject := pullRequestBuildRunPopupState{viewState: detailViewState{yankHighlight: detailYankHighlightState{active: true, start: detailPosition{line: 1, column: 0}, end: detailPosition{line: 1, column: 2}, expiresAt: now.Add(-time.Second)}}}

	activated := subject.withYankHighlightActivated(selection, now.Add(time.Second))
	cleared, changed := subject.withExpiredYankHighlightCleared(now)

	if !activated.hasYankHighlight() {
		t.Fatal("expected activated popup state to report an active yank highlight")
	}
	if actual := activated.viewState.yankHighlight.start; actual != selection.start {
		t.Fatalf("expected activated popup highlight start %+v, actual %+v", selection.start, actual)
	}
	if actual := activated.viewState.yankHighlight.end; actual != selection.end {
		t.Fatalf("expected activated popup highlight end %+v, actual %+v", selection.end, actual)
	}
	if actual := activated.viewState.yankHighlight.expiresAt; !actual.Equal(now.Add(time.Second)) {
		t.Fatalf("expected activated popup expiry %v, actual %v", now.Add(time.Second), actual)
	}
	if !changed {
		t.Fatal("expected expired popup yank highlight cleanup to report a change")
	}
	if cleared.hasYankHighlight() {
		t.Fatalf("expected cleared popup state to drop the yank highlight, actual %+v", cleared.viewState.yankHighlight)
	}
	if !subject.hasYankHighlight() {
		t.Fatal("expected the original popup state to keep its yank highlight")
	}
	if actual := subject.viewState.yankHighlight.start; actual != (detailPosition{line: 1, column: 0}) {
		t.Fatalf("expected the original popup highlight start %+v, actual %+v", detailPosition{line: 1, column: 0}, actual)
	}
}
