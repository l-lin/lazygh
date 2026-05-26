package tui

import "testing"

func TestUpdate_GivenMsgPullRequestCommentSubmitRequested_WhenApplying_ThenItBuildsATypedModalEditorSubmitRequest(t *testing.T) {
	summary := given_pullRequestMutationSummary("OPEN", false)
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: subject.optimisticPullRequestDetailSeed(summary)}

	actual := Update(subject, MsgPullRequestCommentSubmitRequested{
		Target:         pullRequestCommentTarget{repository: "acme/widgets", number: 42},
		Body:           "Ship it",
		FeedbackTarget: FocusDetailView,
	})

	if len(actual) != 1 {
		t.Fatalf("expected one submit command, actual %d", len(actual))
	}

	command, ok := actual[0].(modalEditorSubmitCmd)
	if !ok {
		t.Fatalf("expected a modalEditorSubmitCmd, actual %T", actual[0])
	}

	request, ok := command.request.(pullRequestCommentSubmitRequest)
	if !ok {
		t.Fatalf("expected a typed pullRequestCommentSubmitRequest, actual %T", command.request)
	}
	if actual := request.target.repository; actual != "acme/widgets" {
		t.Fatalf("expected request repository %q, actual %q", "acme/widgets", actual)
	}
	if actual := request.target.number; actual != 42 {
		t.Fatalf("expected request number %d, actual %d", 42, actual)
	}
	if actual := request.body; actual != "Ship it" {
		t.Fatalf("expected request body %q, actual %q", "Ship it", actual)
	}
	if actual := request.feedbackTarget; actual != FocusDetailView {
		t.Fatalf("expected feedback target %v, actual %v", FocusDetailView, actual)
	}
}

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithTypedSuccess_WhenApplying_ThenItAppliesSuccessAndClosesTheModal(t *testing.T) {
	summary := given_pullRequestMutationSummary("OPEN", false)
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorState("Comment", "Ship it")
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: subject.optimisticPullRequestDetailSeed(summary)}

	actual := Update(subject, MsgModalEditorSubmitFinished{Success: pullRequestCommentSubmitSuccess{
		Target:         pullRequestCommentTarget{repository: "acme/widgets", number: 42},
		Body:           "Ship it",
		FeedbackTarget: FocusDetailView,
	}})

	if subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to close after a successful typed submit")
	}

	cachedDetail, ok := subject.pullRequestDetailCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the pull request detail cache to stay populated")
	}
	if len(cachedDetail.detail.Comments) != 1 {
		t.Fatalf("expected one optimistic comment, actual %d", len(cachedDetail.detail.Comments))
	}
	if actual := cachedDetail.detail.Comments[0].Body; actual != "Ship it" {
		t.Fatalf("expected optimistic comment body %q, actual %q", "Ship it", actual)
	}
	if actual := subject.feedbackMessage; actual != pullRequestCommentSuccessMessage {
		t.Fatalf("expected feedback %q, actual %q", pullRequestCommentSuccessMessage, actual)
	}
	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
}
