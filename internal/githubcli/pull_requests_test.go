package githubcli

import (
	"testing"
)

func TestListPullRequests_GivenPullRequestSearchArgumentsWithoutJSON_WhenFetching_ThenItAppendsTheRequiredJSONFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"pr", "list", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected no pull requests, actual %+v", actual)
	}
	then_commandIs(t, runner, "gh", []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"})
}

func TestListPullRequests_GivenPullRequestSearchArgumentsWithConfiguredJSON_WhenFetching_ThenItReplacesThemWithTheRequiredJSONFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--json", "title,url", "--state", "open"})

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected no pull requests, actual %+v", actual)
	}
	then_commandIs(t, runner, "gh", []string{"search", "prs", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"})
}
