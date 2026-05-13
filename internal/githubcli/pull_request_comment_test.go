package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestCommentOnPullRequest_GivenRepositoryNumberAndMultilineBody_WhenPosting_ThenItRunsGhPrCommentWithStdin(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)
	body := "Looks good.\n\n```go\nfmt.Println(\"ok\")\n```"

	actualErr := subject.CommentOnPullRequest("acme/widgets", 42, body)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "comment", "42", "-R", "acme/widgets", "--body-file", "-"})
	then_stdinIs(t, runner, body)
}

func TestCommentOnPullRequest_GivenCommandFailure_WhenPosting_ThenItReturnsTheGhPrCommentError(t *testing.T) {
	runner := &fakeRunner{
		stderr: []byte("boom"),
		err:    errors.New("exit status 1"),
	}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.CommentOnPullRequest("acme/widgets", 42, "Ship it")

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh pr comment") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr comment", actualErr)
	}
}

func TestCommentOnPullRequest_GivenMissingPullRequestIdentity_WhenPosting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewPullRequestMutationServiceWithRunner(&fakeRunner{})

	actualErr := subject.CommentOnPullRequest(" ", 0, "Ship it")

	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}
}

func TestCommentOnPullRequest_GivenEmptyCommentBody_WhenPosting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewPullRequestMutationServiceWithRunner(&fakeRunner{})

	actualErr := subject.CommentOnPullRequest("acme/widgets", 42, " \n\t ")

	if !errors.Is(actualErr, ErrEmptyPullRequestComment) {
		t.Fatalf("expected error %v, actual %v", ErrEmptyPullRequestComment, actualErr)
	}
}
