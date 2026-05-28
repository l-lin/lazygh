package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestReviewInlineCommentSubmitRequest_GivenNewPendingReview_WhenRunning_ThenItReturnsThePreparedMessageWithoutSubmittingTheThread(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_new"}
	request := reviewInlineCommentSubmitRequest{
		target: pullRequestInlineCommentTarget{
			repository: "acme/widgets",
			number:     42,
			threadTarget: githubdomain.PullRequestReviewThreadTarget{
				Path:        "internal/tui/render.go",
				Line:        2,
				Side:        "RIGHT",
				SubjectType: "LINE",
			},
		},
		body: "Please add context",
	}

	actual, actualErr := request.run(modalEditorSubmitCommandDeps{reviewMutations: loader})

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if len(loader.startReviewCalls) != 1 || loader.startReviewCalls[0] != "acme/widgets#42" {
		t.Fatalf("expected one pending-review start call, actual %v", loader.startReviewCalls)
	}
	if len(loader.reviewThreadReviewIDs) != 0 {
		t.Fatalf("expected no thread submit calls before the reducer records the pending review, actual %v", loader.reviewThreadReviewIDs)
	}
	message, ok := actual.(reviewInlineCommentPendingReviewPreparedCompletion)
	if !ok {
		t.Fatalf("expected a reviewInlineCommentPendingReviewPreparedCompletion, actual %T", actual)
	}
	if actual := message.Target.pendingReview; actual != "PRR_new" {
		t.Fatalf("expected prepared pending review id %q, actual %q", "PRR_new", actual)
	}
	if actual := message.Body; actual != "Please add context" {
		t.Fatalf("expected prepared body %q, actual %q", "Please add context", actual)
	}
}

func TestPreparedReviewInlineCommentSubmitRequest_GivenPreparedTarget_WhenRunning_ThenItSubmitsTheThreadAndReturnsTheSuccessMessage(t *testing.T) {
	loader := &fakePullRequestDetailLoader{}
	target := pullRequestInlineCommentTarget{
		repository:    "acme/widgets",
		number:        42,
		pendingReview: "PRR_pending",
		threadTarget: githubdomain.PullRequestReviewThreadTarget{
			Path:        "internal/tui/render.go",
			Line:        2,
			Side:        "RIGHT",
			SubjectType: "LINE",
		},
	}
	request := preparedReviewInlineCommentSubmitRequest{target: target, body: "Please add context"}

	actual, actualErr := request.run(modalEditorSubmitCommandDeps{reviewMutations: loader})

	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if len(loader.reviewThreadReviewIDs) != 1 || loader.reviewThreadReviewIDs[0] != "PRR_pending" {
		t.Fatalf("expected one thread submit call with %q, actual %v", "PRR_pending", loader.reviewThreadReviewIDs)
	}
	if len(loader.reviewThreadBodies) != 1 || loader.reviewThreadBodies[0] != "Please add context" {
		t.Fatalf("expected one thread body %q, actual %v", "Please add context", loader.reviewThreadBodies)
	}
	message, ok := actual.(reviewInlineCommentSubmittedCompletion)
	if !ok {
		t.Fatalf("expected a reviewInlineCommentSubmittedCompletion, actual %T", actual)
	}
	if actual := message.Target.pendingReview; actual != "PRR_pending" {
		t.Fatalf("expected submitted pending review id %q, actual %q", "PRR_pending", actual)
	}
}

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithPreparedReviewInlineCommentCompletion_WhenApplying_ThenItKeepsTheModalOpenRecordsThePendingReviewAndReturnsAFollowUpSubmitCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorState("Inline comment", "Please add context")
	prepared := reviewInlineCommentPendingReviewPreparedCompletion{
		Target: pullRequestInlineCommentTarget{
			repository:    "acme/widgets",
			number:        42,
			pendingReview: "PRR_pending",
			threadTarget: githubdomain.PullRequestReviewThreadTarget{
				Path:        "internal/tui/render.go",
				Line:        2,
				Side:        "RIGHT",
				SubjectType: "LINE",
			},
		},
		Body: "Please add context",
	}

	actual := Update(subject, MsgModalEditorSubmitFinished{Completion: prepared})

	if !subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to stay open until the follow-up inline-comment submit finishes")
	}
	pendingState, ok := subject.pendingPullRequestReviewCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the pending review state to be cached before the follow-up submit command runs")
	}
	if actual := pendingState.id; actual != "PRR_pending" {
		t.Fatalf("expected cached pending review id %q, actual %q", "PRR_pending", actual)
	}
	if len(actual) != 1 {
		t.Fatalf("expected one follow-up submit command, actual %d", len(actual))
	}
	command, ok := actual[0].(modalEditorSubmitCmd)
	if !ok {
		t.Fatalf("expected a modalEditorSubmitCmd, actual %T", actual[0])
	}
	request, ok := command.request.(preparedReviewInlineCommentSubmitRequest)
	if !ok {
		t.Fatalf("expected a preparedReviewInlineCommentSubmitRequest, actual %T", command.request)
	}
	if actual := request.target.pendingReview; actual != "PRR_pending" {
		t.Fatalf("expected follow-up pending review id %q, actual %q", "PRR_pending", actual)
	}
	if actual := request.body; actual != "Please add context" {
		t.Fatalf("expected follow-up body %q, actual %q", "Please add context", actual)
	}
}
