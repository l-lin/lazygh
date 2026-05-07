package githubcli

import (
	"errors"
	"testing"
)

func TestGetPullRequestBuildRun_GivenABuildRunLink_WhenFetching_ThenItUsesGhRunViewVerbose(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("Run #42\nStatus: completed\nConclusion: failure\n")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestBuildRun("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "42", "-R", "acme/widgets", "--verbose"})
	if actual != "Run #42\nStatus: completed\nConclusion: failure" {
		t.Fatalf("expected run output %q, actual %q", "Run #42\nStatus: completed\nConclusion: failure", actual)
	}
}

func TestGetPullRequestBuildRun_GivenABuildRunAttemptLink_WhenFetching_ThenItPassesTheAttemptNumber(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("Run #42 attempt 3")}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetPullRequestBuildRun("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42/attempts/3/job/99"})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "42", "-R", "acme/widgets", "--attempt", "3", "--verbose"})
}

func TestGetPullRequestBuildRun_GivenAMissingBuildRunLink_WhenFetching_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	_, actualErr := subject.GetPullRequestBuildRun("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI"})

	if !errors.Is(actualErr, ErrMissingPullRequestBuildLink) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestBuildLink, actualErr)
	}
}
