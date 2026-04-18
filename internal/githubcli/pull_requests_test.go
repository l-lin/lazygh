package githubcli

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestListMyPullRequests_GivenValidGhResponse_WhenFetching_ThenReturnsPullRequests(t *testing.T) {
	runner := &fakeRunner{
		stdout: []byte(`[{"title":"fix(P3C-6986): exclude dependencies bump PRs + bump GHA","number":422,"repository":{"name":"patient-account","nameWithOwner":"doctolib/patient-account"},"url":"https://github.com/doctolib/patient-account/pull/422","body":"No need to trigger Claude review for PRs that only bump dependencies.","state":"open","isDraft":false,"updatedAt":"2026-04-17T10:39:35Z"}]`),
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListMyPullRequests()

	then_noError(t, actualErr)
	then_commandIs(t, runner, "gh", []string{"search", "prs", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"})

	expected := []PullRequest{{
		Title:  "fix(P3C-6986): exclude dependencies bump PRs + bump GHA",
		Number: 422,
		Repository: Repository{
			Name:          "patient-account",
			NameWithOwner: "doctolib/patient-account",
		},
		URL:       "https://github.com/doctolib/patient-account/pull/422",
		Body:      "No need to trigger Claude review for PRs that only bump dependencies.",
		State:     "open",
		IsDraft:   false,
		UpdatedAt: "2026-04-17T10:39:35Z",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull requests %+v, actual %+v", expected, actual)
	}
}

func TestListMyPullRequests_GivenEmptyGhResponse_WhenFetching_ThenReturnsNoPullRequests(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListMyPullRequests()

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected 0 pull requests, actual %d", len(actual))
	}
}

func TestListMyPullRequests_GivenInvalidJSON_WhenFetching_ThenReturnsAnInvalidResponseError(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[{"title":`)}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.ListMyPullRequests()

	if !errors.Is(actualErr, ErrInvalidPullRequestResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestResponse, actualErr)
	}
}

func TestListMyPullRequests_GivenCommandFailure_WhenFetching_ThenReturnsTheSearchError(t *testing.T) {
	runner := &fakeRunner{
		stderr: []byte("boom"),
		err:    errors.New("exit status 1"),
	}
	subject := NewClientWithRunner(runner)

	_, actualErr := subject.ListMyPullRequests()

	if actualErr == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(actualErr.Error(), "gh search prs") {
		t.Fatalf("expected error to mention %q, actual %v", "gh search prs", actualErr)
	}
}
