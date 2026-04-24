package tui

import "testing"

func TestVisibleListLinePosition_GivenSelectionNearTheBottomOfTheViewport_WhenApplyingScrolloff_ThenItKeepsFourLinesBelowTheSelection(t *testing.T) {
	actualOriginY, actualCursorY := visibleListLinePosition(6, 0, 10, 40)

	if actualOriginY != 1 {
		t.Fatalf("expected origin y %d, actual %d", 1, actualOriginY)
	}
	if actualCursorY != 5 {
		t.Fatalf("expected cursor y %d, actual %d", 5, actualCursorY)
	}
}

func TestVisibleListLinePosition_GivenSelectionNearTheTopOfTheViewport_WhenApplyingScrolloff_ThenItKeepsFourLinesAboveTheSelection(t *testing.T) {
	actualOriginY, actualCursorY := visibleListLinePosition(13, 10, 10, 40)

	if actualOriginY != 9 {
		t.Fatalf("expected origin y %d, actual %d", 9, actualOriginY)
	}
	if actualCursorY != 4 {
		t.Fatalf("expected cursor y %d, actual %d", 4, actualCursorY)
	}
}

func TestDetailViewStateSync_GivenCursorNearTheBottomOfTheViewport_WhenApplyingScrolloff_ThenItKeepsFourLinesBelowTheCursor(t *testing.T) {
	document := newDetailDocument(given_multilineDetail(40), 80)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 6, column: 0}

	subject.sync(document, 10)

	then_detailOriginRowIs(t, subject, 1)
}

func TestDetailViewStateSync_GivenCursorNearTheTopOfTheViewport_WhenApplyingScrolloff_ThenItKeepsFourLinesAboveTheCursor(t *testing.T) {
	document := newDetailDocument(given_multilineDetail(40), 80)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 13, column: 0}
	subject.originRow = 10

	subject.sync(document, 10)

	then_detailOriginRowIs(t, subject, 9)
}
