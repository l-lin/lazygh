package githubcli

import (
	"errors"
	"testing"
)

func TestSubmitPullRequestReview_GivenApprovalEventAndSummaryBody_WhenSubmitting_ThenItRunsGhApiGraphQL(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"submitPullRequestReview":{"pullRequestReview":{"id":"PRR_pending"}}}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.SubmitPullRequestReview("PRR_pending", PullRequestReviewEventApprove, "LGTM")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + submitPullRequestReviewMutation,
		"-f",
		"pullRequestReviewId=PRR_pending",
		"-f",
		"event=APPROVE",
		"-f",
		"body=LGTM",
	})
}

func TestSubmitPullRequestReview_GivenBlankSummaryBody_WhenSubmitting_ThenItOmitsTheBodyField(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"submitPullRequestReview":{"pullRequestReview":{"id":"PRR_pending"}}}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.SubmitPullRequestReview("  PRR_pending  ", PullRequestReviewEventComment, " \n\t ")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + submitPullRequestReviewMutation,
		"-f",
		"pullRequestReviewId=PRR_pending",
		"-f",
		"event=COMMENT",
	})
}

func TestSubmitPullRequestReview_GivenGraphQLErrorPayload_WhenSubmitting_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"A review cannot be submitted right now"}],"data":{"submitPullRequestReview":null}}`)}
	subject := NewClientWithRunner(runner)

	actualErr := subject.SubmitPullRequestReview("PRR_pending", PullRequestReviewEventRequestChanges, "Needs tests")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "A review cannot be submitted right now" {
		t.Fatalf("expected GitHub error %q, actual %v", "A review cannot be submitted right now", actualErr)
	}
}

func TestSubmitPullRequestReview_GivenMissingPendingReviewID_WhenSubmitting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.SubmitPullRequestReview(" ", PullRequestReviewEventApprove, "")

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewSubmission) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewSubmission, actualErr)
	}
}
