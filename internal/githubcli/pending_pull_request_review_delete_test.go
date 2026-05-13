package githubcli

import (
	"errors"
	"testing"
)

func TestDeletePullRequestReview_GivenReviewID_WhenDeleting_ThenItRunsGhApiGraphQLWithTheDeleteMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"deletePullRequestReview":{"pullRequestReview":{"id":"PRR_1"}}}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.DeletePullRequestReview("PRR_1")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"api", "graphql", "-f", "query=" + deletePullRequestReviewMutation, "-f", "pullRequestReviewId=PRR_1"})
}

func TestDeletePullRequestReview_GivenMissingReviewID_WhenDeleting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewReviewServiceWithRunner(&fakeRunner{})

	actualErr := subject.DeletePullRequestReview(" ")

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewDeletion) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewDeletion, actualErr)
	}
}

func TestDeletePullRequestReview_GivenGraphQLError_WhenDeleting_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"Review is not pending"}]}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.DeletePullRequestReview("PRR_1")

	if actualErr == nil || actualErr.Error() != "Review is not pending" {
		t.Fatalf("expected error %q, actual %v", "Review is not pending", actualErr)
	}
}
