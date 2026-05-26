package tui

import "testing"

func TestUpdate_GivenMsgModalEditorSubmitRequestedWithPullRequestCommentDescriptor_WhenApplying_ThenItBuildsATypedSubmitCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorStateWithSubmitDescriptor("Comment", "Ship it", newPullRequestCommentSubmitDescriptor(pullRequestCommentTarget{repository: "acme/widgets", number: 42}, FocusDetailView))

	actual := Update(subject, MsgModalEditorSubmitRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one submit command, actual %d", len(actual))
	}

	command, ok := actual[0].(modalEditorSubmitCmd)
	if !ok {
		t.Fatalf("expected a modalEditorSubmitCmd, actual %T", actual[0])
	}

	request, ok := command.request.(pullRequestCommentSubmitRequest)
	if !ok {
		t.Fatalf("expected a pullRequestCommentSubmitRequest, actual %T", command.request)
	}
	if actual := request.target.repository; actual != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", actual)
	}
	if actual := request.target.number; actual != 42 {
		t.Fatalf("expected pull request number %d, actual %d", 42, actual)
	}
	if actual := request.body; actual != "Ship it" {
		t.Fatalf("expected body %q, actual %q", "Ship it", actual)
	}
	if actual := request.feedbackTarget; actual != FocusDetailView {
		t.Fatalf("expected feedback target %v, actual %v", FocusDetailView, actual)
	}
}

func TestUpdate_GivenMsgModalEditorSubmitRequestedWithPullRequestTitleDescriptor_WhenApplying_ThenItUsesTheEditorTextAsTheTitle(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newLineModalEditorStateWithSubmitDescriptor(pullRequestTitleEditorTitle, "Rename me", newPullRequestTitleEditSubmitDescriptor(pullRequestActionTarget{repository: "acme/widgets", number: 42, title: "First PR"}, FocusDetailView))

	actual := Update(subject, MsgModalEditorSubmitRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one submit command, actual %d", len(actual))
	}

	command, ok := actual[0].(modalEditorSubmitCmd)
	if !ok {
		t.Fatalf("expected a modalEditorSubmitCmd, actual %T", actual[0])
	}

	request, ok := command.request.(pullRequestTitleEditSubmitRequest)
	if !ok {
		t.Fatalf("expected a pullRequestTitleEditSubmitRequest, actual %T", command.request)
	}
	if actual := request.title; actual != "Rename me" {
		t.Fatalf("expected title %q, actual %q", "Rename me", actual)
	}
	if actual := request.target.repository; actual != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", actual)
	}
	if actual := request.target.number; actual != 42 {
		t.Fatalf("expected pull request number %d, actual %d", 42, actual)
	}
	if actual := request.feedbackTarget; actual != FocusDetailView {
		t.Fatalf("expected feedback target %v, actual %v", FocusDetailView, actual)
	}
}
