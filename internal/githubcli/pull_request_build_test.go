package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetPullRequestBuildInfo_GivenMatchingBuildLink_WhenFetching_ThenItReturnsTheNormalizedBuildInfo(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[
		{"bucket":" fail ","completedAt":" 2026-04-18T13:04:00Z ","description":" widget smoke test timed out ","event":" pull_request ","link":" https://github.com/acme/widgets/actions/runs/42 ","name":" test ","startedAt":" 2026-04-18T13:00:00Z ","state":" FAILURE ","workflow":" CI "}
	]`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.GetPullRequestBuildInfo("acme/widgets", 42, PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"pr", "checks", "42", "-R", "acme/widgets", "--json", pullRequestBuildJSONFields})
	expected := PullRequestBuildInfo{
		Bucket:      "fail",
		CompletedAt: "2026-04-18T13:04:00Z",
		Description: "widget smoke test timed out",
		Event:       "pull_request",
		Link:        "https://github.com/acme/widgets/actions/runs/42",
		Name:        "test",
		StartedAt:   "2026-04-18T13:00:00Z",
		State:       "FAILURE",
		Workflow:    "CI",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected build info %+v, actual %+v", expected, actual)
	}
}

func TestGetPullRequestBuildInfo_GivenNoMatchingBuild_WhenFetching_ThenItReturnsNotFound(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[{"name":"lint","workflow":"CI","link":"https://github.com/acme/widgets/actions/runs/1"}]`)}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.GetPullRequestBuildInfo("acme/widgets", 42, PullRequestStatusCheck{Name: "test", WorkflowName: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"})

	if !errors.Is(actualErr, ErrPullRequestBuildInfoNotFound) {
		t.Fatalf("expected error %v, actual %v", ErrPullRequestBuildInfoNotFound, actualErr)
	}
}
