package tui

import (
	"errors"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestOpenModalEditorFromActionsPopup_GivenOpenerCreatesAModal_WhenOpening_ThenItClosesThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.model.OpenActionsPopup(1)

	actualErr := subject.openModalEditorFromActionsPopup(nil, func(_ *gocui.Gui) error {
		subject.overlayState.modalEditor = newLineModalEditorState("Title", "", nil)
		return nil
	})

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after opening the modal editor")
	}
}

func TestOpenModalEditorFromActionsPopup_GivenOpenerLeavesTheModalVisibilityUnchanged_WhenOpening_ThenItReturnsUnavailable(t *testing.T) {
	subject := &Program{}

	actualErr := subject.openModalEditorFromActionsPopup(nil, func(_ *gocui.Gui) error {
		return nil
	})

	if !errors.Is(actualErr, errActionsPopupActionUnavailable) {
		t.Fatalf("expected error %v, actual %v", errActionsPopupActionUnavailable, actualErr)
	}
}

func TestOpenModalEditorFromActionsPopup_GivenOpenerFails_WhenOpening_ThenItReturnsTheFailure(t *testing.T) {
	subject := &Program{}
	expected := errors.New("boom")

	actualErr := subject.openModalEditorFromActionsPopup(nil, func(_ *gocui.Gui) error {
		return expected
	})

	if !errors.Is(actualErr, expected) {
		t.Fatalf("expected error %v, actual %v", expected, actualErr)
	}
}
