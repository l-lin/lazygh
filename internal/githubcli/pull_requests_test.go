package githubcli

import (
	"reflect"
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
	then_commandIs(t, runner, "gh", []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt,id"})
}

func TestListPullRequests_GivenPullRequestSearchArgumentsWithConfiguredJSON_WhenFetching_ThenItReplacesThemWithTheRequiredJSONFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--json", "title,url", "--state", "open"})

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected no pull requests, actual %+v", actual)
	}
	then_commandIs(t, runner, "gh", []string{"search", "prs", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt,id"})
}

func TestListPullRequests_GivenPullRequestIDs_WhenFetching_ThenItHydratesTheReviewDecision(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`[{"id":"PR_kwDOA","title":"Ship it","number":42,"repository":{"nameWithOwner":"acme/widgets"},"url":"https://github.com/acme/widgets/pull/42","state":"OPEN"}]`)},
		{stdout: []byte(`{"data":{"nodes":[{"id":" PR_kwDOA ","reviewDecision":" APPROVED "}]}}`)},
	}}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"search", "prs", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt,id"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestListReviewMetadataQuery, "-F", "ids[]=PR_kwDOA"}},
	})

	expected := []PullRequest{{
		ID:             "PR_kwDOA",
		Title:          "Ship it",
		Number:         42,
		Repository:     Repository{NameWithOwner: "acme/widgets"},
		URL:            "https://github.com/acme/widgets/pull/42",
		State:          "OPEN",
		ReviewDecision: "APPROVED",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull requests %+v, actual %+v", expected, actual)
	}
}
