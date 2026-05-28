package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestUpdate_GivenMsgOpenLinkUnderCursorResolvedWithoutLink_WhenApplying_ThenItSetsUnavailableFeedback(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgOpenLinkUnderCursorResolved{Target: FocusDetailView, OpenerAvailable: true})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.feedbackMessage; actual != openLinkUnavailableMessage {
		t.Fatalf("expected feedback %q, actual %q", openLinkUnavailableMessage, actual)
	}
}

func TestUpdate_GivenMsgOpenLinkUnderCursorResolvedWithURLAndOpener_WhenApplying_ThenItReturnsATypedBrowserOpenCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgOpenLinkUnderCursorResolved{Target: FocusDetailView, URL: "https://example.com/docs", LinkAvailable: true, OpenerAvailable: true})

	if len(actual) != 1 {
		t.Fatalf("expected one browser-open command, actual %d", len(actual))
	}
	command, ok := actual[0].(openBrowserURLCmd)
	if !ok {
		t.Fatalf("expected an openBrowserURLCmd, actual %T", actual[0])
	}
	if actual := command.URL; actual != "https://example.com/docs" {
		t.Fatalf("expected browser URL %q, actual %q", "https://example.com/docs", actual)
	}
	if actual := command.SuccessMessage; actual != openLinkSuccessMessage {
		t.Fatalf("expected success message %q, actual %q", openLinkSuccessMessage, actual)
	}
	if actual := command.FailureMessage; actual != openLinkFailureMessage {
		t.Fatalf("expected failure message %q, actual %q", openLinkFailureMessage, actual)
	}
	if actual := command.Target; actual != FocusDetailView {
		t.Fatalf("expected feedback target %v, actual %v", FocusDetailView, actual)
	}
}

func TestUpdate_GivenMsgOpenPullRequestBuildRunPopupLinkResolvedWithoutOpener_WhenApplying_ThenItSetsUnavailableFeedback(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgOpenPullRequestBuildRunPopupLinkResolved{Target: FocusDetailView, URL: "https://example.com/build/42", LinkAvailable: true})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.feedbackMessage; actual != openLinkOpenerUnavailableMessage {
		t.Fatalf("expected feedback %q, actual %q", openLinkOpenerUnavailableMessage, actual)
	}
}

func TestOpenLinkUnderCursorCommand_GivenResolvedLinkState_WhenExecuting_ThenItDispatchesTheTypedResolvedMessage(t *testing.T) {
	actualDispatched := []Msg(nil)

	executeOpenLinkUnderCursorCommand(linkClipboardCommandRuntime{
		executeMessage: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		currentDetailCursorLink: func() (string, bool) {
			return "https://example.com/docs", true
		},
	}, nil, openLinkUnderCursorCmd{Target: FocusDetailView})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgOpenLinkUnderCursorResolved)
	if !ok {
		t.Fatalf("expected a MsgOpenLinkUnderCursorResolved, actual %T", actualDispatched[0])
	}
	if actual := message.URL; actual != "https://example.com/docs" {
		t.Fatalf("expected resolved URL %q, actual %q", "https://example.com/docs", actual)
	}
	if !message.LinkAvailable {
		t.Fatal("expected the resolved message to preserve link availability")
	}
	if message.OpenerAvailable {
		t.Fatal("expected the resolved message to preserve missing opener availability")
	}
}

func TestOpenPullRequestBuildRunPopupLinkCommand_GivenResolvedLinkState_WhenExecuting_ThenItDispatchesTheTypedResolvedMessage(t *testing.T) {
	actualDispatched := []Msg(nil)

	executeOpenPullRequestBuildRunPopupLinkCommand(linkClipboardCommandRuntime{
		linkOpener: &fakeLinkOpener{},
		executeMessage: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		resolveView: func(gui *gocui.Gui, current *gocui.View, name string) *gocui.View {
			return current
		},
		currentPullRequestBuildRunPopupLink: func(view *gocui.View) (string, bool) {
			return "https://example.com/build/42", true
		},
	}, nil, openPullRequestBuildRunPopupLinkCmd{Target: FocusDetailView})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgOpenPullRequestBuildRunPopupLinkResolved)
	if !ok {
		t.Fatalf("expected a MsgOpenPullRequestBuildRunPopupLinkResolved, actual %T", actualDispatched[0])
	}
	if actual := message.URL; actual != "https://example.com/build/42" {
		t.Fatalf("expected resolved URL %q, actual %q", "https://example.com/build/42", actual)
	}
	if !message.LinkAvailable {
		t.Fatal("expected the resolved message to preserve link availability")
	}
	if !message.OpenerAvailable {
		t.Fatal("expected the resolved message to preserve opener availability")
	}
}
