package tui

import (
	"reflect"
	"testing"
)

func TestActionsPopupState_GivenSevenActions_WhenOpening_ThenItStartsVisibleWithAllActionsSelectedFromTheTop(t *testing.T) {
	subject := given_model()

	subject.OpenActionsPopup(7)

	if !subject.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to be visible")
	}
	if subject.ActionsPopupSearchActive() {
		t.Fatal("expected the actions popup search to start unfocused")
	}
	if subject.ActionsPopupSearchQuery() != "" {
		t.Fatalf("expected an empty search query, actual %q", subject.ActionsPopupSearchQuery())
	}

	expected := []int{0, 1, 2, 3, 4, 5, 6}
	actual := subject.ActionsPopupFilteredActionIndexes()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected filtered action indexes %v, actual %v", expected, actual)
	}
	if subject.ActionsPopupSelectedActionIndex() != 0 {
		t.Fatalf("expected selected action index 0, actual %d", subject.ActionsPopupSelectedActionIndex())
	}
	if subject.ActionsPopupSelectedVisibleIndex() != 0 {
		t.Fatalf("expected selected visible action index 0, actual %d", subject.ActionsPopupSelectedVisibleIndex())
	}
}

func TestActionsPopupState_GivenFilteredIndexes_WhenMovingSelection_ThenItMovesWithinTheVisibleMatches(t *testing.T) {
	subject := given_model()
	subject.OpenActionsPopup(7)
	subject.UpdateActionsPopupSearch("review", []int{2, 3, 4})

	subject.MoveActionsPopupSelectionDown()
	if subject.ActionsPopupSelectedActionIndex() != 3 {
		t.Fatalf("expected selected action index 3, actual %d", subject.ActionsPopupSelectedActionIndex())
	}

	subject.MoveActionsPopupSelectionDown()
	if subject.ActionsPopupSelectedActionIndex() != 4 {
		t.Fatalf("expected selected action index 4, actual %d", subject.ActionsPopupSelectedActionIndex())
	}

	subject.MoveActionsPopupSelectionDown()
	if subject.ActionsPopupSelectedActionIndex() != 4 {
		t.Fatalf("expected selected action index to stay at 4, actual %d", subject.ActionsPopupSelectedActionIndex())
	}

	subject.MoveActionsPopupSelectionUp()
	if subject.ActionsPopupSelectedActionIndex() != 3 {
		t.Fatalf("expected selected action index 3 after moving up, actual %d", subject.ActionsPopupSelectedActionIndex())
	}
}

func TestActionsPopupState_GivenSelectedActionMissingFromTheNewFilter_WhenUpdatingSearch_ThenItClampsToTheFirstVisibleAction(t *testing.T) {
	subject := given_model()
	subject.OpenActionsPopup(7)
	subject.UpdateActionsPopupSearch("review", []int{2, 3, 4})
	subject.MoveActionsPopupSelectionDown()
	subject.MoveActionsPopupSelectionDown()

	subject.UpdateActionsPopupSearch("edit", []int{5, 6})

	if subject.ActionsPopupSelectedActionIndex() != 5 {
		t.Fatalf("expected selected action index 5, actual %d", subject.ActionsPopupSelectedActionIndex())
	}
	if subject.ActionsPopupSelectedVisibleIndex() != 0 {
		t.Fatalf("expected selected visible action index 0, actual %d", subject.ActionsPopupSelectedVisibleIndex())
	}
}

func TestActionsPopupState_GivenVisiblePopup_WhenClosing_ThenItClearsThePopupState(t *testing.T) {
	subject := given_model()
	subject.OpenActionsPopup(7)
	subject.FocusActionsPopupSearch()
	subject.UpdateActionsPopupSearch("review", []int{2, 3, 4})

	subject.CloseActionsPopup()

	if subject.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to be hidden")
	}
	if subject.ActionsPopupSearchActive() {
		t.Fatal("expected the actions popup search to be unfocused")
	}
	if subject.ActionsPopupSearchQuery() != "" {
		t.Fatalf("expected an empty search query, actual %q", subject.ActionsPopupSearchQuery())
	}
	if len(subject.ActionsPopupFilteredActionIndexes()) != 0 {
		t.Fatalf("expected no filtered indexes after close, actual %v", subject.ActionsPopupFilteredActionIndexes())
	}
}
