package githubcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestPullRequestListReviewMetadataQuery_GivenTheStaticGraphQLDocument_WhenReadingIt_ThenItKeepsBalancedSelectionSets(t *testing.T) {
	actualOpenCount := strings.Count(pullRequestListReviewMetadataQuery, "{")
	actualCloseCount := strings.Count(pullRequestListReviewMetadataQuery, "}")

	if actualOpenCount != actualCloseCount {
		t.Fatalf("expected balanced GraphQL braces, actual open=%d close=%d in %q", actualOpenCount, actualCloseCount, pullRequestListReviewMetadataQuery)
	}
}

func TestListPullRequests_GivenPullRequestSearchArgumentsWithoutJSON_WhenFetching_ThenItAppendsTheRequiredJSONFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewPullRequestListServiceWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"pr", "list", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	if len(actual) != 0 {
		t.Fatalf("expected no pull requests, actual %+v", actual)
	}
	then_commandIs(t, runner, "gh", []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt,id"})
}

func TestListPullRequests_GivenPullRequestSearchArgumentsWithConfiguredJSON_WhenFetching_ThenItReplacesThemWithTheRequiredJSONFields(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`[]`)}
	subject := NewPullRequestListServiceWithRunner(runner)

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
	subject := NewPullRequestListServiceWithRunner(runner)

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

func TestListPullRequests_GivenPullRequestIDs_WhenFetching_ThenItHydratesTheRequestedReviewTeamsAndMergeCheckMetadata(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`[{"id":"PR_kwDOA","title":"Need teams","number":42,"repository":{"nameWithOwner":"acme/widgets"},"state":"OPEN"}]`)},
		{stdout: []byte(`{"data":{"nodes":[{"id":"PR_kwDOA","reviewDecision":"REVIEW_REQUIRED","mergeable":"MERGEABLE","mergeStateStatus":"CLEAN","reviewRequests":{"nodes":[{"requestedReviewer":{"__typename":"User","login":"reviewer-one"}},{"requestedReviewer":{"__typename":"Team","name":"VIBE","slug":"vibe","organization":{"login":"acme"}}},{"requestedReviewer":{"__typename":"Team","name":"P3C","slug":"p3c","organization":{"login":"acme"}}},{"requestedReviewer":{"__typename":"Team","name":"FYP","slug":"fyp","organization":{"login":"acme"}}}]},"headRefStatusCheckRollup":{"nodes":[{"commit":{"statusCheckRollup":{"state":"SUCCESS"}}}]}}]}}`)},
	}}
	subject := NewPullRequestListServiceWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	expected := []PullRequest{{
		ID:                     "PR_kwDOA",
		Title:                  "Need teams",
		Number:                 42,
		Repository:             Repository{NameWithOwner: "acme/widgets"},
		State:                  "OPEN",
		ReviewDecision:         "REVIEW_REQUIRED",
		Mergeable:              "MERGEABLE",
		MergeStateStatus:       "CLEAN",
		StatusCheckRollupState: "SUCCESS",
		ReviewRequests: []PullRequestReviewRequest{
			{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-one"}},
			{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "VIBE", Slug: "vibe", Organization: &PullRequestReviewRequestOrganization{Login: "acme"}}},
			{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "P3C", Slug: "p3c", Organization: &PullRequestReviewRequestOrganization{Login: "acme"}}},
			{RequestedReviewer: PullRequestRequestedReviewer{TypeName: "Team", Name: "FYP", Slug: "fyp", Organization: &PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull requests %+v, actual %+v", expected, actual)
	}
}

func TestListPullRequests_GivenPullRequestIDs_WhenFetching_ThenItHydratesTheAutoMergeState(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`[{"id":"PR_kwDOA","title":"Queued merge","number":42,"repository":{"nameWithOwner":"acme/widgets"},"state":"OPEN"}]`)},
		{stdout: []byte(`{"data":{"nodes":[{"id":"PR_kwDOA","autoMergeRequest":{"enabledAt":"2026-05-20T10:00:00Z"}}]}}`)},
	}}
	subject := NewPullRequestListServiceWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	expected := []PullRequest{{
		ID:               "PR_kwDOA",
		Title:            "Queued merge",
		Number:           42,
		Repository:       Repository{NameWithOwner: "acme/widgets"},
		State:            "OPEN",
		AutoMergeRequest: &PullRequestAutoMergeRequest{EnabledAt: "2026-05-20T10:00:00Z"},
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull requests %+v, actual %+v", expected, actual)
	}
}

