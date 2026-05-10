package githubcli

import (
	"errors"
	"testing"
)

func TestRequestPullRequestReviewer_GivenReviewerLogin_WhenRequesting_ThenItRunsGhPrEditWithAddReviewer(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.RequestPullRequestReviewer("acme/widgets", 42, " reviewer-one ")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "edit", "42", "-R", "acme/widgets", "--add-reviewer", "reviewer-one"})
}

func TestRequestPullRequestReviewer_GivenMissingReviewerLogin_WhenRequesting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	actualErr := subject.RequestPullRequestReviewer("acme/widgets", 42, " \n\t ")

	if !errors.Is(actualErr, ErrMissingPullRequestReviewer) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestReviewer, actualErr)
	}
}
