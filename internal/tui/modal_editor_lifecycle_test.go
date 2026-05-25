package tui

import (
	"errors"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestOpenModalEditorFromActionsPopup_GivenOpenerCreatesAModal_WhenOpening_ThenItClosesThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.model.OpenActionsPopup(1)

	actual := subject.openModalEditorFromActionsPopup(nil, func(_ *gocui.Gui) error {
		subject.overlayState.modalEditor = newLineModalEditorState("Title", "", nil)
		return nil
	})

	if actual.err != nil {
		t.Fatalf("expected no error, actual %v", actual.err)
	}
	if actual.closePopup {
		t.Fatal("expected popup close to route through the reducer instead of the legacy action result")
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after opening the modal editor")
	}
}

func TestOpenModalEditorFromActionsPopup_GivenOpenerLeavesTheModalVisibilityUnchanged_WhenOpening_ThenItReturnsUnavailable(t *testing.T) {
	subject := &Program{}

	actual := subject.openModalEditorFromActionsPopup(nil, func(_ *gocui.Gui) error {
		return nil
	})

	if actual.closePopup {
		t.Fatal("expected the actions popup to stay open when no modal editor was opened")
	}
	if !errors.Is(actual.err, errActionsPopupActionUnavailable) {
		t.Fatalf("expected error %v, actual %v", errActionsPopupActionUnavailable, actual.err)
	}
}

func TestOpenModalEditorFromActionsPopup_GivenOpenerFails_WhenOpening_ThenItReturnsTheFailure(t *testing.T) {
	subject := &Program{}
	expected := errors.New("boom")

	actual := subject.openModalEditorFromActionsPopup(nil, func(_ *gocui.Gui) error {
		return expected
	})

	if actual.closePopup {
		t.Fatal("expected the actions popup to stay open after a failure")
	}
	if !errors.Is(actual.err, expected) {
		t.Fatalf("expected error %v, actual %v", expected, actual.err)
	}
}
