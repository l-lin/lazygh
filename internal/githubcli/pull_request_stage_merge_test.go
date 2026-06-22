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

func TestEnablePullRequestAutoMerge_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrMergeAutoSquash(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.EnablePullRequestAutoMerge("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "merge", "42", "-R", "acme/widgets", "--auto", "--squash"})
}

func TestDisablePullRequestAutoMerge_GivenRepositoryAndNumber_WhenSubmitting_ThenItRunsGhPrMergeDisableAuto(t *testing.T) {
	runner := &fakeRunner{}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.DisablePullRequestAutoMerge("acme/widgets", 42)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "merge", "42", "-R", "acme/widgets", "--disable-auto"})
}

func TestMergePullRequestWhenReady_GivenRepositoryWithAutoMergeAllowed_WhenSubmitting_ThenItRunsGhPrMerge(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"data":{"repository":{"autoMergeAllowed":true}}}`)},
		{},
	}}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.MergePullRequestWhenReady("acme/widgets", 42, "")

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + repositoryMergeCapabilitiesQuery, "-F", "owner=acme", "-F", "name=widgets"}},
		{name: "gh", args: []string{"pr", "merge", "42", "-R", "acme/widgets"}},
	})
}

func TestMergePullRequestWhenReady_GivenRepositoryWithAutoMergeDisabled_WhenSubmitting_ThenItBypassesGhPrMergeAndRunsTheQueueMutation(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"data":{"repository":{"autoMergeAllowed":false}}}`)},
		{stdout: []byte(`{"data":{"enqueuePullRequest":{"mergeQueueEntry":{"id":"MQE_1","state":"QUEUED","position":1}}}}`)},
	}}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.MergePullRequestWhenReady("acme/widgets", 42, " PR_kwDOA ")

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + repositoryMergeCapabilitiesQuery, "-F", "owner=acme", "-F", "name=widgets"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + enqueuePullRequestMutation, "-F", "pullRequestId=PR_kwDOA"}},
	})
}

func TestMergePullRequestWhenReady_GivenRepositoryWithAutoMergeDisabledAndMissingPullRequestID_WhenSubmitting_ThenItReturnsTheMissingIDError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"repository":{"autoMergeAllowed":false}}}`)}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.MergePullRequestWhenReady("acme/widgets", 42, "   ")

	if !errors.Is(actualErr, ErrMissingPullRequestID) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestID, actualErr)
	}
	then_commandsAre(t, runner, []fakeCommandCall{{name: "gh", args: []string{"api", "graphql", "-f", "query=" + repositoryMergeCapabilitiesQuery, "-F", "owner=acme", "-F", "name=widgets"}}})
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

func TestEnablePullRequestAutoMerge_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrMergeError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.EnablePullRequestAutoMerge("acme/widgets", 42)

	if !strings.Contains(actualErr.Error(), "gh pr merge") {
		t.Fatalf("expected error to mention %q, actual %v", "gh pr merge", actualErr)
	}
}

func TestDisablePullRequestAutoMerge_GivenCommandFailure_WhenSubmitting_ThenItReturnsTheGhPrMergeError(t *testing.T) {
	runner := &fakeRunner{stderr: []byte("boom"), err: errors.New("exit status 1")}
	subject := NewPullRequestMutationServiceWithRunner(runner)

	actualErr := subject.DisablePullRequestAutoMerge("acme/widgets", 42)

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
