package tui

import (
	"errors"
	"testing"
)

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
	if actual := subject.ghCommandLoadingMessage; actual != formatRunningCommandStatus("gh pr comment 42 -R acme/widgets --body-file -") {
		t.Fatalf("expected gh command loading message %q, actual %q", formatRunningCommandStatus("gh pr comment 42 -R acme/widgets --body-file -"), actual)
	}
}

func TestUpdate_GivenMsgPullRequestTitleEditRequested_WhenApplying_ThenItStartsGHLoadingAndQueuesATypedAsyncSubmitCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgPullRequestTitleEditRequested{
		Target:         pullRequestActionTarget{repository: "acme/widgets", number: 42},
		Title:          "Rename me",
		FeedbackTarget: FocusDetailView,
	})

	if len(actual) != 1 {
		t.Fatalf("expected one submit command, actual %d", len(actual))
	}

	command, ok := actual[0].(modalEditorSubmitCmd)
	if !ok {
		t.Fatalf("expected a modalEditorSubmitCmd, actual %T", actual[0])
	}

	request, ok := command.request.(pullRequestTitleEditSubmitRequest)
	if !ok {
		t.Fatalf("expected a typed pullRequestTitleEditSubmitRequest, actual %T", command.request)
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
	if actual := subject.ghCommandLoadingMessage; actual != formatRunningCommandStatus("gh pr edit 42 -R acme/widgets --title Rename me") {
		t.Fatalf("expected gh command loading message %q, actual %q", formatRunningCommandStatus("gh pr edit 42 -R acme/widgets --title Rename me"), actual)
	}
}

func TestUpdate_GivenMsgReviewInlineCommentPendingReviewPrepared_WhenApplying_ThenItStartsGHLoadingForTheFollowUpSubmitCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgReviewInlineCommentPendingReviewPrepared{
		Target: pullRequestInlineCommentTarget{repository: "acme/widgets", number: 42, pendingReview: "PRR_pending"},
		Body:   "Please add context",
	})

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
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", actual)
	}
	if actual := request.body; actual != "Please add context" {
		t.Fatalf("expected body %q, actual %q", "Please add context", actual)
	}
	if actual := subject.ghCommandLoadingMessage; actual != formatRunningCommandStatus("gh api graphql") {
		t.Fatalf("expected gh command loading message %q, actual %q", formatRunningCommandStatus("gh api graphql"), actual)
	}
}

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithTypedCompletion_WhenApplying_ThenItAppliesTheCompletionClearsLoadingAndClosesTheModal(t *testing.T) {
	summary := given_pullRequestMutationSummary("OPEN", false)
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorState("Comment", "Ship it")
	subject.ghCommandLoadingMessage = "Running `gh pr comment 42 -R acme/widgets --body-file -`."
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: subject.optimisticPullRequestDetailSeed(summary)}

	actual := Update(subject, MsgModalEditorSubmitFinished{Completion: pullRequestCommentSubmittedCompletion{
		Target:         pullRequestCommentTarget{repository: "acme/widgets", number: 42},
		Body:           "Ship it",
		FeedbackTarget: FocusDetailView,
	}})

	if subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to close after a successful typed submit")
	}
	if actual := subject.ghCommandLoadingMessage; actual != "" {
		t.Fatalf("expected gh command loading message %q, actual %q", "", actual)
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

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithError_WhenApplying_ThenItClearsGHLoadingKeepsTheModalOpenAndShowsTheError(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorState("Comment", "Broken")
	subject.ghCommandLoadingMessage = "Running `gh pr edit 42 -R acme/widgets --title Broken`."

	actual := Update(subject, MsgModalEditorSubmitFinished{Err: errors.New("boom")})

	if !subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to stay open after a failed typed submit")
	}
	if actual := subject.ghCommandLoadingMessage; actual != "" {
		t.Fatalf("expected gh command loading message %q, actual %q", "", actual)
	}
	if actual := subject.overlayState.modalEditor.errorMessage; actual != "boom" {
		t.Fatalf("expected modal editor error %q, actual %q", "boom", actual)
	}
	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
}

func TestUpdate_GivenMsgModalEditorSubmitFinishedWithTypedCompletionReturningCommands_WhenApplying_ThenItReturnsTheCompletionCommandsAndClearsLoading(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.overlayState.modalEditor = newModalEditorState("Rename", "Retitle")
	subject.ghCommandLoadingMessage = "Running `gh pr edit 42 -R acme/widgets --title Retitle`."

	actual := Update(subject, MsgModalEditorSubmitFinished{Completion: pullRequestTitleEditAppliedCompletion{
		Target:         pullRequestActionTarget{repository: "acme/widgets", number: 42},
		Title:          "Retitle",
		FeedbackTarget: FocusDetailView,
	}})

	if len(actual) != 1 {
		t.Fatalf("expected one completion command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reloadPullRequestsTabCmd); !ok {
		t.Fatalf("expected a reloadPullRequestsTabCmd, actual %T", actual[0])
	}
	if subject.modalEditorVisible() {
		t.Fatal("expected the modal editor to close after a successful typed submit")
	}
	if actual := subject.ghCommandLoadingMessage; actual != "" {
		t.Fatalf("expected gh command loading message %q, actual %q", "", actual)
	}
}
