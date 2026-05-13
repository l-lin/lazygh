package githubcli

import (
	"errors"
	"testing"
)

func TestResolvePullRequestReviewThread_GivenAThreadID_WhenSubmitting_ThenItRunsGhApiGraphQLWithTheResolveMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"resolveReviewThread":{"thread":{"id":"PRRT_1","isResolved":true}}}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.ResolvePullRequestReviewThread("PRRT_1")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + resolveReviewThreadMutation,
		"-f",
		"threadId=PRRT_1",
	})
}

func TestUnresolvePullRequestReviewThread_GivenAThreadID_WhenSubmitting_ThenItRunsGhApiGraphQLWithTheUnresolveMutation(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"unresolveReviewThread":{"thread":{"id":"PRRT_1","isResolved":false}}}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.UnresolvePullRequestReviewThread("PRRT_1")

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{
		"api",
		"graphql",
		"-f",
		"query=" + unresolveReviewThreadMutation,
		"-f",
		"threadId=PRRT_1",
	})
}

func TestResolvePullRequestReviewThread_GivenAGraphQLErrorPayload_WhenSubmitting_ThenItReturnsTheGitHubMessage(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"errors":[{"message":"You cannot resolve this thread"}],"data":{"resolveReviewThread":null}}`)}
	subject := NewReviewServiceWithRunner(runner)

	actualErr := subject.ResolvePullRequestReviewThread("PRRT_1")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if actualErr.Error() != "You cannot resolve this thread" {
		t.Fatalf("expected GitHub error %q, actual %v", "You cannot resolve this thread", actualErr)
	}
}

func TestResolvePullRequestReviewThread_GivenAnEmptyThreadID_WhenSubmitting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewReviewServiceWithRunner(&fakeRunner{})

	actualErr := subject.ResolvePullRequestReviewThread(" ")

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewThreadMutation) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewThreadMutation, actualErr)
	}
}
