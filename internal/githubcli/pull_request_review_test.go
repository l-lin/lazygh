package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestApprovePullRequest_GivenRepositoryAndNumber_WhenApproving_ThenItRunsGhPrReviewApprove(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.ApprovePullRequest("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "review", "42", "-R", "acme/widgets", "--approve"})
}

func TestReviewPullRequestWithComment_GivenRepositoryNumberAndBody_WhenSubmitting_ThenItRunsGhPrReviewCommentWithStdin(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)
	body := "Please split this diff."

	actualErr := subject.ReviewPullRequestWithComment("acme/widgets", 42, body)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "review", "42", "-R", "acme/widgets", "--comment", "--body-file", "-"})
	then_stdinIs(t, runner, body)
}

func TestRequestChangesOnPullRequest_GivenRepositoryNumberAndBody_WhenSubmitting_ThenItRunsGhPrReviewRequestChangesWithStdin(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)
	body := "Needs tests."

	actualErr := subject.RequestChangesOnPullRequest("acme/widgets", 42, body)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "review", "42", "-R", "acme/widgets", "--request-changes", "--body-file", "-"})
	then_stdinIs(t, runner, body)
}

func TestApprovePullRequest_GivenMissingPullRequestIdentity_WhenApproving_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.ApprovePullRequest(" ", 0)

	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}
}

func TestReviewPullRequestWithComment_GivenEmptyBody_WhenSubmitting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.ReviewPullRequestWithComment("acme/widgets", 42, " \n\t ")

	if !errors.Is(actualErr, ErrEmptyPullRequestReviewBody) {
		t.Fatalf("expected error %v, actual %v", ErrEmptyPullRequestReviewBody, actualErr)
	}
}

func TestRequestChangesOnPullRequest_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrReviewError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.RequestChangesOnPullRequest("acme/widgets", 42, "Needs tests")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh pr review") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr review", actualErr)
	}
}
