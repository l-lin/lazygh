package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestUpdate_GivenMsgSelectedDetailClipboardPrepared_WhenApplying_ThenItExitsVisualModeAndReturnsAClipboardWriteCommand(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha Beta"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	document := newDetailDocument("Alpha Beta", 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 3, subject.model.DetailSearchQuery())
	subject.detailState.viewState.cursor = detailPosition{line: 0, column: 0}
	subject.detailState.viewState.enterVisualMode()
	subject.detailState.viewState.cursor = detailPosition{line: 0, column: 4}

	actual := Update(subject, MsgSelectedDetailClipboardPrepared{Target: FocusDetailView, Document: document, ViewportHeight: 3})

	if len(actual) != 1 {
		t.Fatalf("expected one clipboard command, actual %d", len(actual))
	}
	command, ok := actual[0].(writeClipboardCmd)
	if !ok {
		t.Fatalf("expected a writeClipboardCmd, actual %T", actual[0])
	}
	if actual := command.Text; actual != "Alpha" {
		t.Fatalf("expected clipboard text %q, actual %q", "Alpha", actual)
	}
	if actual := command.SelectionTarget; actual != clipboardWriteSelectionDetail {
		t.Fatalf("expected clipboard selection target %v, actual %v", clipboardWriteSelectionDetail, actual)
	}
	if actual := subject.detailState.viewState.mode; actual != detailNormalMode {
		t.Fatalf("expected detail mode %v after preparation, actual %v", detailNormalMode, actual)
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunPopupClipboardPreparedWithVisualSelection_WhenApplying_ThenItExitsVisualModeAndReturnsAClipboardWriteCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"})
	document := newDetailDocumentWithWrap(subject.pullRequestBuildRunPopup.body, 40, false)
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: 0, column: 0}
	subject.pullRequestBuildRunPopup.viewState.enterVisualMode()
	subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: 0, column: 4}

	actual := Update(subject, MsgPullRequestBuildRunPopupClipboardPrepared{Target: FocusDetailView, Document: document, ViewportHeight: 3})

	if len(actual) != 1 {
		t.Fatalf("expected one clipboard command, actual %d", len(actual))
	}
	command, ok := actual[0].(writeClipboardCmd)
	if !ok {
		t.Fatalf("expected a writeClipboardCmd, actual %T", actual[0])
	}
	if actual := command.Text; actual != "Alpha" {
		t.Fatalf("expected clipboard text %q, actual %q", "Alpha", actual)
	}
	if actual := command.SelectionTarget; actual != clipboardWriteSelectionBuildPopup {
		t.Fatalf("expected clipboard selection target %v, actual %v", clipboardWriteSelectionBuildPopup, actual)
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.mode; actual != detailNormalMode {
		t.Fatalf("expected popup mode %v after preparation, actual %v", detailNormalMode, actual)
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunPopupClipboardPreparedWithoutSelectionOrRunURL_WhenApplying_ThenItClearsThePendingPrefixAndSetsFeedback(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"})
	subject.pullRequestBuildRunPopup.viewState.armPendingYank()

	actual := Update(subject, MsgPullRequestBuildRunPopupClipboardPrepared{Target: FocusDetailView, Document: newDetailDocumentWithWrap(subject.pullRequestBuildRunPopup.body, 40, false), ViewportHeight: 3})

	if len(actual) != 0 {
		t.Fatalf("expected no clipboard command, actual %d", len(actual))
	}
	if subject.pullRequestBuildRunPopup.viewState.hasPendingYank() {
		t.Fatal("expected the popup pending prefix to clear during preparation")
	}
	if actual := subject.feedbackMessage; actual != yankUnavailableMessage {
		t.Fatalf("expected feedback %q, actual %q", yankUnavailableMessage, actual)
	}
}

func TestUpdate_GivenMsgCopyPullRequestBuildRunPopupContentRequested_WhenApplying_ThenItReturnsATypedClipboardPreparationCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha Beta"})

	actual := Update(subject, MsgCopyPullRequestBuildRunPopupContentRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one clipboard preparation command, actual %d", len(actual))
	}
	if _, ok := actual[0].(preparePullRequestBuildRunPopupClipboardWriteCmd); !ok {
		t.Fatalf("expected a preparePullRequestBuildRunPopupClipboardWriteCmd, actual %T", actual[0])
	}
}

func TestPrepareSelectedDetailClipboardWriteCommand_GivenResolvedDocument_WhenExecuting_ThenItDispatchesTheTypedPreparedMessage(t *testing.T) {
	expectedDocument := newDetailDocument("Alpha", 40)
	actualDispatched := []Msg(nil)

	executePrepareSelectedDetailClipboardWriteCommand(linkClipboardCommandRuntime{
		dispatch: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		currentDetailDocument: func(view *gocui.View) detailDocument {
			return expectedDocument
		},
	}, nil, prepareSelectedDetailClipboardWriteCmd{Target: FocusDetailView})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgSelectedDetailClipboardPrepared)
	if !ok {
		t.Fatalf("expected a MsgSelectedDetailClipboardPrepared, actual %T", actualDispatched[0])
	}
	if actual := message.Document.id; actual != expectedDocument.id {
		t.Fatalf("expected resolved document %d, actual %d", expectedDocument.id, actual)
	}
	if actual := message.ViewportHeight; actual != 1 {
		t.Fatalf("expected fallback viewport height %d, actual %d", 1, actual)
	}
	if actual := message.Target; actual != FocusDetailView {
		t.Fatalf("expected feedback target %v, actual %v", FocusDetailView, actual)
	}
}

func TestPreparePullRequestBuildRunPopupClipboardWriteCommand_GivenResolvedDocument_WhenExecuting_ThenItDispatchesTheTypedPreparedMessage(t *testing.T) {
	expectedDocument := newDetailDocumentWithWrap("Alpha", 40, false)
	actualDispatched := []Msg(nil)

	executePreparePullRequestBuildRunPopupClipboardWriteCommand(linkClipboardCommandRuntime{
		dispatch: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		currentPullRequestBuildRunPopupDocument: func(view *gocui.View) detailDocument {
			return expectedDocument
		},
	}, nil, preparePullRequestBuildRunPopupClipboardWriteCmd{Target: FocusDetailView})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgPullRequestBuildRunPopupClipboardPrepared)
	if !ok {
		t.Fatalf("expected a MsgPullRequestBuildRunPopupClipboardPrepared, actual %T", actualDispatched[0])
	}
	if actual := message.Document.id; actual != expectedDocument.id {
		t.Fatalf("expected resolved document %d, actual %d", expectedDocument.id, actual)
	}
	if actual := message.ViewportHeight; actual != 1 {
		t.Fatalf("expected fallback viewport height %d, actual %d", 1, actual)
	}
	if actual := message.Target; actual != FocusDetailView {
		t.Fatalf("expected feedback target %v, actual %v", FocusDetailView, actual)
	}
}
