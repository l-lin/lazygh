package tui

import "testing"

func TestReviewCommentNavigationReadModel_GivenASyncedSelection_WhenResolvingTheCurrentPosition_ThenItUsesTheSnapshot(t *testing.T) {
	renderedRows := []reviewDiffRenderedRow{
		{Kind: reviewDiffRenderedRowKindDiffLine, Text: "first"},
		{Kind: reviewDiffRenderedRowKindDiffLine, Text: "second"},
	}
	document := newReviewDiffDetailDocumentWithWordWrap(renderedRows, 72, true)
	state := newDetailViewState()
	state.cursor = detailPosition{line: 1, column: 0}
	state.sync(document, 4)
	subject := reviewCommentNavigationReadModel{
		active:              true,
		selectedFileTreeRow: 3,
		selection:           detailCursorSelection{document: document, state: state},
	}

	actualFileTreeRow, actualRenderedLine := subject.currentPosition()

	if actualFileTreeRow != 3 {
		t.Fatalf("expected file tree row %d, actual %d", 3, actualFileTreeRow)
	}
	if actualRenderedLine != 1 {
		t.Fatalf("expected rendered line %d, actual %d", 1, actualRenderedLine)
	}
}

func TestReviewCommentNavigationReadModel_GivenVisibleCommentLocations_WhenSelectingTargets_ThenItUsesTheSnapshotFiles(t *testing.T) {
	threadOne := reviewDiffThread{ID: "thread-1"}
	threadTwo := reviewDiffThread{ID: "thread-2"}
	document := newReviewDiffDetailDocumentWithWordWrap([]reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindInlineCommentHeader, Thread: &threadOne}}, 72, true)
	state := newDetailViewState()
	state.sync(document, 4)
	subject := reviewCommentNavigationReadModel{
		active:              true,
		selectedFileTreeRow: 1,
		selection:           detailCursorSelection{document: document, state: state},
		files: []reviewCommentNavigationFile{
			{fileTreeRow: 1, renderedRows: []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindInlineCommentHeader, Thread: &threadOne}}},
			{fileTreeRow: 4, renderedRows: []reviewDiffRenderedRow{{Kind: reviewDiffRenderedRowKindInlineCommentHeader, Thread: &threadTwo}}},
		},
	}

	actualForward, forwardOK := subject.target(reviewNavigationForward)
	if !forwardOK {
		t.Fatal("expected a next review comment target")
	}
	if actualForward.fileTreeRow != 4 || actualForward.renderedLine != 0 {
		t.Fatalf("expected next target {fileTreeRow:%d renderedLine:%d}, actual %+v", 4, 0, actualForward)
	}

	subject.selectedFileTreeRow = 4
	actualBackward, backwardOK := subject.target(reviewNavigationBackward)
	if !backwardOK {
		t.Fatal("expected a previous review comment target")
	}
	if actualBackward.fileTreeRow != 1 || actualBackward.renderedLine != 0 {
		t.Fatalf("expected previous target {fileTreeRow:%d renderedLine:%d}, actual %+v", 1, 0, actualBackward)
	}
}
