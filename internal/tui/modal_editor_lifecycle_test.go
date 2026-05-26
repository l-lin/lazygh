package tui

import "testing"

func TestUpdate_GivenMsgModalEditorOpenedWhileActionsPopupVisible_WhenApplying_ThenItClosesThePopupAndOpensTheEditor(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.model.OpenActionsPopup(1)
	expected := newLineModalEditorState("Title", "")

	Update(subject, MsgModalEditorOpened{State: expected})

	if subject.overlayState.modalEditor != expected {
		t.Fatalf("expected modal editor state %p, actual %p", expected, subject.overlayState.modalEditor)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after opening the modal editor")
	}
}

func TestUpdate_GivenMsgModalEditorOpenedWithoutActionsPopup_WhenApplying_ThenItKeepsTheOpenedEditorState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	expected := newLineModalEditorState("Title", "")

	Update(subject, MsgModalEditorOpened{State: expected})

	if subject.overlayState.modalEditor != expected {
		t.Fatalf("expected modal editor state %p, actual %p", expected, subject.overlayState.modalEditor)
	}
}
