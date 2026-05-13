package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetPullRequestBuildRun_GivenABuildRunLink_WhenFetching_ThenItUsesGhRunViewVerbose(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("Run #42\nStatus: completed\nConclusion: failure\n")}
	subject := NewBuildServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestBuildRun("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "42", "-R", "acme/widgets", "--verbose"})
	if actual != "Run #42\nStatus: completed\nConclusion: failure" {
		t.Fatalf("expected run output %q, actual %q", "Run #42\nStatus: completed\nConclusion: failure", actual)
	}
}

func TestGetPullRequestBuildRun_GivenABuildRunAttemptLink_WhenFetching_ThenItPassesTheAttemptNumber(t *testing.T) {
	runner := &fakeRunner{stdout: []byte("Run #42 attempt 3")}
	subject := NewBuildServiceWithRunner(runner)

	_, actualErr := subject.GetPullRequestBuildRun("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42/attempts/3/job/99"})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "42", "-R", "acme/widgets", "--attempt", "3", "--verbose"})
}

func TestGetPullRequestBuildRun_GivenAMissingBuildRunLink_WhenFetching_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewBuildServiceWithRunner(&fakeRunner{})

	_, actualErr := subject.GetPullRequestBuildRun("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI"})

	if !errors.Is(actualErr, ErrMissingPullRequestBuildLink) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestBuildLink, actualErr)
	}
}

func TestGetPullRequestBuildRunJobs_GivenABuildRunLink_WhenFetching_ThenItUsesGhRunViewJSONJobsAndReturnsNormalizedJobs(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"jobs":[{"databaseId":1234,"name":" Test ","status":" completed ","conclusion":" failure ","url":" https://github.com/acme/widgets/actions/runs/42/job/1234 "}]}`)}
	subject := NewBuildServiceWithRunner(runner)

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
	subject := NewBuildServiceWithRunner(runner)

	actual, actualErr := subject.GetPullRequestBuildRunJobLog("acme/widgets", 1234)

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"run", "view", "--job=1234", "--log", "--repo=acme/widgets"})
	if actual != "job logs" {
		t.Fatalf("expected logs %q, actual %q", "job logs", actual)
	}
}

func TestGetPullRequestBuildRunJobLogForCheck_GivenAJobNameMatchingTheCheck_WhenFetching_ThenItLoadsThatJobLog(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"jobs":[{"databaseId":999,"name":"lint","url":"https://github.com/acme/widgets/actions/runs/42/job/999"},{"databaseId":1234,"name":"test","url":"https://github.com/acme/widgets/actions/runs/42/job/1234"}]}`)},
		{stdout: []byte("job logs\n")},
	}}
	subject := NewBuildServiceWithRunner(runner)

	actualJob, actualLog, actualErr := subject.GetPullRequestBuildRunJobLogForCheck("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	expectedJob := PullRequestBuildRunJob{DatabaseID: 1234, Name: "test", URL: "https://github.com/acme/widgets/actions/runs/42/job/1234"}
	if !reflect.DeepEqual(actualJob, expectedJob) {
		t.Fatalf("expected job %+v, actual %+v", expectedJob, actualJob)
	}
	if actualLog != "job logs" {
		t.Fatalf("expected logs %q, actual %q", "job logs", actualLog)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 command calls, actual %d", len(runner.calls))
	}
	if actual := runner.calls[0]; actual.name != "gh" || !reflect.DeepEqual(actual.args, []string{"run", "view", "42", "-R", "acme/widgets", "--json", "jobs"}) {
		t.Fatalf("expected first call gh %v, actual %s %v", []string{"run", "view", "42", "-R", "acme/widgets", "--json", "jobs"}, actual.name, actual.args)
	}
	if actual := runner.calls[1]; actual.name != "gh" || !reflect.DeepEqual(actual.args, []string{"run", "view", "--job=1234", "--log", "--repo=acme/widgets"}) {
		t.Fatalf("expected second call gh %v, actual %s %v", []string{"run", "view", "--job=1234", "--log", "--repo=acme/widgets"}, actual.name, actual.args)
	}
}

func TestGetPullRequestBuildRunJobLogForCheck_GivenOnlyOneJobWithoutANameMatch_WhenFetching_ThenItUsesThatJobLog(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"jobs":[{"databaseId":1234,"name":"lint","url":"https://github.com/acme/widgets/actions/runs/42/job/1234"}]}`)},
		{stdout: []byte("job logs\n")},
	}}
	subject := NewBuildServiceWithRunner(runner)

	actualJob, actualLog, actualErr := subject.GetPullRequestBuildRunJobLogForCheck("acme/widgets", PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	expectedJob := PullRequestBuildRunJob{DatabaseID: 1234, Name: "lint", URL: "https://github.com/acme/widgets/actions/runs/42/job/1234"}
	if !reflect.DeepEqual(actualJob, expectedJob) {
		t.Fatalf("expected job %+v, actual %+v", expectedJob, actualJob)
	}
	if actualLog != "job logs" {
		t.Fatalf("expected logs %q, actual %q", "job logs", actualLog)
	}
}