func TestListPullRequests_GivenReviewMetadataHydrationFailure_WhenFetching_ThenItReturnsTheSearchResultsWithoutFailing(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`[{"id":"PR_kwDOA","title":"Ship it","number":42,"repository":{"nameWithOwner":"acme/widgets"},"url":"https://github.com/acme/widgets/pull/42","state":"OPEN"}]`)},
		{stderr: []byte("gh: HTTP 502"), err: errors.New("exit status 1")},
	}}
	subject := NewPullRequestListServiceWithRunner(runner)

	actual, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"search", "prs", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt,id"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pullRequestListReviewMetadataQuery, "-F", "ids[]=PR_kwDOA"}},
	})

	expected := []PullRequest{{
		ID:         "PR_kwDOA",
		Title:      "Ship it",
		Number:     42,
		Repository: Repository{NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		State:      "OPEN",
	}}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected pull requests %+v, actual %+v", expected, actual)
	}
}

func TestListPullRequests_GivenInvalidReviewMetadataResponse_WhenFetching_ThenItReturnsTheMetadataError(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`[{"id":"PR_kwDOA","title":"Ship it","number":42,"repository":{"nameWithOwner":"acme/widgets"},"state":"OPEN"}]`)},
		{stdout: []byte(`{"data":`)},
	}}
	subject := NewPullRequestListServiceWithRunner(runner)

	_, actualErr := subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--state", "open"})

	if !errors.Is(actualErr, ErrInvalidPullRequestReviewMetadataResponse) {
		t.Fatalf("expected error %v, actual %v", ErrInvalidPullRequestReviewMetadataResponse, actualErr)
	}
}

func TestListPullRequests_GivenMorePullRequestIDsThanOneBatch_WhenFetching_ThenItBatchesTheMetadataGraphQLRequests(t *testing.T) {
	searchResults := make([]PullRequest, 0, pullRequestListReviewMetadataBatchSize+1)
	for index := range pullRequestListReviewMetadataBatchSize + 1 {
		searchResults = append(searchResults, PullRequest{
			ID:         fmt.Sprintf("PR_%03d", index),
			Title:      fmt.Sprintf("Pull Request %03d", index),
			Number:     index,
			Repository: Repository{NameWithOwner: "acme/widgets"},
			State:      "OPEN",
		})
	}
	searchResultJSON, actualErr := json.Marshal(searchResults)
	then_noError(t, actualErr)

	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: searchResultJSON},
		{stdout: []byte(`{"data":{"nodes":[]}}`)},
		{stdout: []byte(`{"data":{"nodes":[]}}`)},
	}}
	subject := NewPullRequestListServiceWithRunner(runner)

	_, actualErr = subject.ListPullRequests([]string{"search", "prs", "--author", "@me", "--state", "open"})

	then_noError(t, actualErr)
	if len(runner.calls) != 3 {
		t.Fatalf("expected 3 command calls, actual %d", len(runner.calls))
	}
	if actual := pullRequestListReviewMetadataArgumentCount(runner.calls[1].args); actual != pullRequestListReviewMetadataBatchSize {
		t.Fatalf("expected first metadata batch size %d, actual %d", pullRequestListReviewMetadataBatchSize, actual)
	}
	if actual := pullRequestListReviewMetadataArgumentCount(runner.calls[2].args); actual != 1 {
		t.Fatalf("expected second metadata batch size %d, actual %d", 1, actual)
	}
}

func pullRequestListReviewMetadataArgumentCount(arguments []string) int {
	count := 0
	for _, argument := range arguments {
		if strings.HasPrefix(argument, "ids[]=") {
			count++
		}
	}
	return count
}
