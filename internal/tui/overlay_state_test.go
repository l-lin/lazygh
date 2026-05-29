package tui

import (
	"testing"
	"time"
)

func TestOverlayStateModel_GivenHelpErrorPopupAndModalTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	now := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.UTC)
	subject := overlayStateModel{
		errorMessages:       []string{"old"},
		transientErrorPopup: transientErrorPopupState{message: "stale", generation: 7},
		modalEditor:         newLineModalEditorState("Prompt", "draft"),
	}

	helpShown := subject.withHelpVisible(true)
	reported, popup := subject.withReportedError("boom", now, time.Second)
	popupCleared := reported.withClearedTransientErrorPopup()
	replacedModal := subject.withModalEditor(newModalEditorState("Comment", "Ship it"))
	clearedModal := replacedModal.withClearedModalEditor()

	if !helpShown.helpVisible {
		t.Fatal("expected the updated overlay state to show help")
	}
	if actual := popup.message; actual != "boom" {
		t.Fatalf("expected reported popup message %q, actual %q", "boom", actual)
	}
	if actual := popup.generation; actual != 8 {
		t.Fatalf("expected reported popup generation %d, actual %d", 8, actual)
	}
	if !popup.expiresAt.Equal(now.Add(time.Second)) {
		t.Fatalf("expected popup expiry %v, actual %v", now.Add(time.Second), popup.expiresAt)
	}
	if len(reported.errorMessages) != 2 || reported.errorMessages[1] != "boom" {
		t.Fatalf("expected recorded errors %v, actual %v", []string{"old", "boom"}, reported.errorMessages)
	}
	if actual := reported.transientErrorPopup.message; actual != "boom" {
		t.Fatalf("expected reported transient popup message %q, actual %q", "boom", actual)
	}
	if popupCleared.transientErrorPopup != (transientErrorPopupState{}) {
		t.Fatalf("expected the cleared transient popup state %+v, actual %+v", transientErrorPopupState{}, popupCleared.transientErrorPopup)
	}
	if actual := replacedModal.modalEditor.title; actual != "Comment" {
		t.Fatalf("expected replaced modal title %q, actual %q", "Comment", actual)
	}
	if actual := replacedModal.modalEditor.Text(); actual != "Ship it" {
		t.Fatalf("expected replaced modal text %q, actual %q", "Ship it", actual)
	}
	if clearedModal.modalEditor.visible() {
		t.Fatal("expected the cleared modal editor to be hidden")
	}
	if subject.helpVisible {
		t.Fatal("expected the original help visibility to stay false")
	}
	if len(subject.errorMessages) != 1 || subject.errorMessages[0] != "old" {
		t.Fatalf("expected the original recorded errors %v, actual %v", []string{"old"}, subject.errorMessages)
	}
	if actual := subject.transientErrorPopup.message; actual != "stale" {
		t.Fatalf("expected the original transient popup message %q, actual %q", "stale", actual)
	}
	if actual := subject.modalEditor.Text(); actual != "draft" {
		t.Fatalf("expected the original modal text %q, actual %q", "draft", actual)
	}
}

func TestUpdate_GivenMsgToggleHelpAndMsgCloseHelp_WhenApplying_ThenItUpdatesHelpVisibilityThroughOverlayStateTransitions(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgToggleHelp{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if !subject.overlayState.helpVisible {
		t.Fatal("expected help to become visible after toggling it")
	}

	actual = Update(subject, MsgCloseHelp{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if subject.overlayState.helpVisible {
		t.Fatal("expected help to become hidden after closing it")
	}
}

func TestUpdate_GivenMsgModalEditorClosed_WhenApplying_ThenItClearsTheModalEditorThroughOverlayStateTransitions(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.overlayState.modalEditor = newLineModalEditorState("Prompt", "draft")

	actual := Update(subject, MsgModalEditorClosed{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to be closed")
	}
}

func TestProgram_GivenExpiredTransientErrorPopup_WhenClearing_ThenItClearsThePopupThroughOverlayStateTransitions(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	now := time.Date(2026, time.May, 29, 13, 0, 0, 0, time.UTC)
	subject.overlayState.transientErrorPopup = transientErrorPopupState{message: "boom", generation: 7, expiresAt: now.Add(-time.Second)}

	actual := subject.clearExpiredTransientErrorPopup(now)

	if !actual {
		t.Fatal("expected the expired popup to report a cleared state")
	}
	if subject.transientErrorPopupVisible() {
		t.Fatalf("expected the transient popup to be cleared, actual %+v", subject.overlayState.transientErrorPopup)
	}
}
