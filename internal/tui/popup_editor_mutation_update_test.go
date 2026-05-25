package tui

import "testing"

func TestUpdate_GivenMsgPullRequestCommentSubmitRequested_WhenApplying_ThenItBuildsAModalEditorCommandWithReducerOwnedSuccess(t *testing.T) {
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
	if command.Submit == nil {
		t.Fatal("expected the submit command to include work")
	}
	if command.AfterSubmit == nil {
		t.Fatal("expected the submit command to include a reducer-owned success hook")
	}
	if actual := command.Text; actual != "Ship it" {
		t.Fatalf("expected submit text %q, actual %q", "Ship it", actual)
	}

	followUpCommands := command.AfterSubmit(subject)

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
	if len(followUpCommands) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(followUpCommands))
	}
}
