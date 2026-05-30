package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestModalEditorExternalEditCommand_GivenEditedText_WhenExecuting_ThenItAppliesTheFinishedRuntimeMessage(t *testing.T) {
	actualMessages := []Msg(nil)
	editor := &fakeExternalEditor{editedText: "Edited in $EDITOR"}

	executeModalEditorExternalEditCommand(modalEditorCommandRuntime{
		externalEditor: editor,
		executeMessage: func(gui *gocui.Gui, msg Msg) error {
			actualMessages = append(actualMessages, msg)
			return nil
		},
	}, nil, modalEditorExternalEditCmd{Text: "Draft inline comment"})

	if actual := editor.receivedText; actual != "Draft inline comment" {
		t.Fatalf("expected external editor input %q, actual %q", "Draft inline comment", actual)
	}
	if len(actualMessages) != 1 {
		t.Fatalf("expected one finished message, actual %d", len(actualMessages))
	}
	message, ok := actualMessages[0].(MsgModalEditorExternalEditFinished)
	if !ok {
		t.Fatalf("expected a MsgModalEditorExternalEditFinished, actual %T", actualMessages[0])
	}
	if actual := message.Text; actual != "Edited in $EDITOR" {
		t.Fatalf("expected edited text %q, actual %q", "Edited in $EDITOR", actual)
	}
	if message.Err != nil {
		t.Fatalf("expected no error, actual %v", message.Err)
	}
}

func TestModalEditorExternalEditCommand_GivenMissingEditor_WhenExecuting_ThenItAppliesTheFinishedErrorMessage(t *testing.T) {
	actualMessages := []Msg(nil)

	executeModalEditorExternalEditCommand(modalEditorCommandRuntime{
		executeMessage: func(gui *gocui.Gui, msg Msg) error {
			actualMessages = append(actualMessages, msg)
			return nil
		},
	}, nil, modalEditorExternalEditCmd{Text: "Draft inline comment"})

	if len(actualMessages) != 1 {
		t.Fatalf("expected one finished message, actual %d", len(actualMessages))
	}
	message, ok := actualMessages[0].(MsgModalEditorExternalEditFinished)
	if !ok {
		t.Fatalf("expected a MsgModalEditorExternalEditFinished, actual %T", actualMessages[0])
	}
	if message.Err == nil {
		t.Fatal("expected an unavailable-editor error")
	}
	if actual := message.Err.Error(); actual != "external editor is unavailable" {
		t.Fatalf("expected error %q, actual %q", "external editor is unavailable", actual)
	}
}

func TestModalEditorSubmitCommand_GivenCompletion_WhenExecuting_ThenItAppliesTheFinishedRuntimeMessage(t *testing.T) {
	actualMessages := []Msg(nil)
	expectedCompletion := pullRequestCommentSubmittedCompletion(MsgPullRequestCommentSubmitted{Body: "Submitted body"})

	executeModalEditorSubmitCommand(modalEditorCommandRuntime{
		executeMessage: func(gui *gocui.Gui, msg Msg) error {
			actualMessages = append(actualMessages, msg)
			return nil
		},
	}, nil, modalEditorSubmitCmd{request: fakeModalEditorSubmitRequest{completion: expectedCompletion}})

	if len(actualMessages) != 1 {
		t.Fatalf("expected one finished message, actual %d", len(actualMessages))
	}
	message, ok := actualMessages[0].(MsgModalEditorSubmitFinished)
	if !ok {
		t.Fatalf("expected a MsgModalEditorSubmitFinished, actual %T", actualMessages[0])
	}
	if message.Err != nil {
		t.Fatalf("expected no error, actual %v", message.Err)
	}
	completion, ok := message.Completion.(pullRequestCommentSubmittedCompletion)
	if !ok {
		t.Fatalf("expected a pullRequestCommentSubmittedCompletion, actual %T", message.Completion)
	}
	if actual := MsgPullRequestCommentSubmitted(completion).Body; actual != "Submitted body" {
		t.Fatalf("expected submitted body %q, actual %q", "Submitted body", actual)
	}
}

func TestModalEditorSubmitCommand_GivenAsyncRequest_WhenExecuting_ThenItUsesTheAsyncRunnerAndDispatchesTheFinishedMessageAsynchronously(t *testing.T) {
	actualMessages := []Msg(nil)
	actualRuns := 0
	expectedCompletion := pullRequestCommentSubmittedCompletion(MsgPullRequestCommentSubmitted{Body: "Submitted body"})

	executeModalEditorSubmitCommand(modalEditorCommandRuntime{
		dispatchAsyncMessage: func(msg Msg) {
			actualMessages = append(actualMessages, msg)
		},
		runAsync: func(run func()) {
			actualRuns++
			run()
		},
	}, nil, modalEditorSubmitCmd{request: fakeModalEditorSubmitRequest{completion: expectedCompletion, async: true}})

	if actual := actualRuns; actual != 1 {
		t.Fatalf("expected one async run, actual %d", actual)
	}
	if len(actualMessages) != 1 {
		t.Fatalf("expected one async finished message, actual %d", len(actualMessages))
	}
	message, ok := actualMessages[0].(MsgModalEditorSubmitFinished)
	if !ok {
		t.Fatalf("expected a MsgModalEditorSubmitFinished, actual %T", actualMessages[0])
	}
	if message.Err != nil {
		t.Fatalf("expected no error, actual %v", message.Err)
	}
}

type fakeModalEditorSubmitRequest struct {
	completion      modalEditorSubmitCompletion
	err             error
	async           bool
	statusLineValue string
}

func (request fakeModalEditorSubmitRequest) statusCommand() string {
	return request.statusLineValue
}

func (request fakeModalEditorSubmitRequest) asyncRequested() bool {
	return request.async
}

func (request fakeModalEditorSubmitRequest) run(modalEditorSubmitCommandDeps) (modalEditorSubmitCompletion, error) {
	return request.completion, request.err
}
