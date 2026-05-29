package tui

import (
	"errors"
	"testing"
)

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithStatusLineError_WhenApplying_ThenItKeepsTheModalOpenAndShowsStatusFeedback(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.overlayState.modalEditor = newModalEditorState("Comment", "Ship it")
	subject.overlayState.modalEditor.errorMessage = "stale"

	actual := Update(subject, MsgModalEditorSubmitFinished{Err: newModalEditorStatusLineError(FocusDetailView, errors.New("boom"))})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to stay open on a status-line error")
	}
	if actual := subject.overlayState.modalEditor.errorMessage; actual != "stale" {
		t.Fatalf("expected modal editor error message %q, actual %q", "stale", actual)
	}
	if actual := subject.feedbackMessage; actual != "boom" {
		t.Fatalf("expected feedback message %q, actual %q", "boom", actual)
	}
}

func TestUpdate_GivenMsgModalEditorExternalEditFinishedWithError_WhenApplying_ThenItKeepsTheTextAndShowsAModalError(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.overlayState.modalEditor = newLineModalEditorState("Prompt", "draft")

	Update(subject, MsgModalEditorExternalEditFinished{Err: errors.New("edit failed")})

	if actual := subject.overlayState.modalEditor.Text(); actual != "draft" {
		t.Fatalf("expected modal editor text %q, actual %q", "draft", actual)
	}
	if actual := subject.overlayState.modalEditor.errorMessage; actual != "edit failed" {
		t.Fatalf("expected modal editor error message %q, actual %q", "edit failed", actual)
	}
}

func TestUpdate_GivenMsgModalEditorExternalEditFinished_WhenSuccessful_ThenItUpdatesTheTextAndClearsTheModalError(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.overlayState.modalEditor = newLineModalEditorState("Prompt", "draft")
	subject.overlayState.modalEditor.errorMessage = "stale"

	Update(subject, MsgModalEditorExternalEditFinished{Text: "Alpha\nBeta\n"})

	if actual := subject.overlayState.modalEditor.Text(); actual != "Alpha Beta" {
		t.Fatalf("expected modal editor text %q, actual %q", "Alpha Beta", actual)
	}
	if actual := subject.overlayState.modalEditor.errorMessage; actual != "" {
		t.Fatalf("expected modal editor error message %q, actual %q", "", actual)
	}
}
