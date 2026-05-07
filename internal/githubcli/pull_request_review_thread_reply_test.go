package githubcli

import (
	"errors"
	"testing"
)

func TestAddPullRequestReviewThreadReply_GivenAPendingReview_WhenSubmitting_ThenItRunsGhApiGraphQLWithTheReplyInput(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"addPullRequestReviewThreadReply":{"comment":{"id":"PRRC_2"}}}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.AddPullRequestReviewThreadReply("PRR_pending", "PRRT_1", "Please add context")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + addPullRequestReviewThreadReplyMutation,
		"-f",
		"pullRequestReviewThreadId=PRRT_1",
		"-f",
		"body=Please add context",
		"-f",
		"pullRequestReviewId=PRR_pending",
	})
}

func TestAddPullRequestReviewThreadReply_GivenAStandaloneReply_WhenSubmitting_ThenItOmitsThePendingReviewID(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"addPullRequestReviewThreadReply":{"comment":{"id":"PRRC_2"}}}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.AddPullRequestReviewThreadReply("", "PRRT_1", "Standalone reply")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + addPullRequestReviewThreadReplyMutation,
		"-f",
		"pullRequestReviewThreadId=PRRT_1",
		"-f",
		"body=Standalone reply",
	})
}

func TestAddPullRequestReviewThreadReply_GivenGraphQLErrorPayload_WhenSubmitting_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"Pull request review thread reply body cannot be blank"}],"data":{"addPullRequestReviewThreadReply":null}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.AddPullRequestReviewThreadReply("PRR_pending", "PRRT_1", "Please add context")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "Pull request review thread reply body cannot be blank" {
		t.Fatalf("expected GitHub error %q, actual %v", "Pull request review thread reply body cannot be blank", actualErr)
	}
}

func TestAddPullRequestReviewThreadReply_GivenAMissingThreadID_WhenSubmitting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.AddPullRequestReviewThreadReply("PRR_pending", " ", "Please add context")

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewThreadReply) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewThreadReply, actualErr)
	}
}
