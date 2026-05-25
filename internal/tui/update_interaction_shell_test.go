package tui

import (
	"errors"
	"testing"
)

func TestUpdate_GivenMsgOpenLinkUnderCursorRequested_WhenApplying_ThenItReturnsATypedOpenLinkCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgOpenLinkUnderCursorRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one open-link command, actual %d", len(actual))
	}
	if _, ok := actual[0].(openLinkUnderCursorCmd); !ok {
		t.Fatalf("expected an openLinkUnderCursorCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgCopySelectedDetailTextRequested_WhenApplying_ThenItReturnsATypedClipboardPreparationCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgCopySelectedDetailTextRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one clipboard preparation command, actual %d", len(actual))
	}
	if _, ok := actual[0].(prepareSelectedDetailClipboardWriteCmd); !ok {
		t.Fatalf("expected a prepareSelectedDetailClipboardWriteCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgActionsPopupActionErrorHandledWithPopupError_WhenApplying_ThenItReturnsATypedReportErrorCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(3)
	subject.actionsPopupWidget.errorMessage = "stale"

	actual := Update(subject, MsgActionsPopupActionErrorHandled{Err: newTransientErrorPopupActionError(errors.New("boom"))})

	if len(actual) != 1 {
		t.Fatalf("expected one report-error command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reportErrorCmd); !ok {
		t.Fatalf("expected a reportErrorCmd, actual %T", actual[0])
	}
	if actualMessage := subject.actionsPopupWidget.errorMessage; actualMessage != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actualMessage)
	}
}

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithTransientPopupError_WhenApplying_ThenItReturnsATypedReportErrorCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorStateWithSubmitRequested("Comment", "Ship it", nil)
	subject.overlayState.modalEditor.errorMessage = "stale"

	actual := Update(subject, MsgModalEditorSubmitFinished{Err: newTransientErrorPopupActionError(errors.New("boom"))})

	if len(actual) != 1 {
		t.Fatalf("expected one report-error command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reportErrorCmd); !ok {
		t.Fatalf("expected a reportErrorCmd, actual %T", actual[0])
	}
	if actualMessage := subject.overlayState.modalEditor.errorMessage; actualMessage != "" {
		t.Fatalf("expected modal editor error message %q, actual %q", "", actualMessage)
	}
}
