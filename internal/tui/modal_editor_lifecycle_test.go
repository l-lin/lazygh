package tui

import "testing"

func TestUpdate_GivenMsgModalEditorOpenedWhileActionsPopupVisible_WhenApplying_ThenItClosesThePopupAndOpensTheEditor(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.model.OpenActionsPopup(1)
	expected := newLineModalEditorState("Title", "")

	Update(subject, MsgModalEditorOpened{State: expected})

	then_modalEditorStateMatches(t, subject.overlayState.modalEditor, expected)
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after opening the modal editor")
	}
}

func TestUpdate_GivenMsgModalEditorOpenedWithoutActionsPopup_WhenApplying_ThenItKeepsTheOpenedEditorState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	expected := newLineModalEditorState("Title", "")

	Update(subject, MsgModalEditorOpened{State: expected})

	then_modalEditorStateMatches(t, subject.overlayState.modalEditor, expected)
}

func then_modalEditorStateMatches(t *testing.T, actual modalEditorState, expected modalEditorState) {
	t.Helper()

	if actual.kind != expected.kind {
		t.Fatalf("expected modal editor kind %v, actual %v", expected.kind, actual.kind)
	}
	if actual.title != expected.title {
		t.Fatalf("expected modal editor title %q, actual %q", expected.title, actual.title)
	}
	if actual.Text() != expected.Text() {
		t.Fatalf("expected modal editor text %q, actual %q", expected.Text(), actual.Text())
	}
	if actual.Height() != expected.Height() {
		t.Fatalf("expected modal editor height %d, actual %d", expected.Height(), actual.Height())
	}
}
