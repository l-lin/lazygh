package tui

import "testing"

func TestDetailViewState_GivenCurrentCursorBeforeAMatch_WhenFollowingASubmittedSearch_ThenItMovesToTheFirstMatchAtOrAfterTheCursor(t *testing.T) {
	document := newDetailDocument("alpha\nbeta alpha\ngamma alpha", 40)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 1, column: 0}

	actualFound := subject.followSubmittedSearch(document, "alpha", 3)
	if !actualFound {
		t.Fatal("expected the submitted search to find a match")
	}

	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 5})
	then_currentDetailSearchMatchIs(t, subject, 1)
}

func TestDetailViewState_GivenCurrentCursorOnAMatch_WhenRepeatingSearchForwardAndBackward_ThenItUsesTheNextOrPreviousMatchFromTheCursor(t *testing.T) {
	document := newDetailDocument("alpha\nbeta alpha\ngamma alpha", 40)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 1, column: 5}

	actualFound := subject.followNextSearchMatch(document, "alpha", 3)
	if !actualFound {
		t.Fatal("expected the forward search repeat to find a match")
	}

	then_detailCursorIs(t, subject, detailPosition{line: 2, column: 6})
	then_currentDetailSearchMatchIs(t, subject, 2)

	actualFound = subject.followPreviousSearchMatch(document, "alpha", 3)
	if !actualFound {
		t.Fatal("expected the backward search repeat to find a match")
	}

	then_detailCursorIs(t, subject, detailPosition{line: 1, column: 5})
	then_currentDetailSearchMatchIs(t, subject, 1)
}

func TestDetailViewState_GivenCurrentCursorPastTheLastOrFirstMatch_WhenFollowingSearch_ThenItWrapsAround(t *testing.T) {
	document := newDetailDocument("alpha\nbeta alpha\ngamma alpha", 40)
	subject := newDetailViewState()
	subject.cursor = detailPosition{line: 2, column: 10}

	actualFound := subject.followSubmittedSearch(document, "alpha", 3)
	if !actualFound {
		t.Fatal("expected the submitted search to wrap to a match")
	}

	then_detailCursorIs(t, subject, detailPosition{line: 0, column: 0})
	then_currentDetailSearchMatchIs(t, subject, 0)

	actualFound = subject.followPreviousSearchMatch(document, "alpha", 3)
	if !actualFound {
		t.Fatal("expected the backward search repeat to wrap to a match")
	}

	then_detailCursorIs(t, subject, detailPosition{line: 2, column: 6})
	then_currentDetailSearchMatchIs(t, subject, 2)
}

func then_currentDetailSearchMatchIs(t *testing.T, subject detailViewState, expected int) {
	t.Helper()

	if subject.currentSearchMatch != expected {
		t.Fatalf("expected current detail search match %d, actual %d", expected, subject.currentSearchMatch)
	}
}
