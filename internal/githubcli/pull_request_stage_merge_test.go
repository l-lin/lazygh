package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestMarkPullRequestReadyForReview_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrReady(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.MarkPullRequestReadyForReview("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "ready", "42", "-R", "acme/widgets"})
}

func TestConvertPullRequestToDraft_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrReadyUndo(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.ConvertPullRequestToDraft("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "ready", "42", "-R", "acme/widgets", "--undo"})
}

func TestSquashMergePullRequest_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrMergeSquash(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewClientWithRunner(runner)

	actualErr := subject.SquashMergePullRequest("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "merge", "42", "-R", "acme/widgets", "--squash"})
}

func TestMarkPullRequestReadyForReview_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrReadyError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.MarkPullRequestReadyForReview("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr ready") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr ready", actualErr)
	}
}

func TestSquashMergePullRequest_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrMergeError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewClientWithRunner(runner)

	actualErr := subject.SquashMergePullRequest("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr merge") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr merge", actualErr)
	}
}
