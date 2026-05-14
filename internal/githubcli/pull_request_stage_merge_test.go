package githubcli

import (
	"errors"
	"strings"
	"testing"
)

func TestMarkPullRequestReadyForReview_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrReady(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.MarkPullRequestReadyForReview("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "ready", "42", "-R", "acme/widgets"})
}

func TestConvertPullRequestToDraft_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrReadyUndo(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.ConvertPullRequestToDraft("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "ready", "42", "-R", "acme/widgets", "--undo"})
}

func TestClosePullRequest_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrClose(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.ClosePullRequest("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "close", "42", "-R", "acme/widgets"})
}

func TestReopenPullRequest_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrReopen(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.ReopenPullRequest("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "reopen", "42", "-R", "acme/widgets"})
}

func TestSquashMergePullRequest_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrMergeSquash(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.SquashMergePullRequest("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "merge", "42", "-R", "acme/widgets", "--squash"})
}

func TestUpdatePullRequestBranch_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrUpdateBranch(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestBranch("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "update-branch", "42", "-R", "acme/widgets"})
}

func TestMarkPullRequestReadyForReview_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrReadyError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.MarkPullRequestReadyForReview("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr ready") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr ready", actualErr)
	}
}

func TestClosePullRequest_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrCloseError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.ClosePullRequest("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr close") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr close", actualErr)
	}
}

func TestReopenPullRequest_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrReopenError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.ReopenPullRequest("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr reopen") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr reopen", actualErr)
	}
}

func TestSquashMergePullRequest_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrMergeError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.SquashMergePullRequest("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr merge") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr merge", actualErr)
	}
}

func TestUpdatePullRequestBranch_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrUpdateBranchError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.UpdatePullRequestBranch("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr update-branch") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr update-branch", actualErr)
	}
}
