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
