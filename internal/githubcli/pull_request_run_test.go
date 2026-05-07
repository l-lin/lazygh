package githubcli

import (
	"errors"
	"reflect"
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

func TestGetPullRequestBuildRunJobs_GivenABuildRunLink_WhenFetching_ThenItUsesGhRunViewJSONJobsAndReturnsNormalizedJobs(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"jobs":[{"databaseId":1234,"name":" Test ","status":" completed ","conclusion":" failure ","url":" https://github.com/acme/widgets/actions/runs/42/job/1234 "}]}`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestBuildRunJobs("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "42", "-R", "acme/widgets", "--json", "jobs"})
	expected := []PullRequestBuildRunJob{{DatabaseID: 1234, Name: "Test", Status: "completed", Conclusion: "failure", URL: "https://github.com/acme/widgets/actions/runs/42/job/1234"}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected jobs %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestBuildRunJobLog_GivenAJobID_WhenFetching_ThenItUsesGhRunViewLogForThatJob(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("job logs\n")}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestBuildRunJobLog("acme/widgets", 1234)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "--job=1234", "--log", "--repo=acme/widgets"})
	if actual != "job logs" {
		t.Fatalf("expected logs %q, actual %q", "job logs", actual)
	}
}
