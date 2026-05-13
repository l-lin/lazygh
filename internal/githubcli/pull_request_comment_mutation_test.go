package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestUpdatePullRequestComment_GivenCommentIDAndBody_WhenUpdating_ThenItRunsGhApiGraphQLWithTheUpdateMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"updateIssueComment":{"issueComment":{"id":"IC_kwDOA"}}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestComment("IC_kwDOA", "Updated body")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + updatePullRequestCommentMutation,
		"-f",
		"id=IC_kwDOA",
		"-f",
		"body=Updated body",
	})
}

func TestDeletePullRequestComment_GivenCommentID_WhenDeleting_ThenItRunsGhApiGraphQLWithTheDeleteMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"deleteIssueComment":{"clientMutationId":null}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.DeletePullRequestComment("IC_kwDOA")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + deletePullRequestCommentMutation,
		"-f",
		"id=IC_kwDOA",
	})
}

func TestUpdatePullRequestComment_GivenGraphQLErrorPayload_WhenUpdating_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"Only the author can update this comment"}],"data":{"updateIssueComment":null}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestComment("IC_kwDOA", "Updated body")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "Only the author can update this comment" {
		t.Fatalf("expected GitHub error %q, actual %v", "Only the author can update this comment", actualErr)
	}
}

func TestDeletePullRequestComment_GivenAnEmptyCommentID_WhenDeleting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewPullRequestMutationServiceWithRunner(&fakeRunner{})

	actualErr := subject.DeletePullRequestComment(" ")

	if !errors.Is(actualErr, ErrInvalidPullRequestCommentMutation) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestCommentMutation, actualErr)
	}
}

func TestUpdatePullRequestComment_GivenCommandFailure_WhenUpdating_ThenItReturnsTheGhApiGraphQLError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestComment("IC_kwDOA", "Updated body")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh api graphql") {
		t.Fatalf("expected error to mention %q, actual %v", "gh api graphql", actualErr)
	}
}
