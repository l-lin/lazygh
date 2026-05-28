package tui

import (
	"strings"
	"testing"
)

func TestActionsPopup_GivenStaleStoredSelection_WhenResolvingTheSelectedRenderedLine_ThenItClampsToTheFirstLiveMatch(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(len(subject.currentActionsPopupActions()))
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, []int{99})

	actual := subject.currentActionsPopupSelectedRenderedLine()

	if actual != 1 {
		t.Fatalf("expected selected rendered line %d, actual %d", 1, actual)
	}
}

func TestUpdate_GivenMoveActionsPopupSelectionWithStaleStoredMatches_WhenApplying_ThenItResyncsBeforeMoving(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(len(subject.currentActionsPopupActions()))
	expected := matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit")
	if len(expected) < 2 {
		t.Fatalf("expected at least two edit actions, actual %v", expected)
	}
	subject.model.UpdateActionsPopupSearch("edit", []int{99})

	Update(subject, MsgMoveActionsPopupSelection{Delta: 1})

	if actual := subject.model.ActionsPopupSelectedActionIndex(); actual != expected[1] {
		t.Fatalf("expected selected action index %d after the resynced move, actual %d", expected[1], actual)
	}
}

func TestUpdate_GivenExecuteSelectedActionsPopupActionRequestedWithStaleStoredMatches_WhenApplying_ThenItResyncsBeforeExecuting(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(len(subject.currentActionsPopupActions()))
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, []int{99})

	actual := Update(subject, MsgExecuteSelectedActionsPopupActionRequested{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to open after resyncing the selected action")
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after executing the resynced action")
	}
	if !strings.Contains(subject.overlayState.modalEditor.title, pullRequestTitleEditorTitle) {
		t.Fatalf("expected modal editor title to contain %q, actual %q", pullRequestTitleEditorTitle, subject.overlayState.modalEditor.title)
	}
}
