package tui

import "testing"

func TestDetailStateModel_GivenStaleIdentity_WhenSyncing_ThenItReturnsUpdatedCopyWithoutMutatingTheOriginal(t *testing.T) {
	document := newDetailDocument("alpha\nbeta", 40)
	subject := detailStateModel{
		lastIdentity: "stale",
		viewState: detailViewState{
			cursor:          detailPosition{line: 99, column: 7},
			preferredColumn: 7,
		},
	}

	actual := subject.synced("fresh", document, 3, "")

	if actual.lastIdentity != "fresh" {
		t.Fatalf("expected synced identity %q, actual %q", "fresh", actual.lastIdentity)
	}
	if actual.viewState.cursor.line != 0 || actual.viewState.cursor.column != 0 {
		t.Fatalf("expected the synced cursor to reset to the top of the document, actual %+v", actual.viewState.cursor)
	}
	if actual.viewState.preferredColumn != 0 {
		t.Fatalf("expected the synced preferred column 0, actual %d", actual.viewState.preferredColumn)
	}
	if subject.lastIdentity != "stale" {
		t.Fatalf("expected the original identity to stay %q, actual %q", "stale", subject.lastIdentity)
	}
	if subject.viewState.cursor.line != 99 || subject.viewState.preferredColumn != 7 {
		t.Fatalf("expected the original detail view state to stay unchanged, actual %+v", subject.viewState)
	}
}

func TestReviewSessionState_GivenFoldableTree_WhenCollapsingAllRows_ThenItReturnsUpdatedCopyWithoutMutatingTheOriginal(t *testing.T) {
	tree := reviewDiffTree{Rows: []reviewDiffTreeRow{
		{ID: "dir", Foldable: true},
		{ID: "child", Foldable: true},
		{ID: "plain", Foldable: false},
	}}
	subject := reviewSessionState{collapsedTreeRowIDs: map[string]bool{"dir": false}}

	actual, changed := subject.withAllTreeRowsCollapsed(tree, true)
	if !changed {
		t.Fatal("expected the fold reducer to report a change")
	}
	if !actual.collapsedTreeRowIDs["dir"] || !actual.collapsedTreeRowIDs["child"] {
		t.Fatalf("expected all foldable rows to be collapsed, actual %v", actual.collapsedTreeRowIDs)
	}
	if subject.collapsedTreeRowIDs["dir"] {
		t.Fatalf("expected the original fold map to stay unchanged, actual %v", subject.collapsedTreeRowIDs)
	}
	if _, ok := subject.collapsedTreeRowIDs["child"]; ok {
		t.Fatalf("expected the original fold map to stay free of new entries, actual %v", subject.collapsedTreeRowIDs)
	}
}

func TestDetailStateModel_GivenActiveTabTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := detailStateModel{activeTab: CommentsDetailTab, wrapWidth: 72}

	selected := subject.withActiveTab(CommitsDetailTab)
	projected := subject.withProjectedScreenStateApplication(projectedScreenStateApplication{hasDetailTab: true, activeDetailTab: ChangesDetailTab})
	advancedForward := subject.withAdvancedActiveTab(1, len(browserDetailTabs))
	advancedBackward := subject.withAdvancedActiveTab(-1, len(browserDetailTabs))
	unchanged := subject.withProjectedScreenStateApplication(projectedScreenStateApplication{})

	if actual := selected.activeTab; actual != CommitsDetailTab {
		t.Fatalf("expected selected active tab %v, actual %v", CommitsDetailTab, actual)
	}
	if actual := projected.activeTab; actual != ChangesDetailTab {
		t.Fatalf("expected projected active tab %v, actual %v", ChangesDetailTab, actual)
	}
	if actual := advancedForward.activeTab; actual != CommitsDetailTab {
		t.Fatalf("expected forward advanced active tab %v, actual %v", CommitsDetailTab, actual)
	}
	if actual := advancedBackward.activeTab; actual != DescriptionDetailTab {
		t.Fatalf("expected backward advanced active tab %v, actual %v", DescriptionDetailTab, actual)
	}
	if actual := unchanged.activeTab; actual != CommentsDetailTab {
		t.Fatalf("expected unchanged active tab %v, actual %v", CommentsDetailTab, actual)
	}
	if actual := subject.activeTab; actual != CommentsDetailTab {
		t.Fatalf("expected the original active tab %v, actual %v", CommentsDetailTab, actual)
	}
	if actual := subject.wrapWidth; actual != 72 {
		t.Fatalf("expected the original wrap width %d, actual %d", 72, actual)
	}
}

func TestDetailStateModel_GivenPendingPrefixAndVisualModeTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := detailStateModel{viewState: detailViewState{
		cursor:                 detailPosition{line: 3, column: 4},
		visualAnchor:           detailPosition{line: 1, column: 2},
		mode:                   detailVisualMode,
		pendingKeySequence:     keySequenceState{pendingTarget: keySequenceTarget{viewName: viewDetailName}},
		pendingCharacterMotion: detailPendingCharacterMotion{active: true, direction: detailCharacterMotionDirectionForward, mode: detailCharacterMotionMatch},
		pendingYank:            true,
	}}

	cleared := subject.withPendingPrefixCleared()
	exited := subject.withVisualModeExited()

	if cleared.viewState.pendingKeySequence != (keySequenceState{}) {
		t.Fatalf("expected the cleared pending key sequence to reset, actual %+v", cleared.viewState.pendingKeySequence)
	}
	if cleared.viewState.pendingCharacterMotion != (detailPendingCharacterMotion{}) {
		t.Fatalf("expected the cleared pending character motion to reset, actual %+v", cleared.viewState.pendingCharacterMotion)
	}
	if cleared.viewState.pendingYank {
		t.Fatal("expected the cleared pending yank flag to reset")
	}
	if actual := cleared.viewState.mode; actual != detailVisualMode {
		t.Fatalf("expected clear-only mode %v, actual %v", detailVisualMode, actual)
	}
	if actual := exited.viewState.mode; actual != detailNormalMode {
		t.Fatalf("expected exited mode %v, actual %v", detailNormalMode, actual)
	}
	if actual := exited.viewState.visualAnchor; actual != subject.viewState.cursor {
		t.Fatalf("expected exited visual anchor %+v, actual %+v", subject.viewState.cursor, actual)
	}
	if exited.viewState.pendingYank {
		t.Fatal("expected exiting visual mode to clear the pending yank flag")
	}
	if subject.viewState.pendingKeySequence == (keySequenceState{}) {
		t.Fatalf("expected the original pending key sequence to stay armed, actual %+v", subject.viewState.pendingKeySequence)
	}
	if subject.viewState.pendingCharacterMotion == (detailPendingCharacterMotion{}) {
		t.Fatalf("expected the original pending character motion to stay armed, actual %+v", subject.viewState.pendingCharacterMotion)
	}
	if !subject.viewState.pendingYank {
		t.Fatal("expected the original pending yank flag to stay true")
	}
	if actual := subject.viewState.mode; actual != detailVisualMode {
		t.Fatalf("expected the original mode %v, actual %v", detailVisualMode, actual)
	}
	if actual := subject.viewState.visualAnchor; actual != (detailPosition{line: 1, column: 2}) {
		t.Fatalf("expected the original visual anchor %+v, actual %+v", detailPosition{line: 1, column: 2}, actual)
	}
}
