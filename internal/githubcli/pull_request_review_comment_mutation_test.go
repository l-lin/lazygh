package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestUpdatePullRequestReviewComment_GivenCommentIDAndBody_WhenUpdating_ThenItRunsGhApiGraphQLWithTheUpdateMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"updatePullRequestReviewComment":{"pullRequestReviewComment":{"id":"PRRC_1"}}}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestReviewComment("PRRC_1", "Updated body")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + updatePullRequestReviewCommentMutation,
		"-f",
		"pullRequestReviewCommentId=PRRC_1",
		"-f",
		"body=Updated body",
	})
}

func TestDeletePullRequestReviewComment_GivenCommentID_WhenDeleting_ThenItRunsGhApiGraphQLWithTheDeleteMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"deletePullRequestReviewComment":{"pullRequestReviewComment":{"id":"PRRC_1"}}}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.DeletePullRequestReviewComment("PRRC_1")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + deletePullRequestReviewCommentMutation,
		"-f",
		"id=PRRC_1",
	})
}

func TestUpdatePullRequestReviewComment_GivenGraphQLErrorPayload_WhenUpdating_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"Only the author can update this comment"}],"data":{"updatePullRequestReviewComment":null}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestReviewComment("PRRC_1", "Updated body")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "Only the author can update this comment" {
		t.Fatalf("expected GitHub error %q, actual %v", "Only the author can update this comment", actualErr)
	}
}

func TestDeletePullRequestReviewComment_GivenAnEmptyCommentID_WhenDeleting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewReviewServiceWithRunner(&fakeRunner{})

	actualErr := subject.DeletePullRequestReviewComment(" ")

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewCommentMutation) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewCommentMutation, actualErr)
	}
}

func TestUpdatePullRequestReviewComment_GivenCommandFailure_WhenUpdating_ThenItReturnsTheGhApiGraphQLError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestReviewComment("PRRC_1", "Updated body")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh api graphql") {
		t.Fatalf("expected error to mention %q, actual %v", "gh api graphql", actualErr)
	}
}
