package githubcli

import (
	"errors"
	"reflect"
	"testing"
)

func TestGetPendingPullRequestReviewID_GivenViewerPendingReview_WhenListingReviews_ThenItReturnsThePendingReviewID(t *testing.T) {
	runner := &fakeRunner{
		stdout: []byte(`{"data":{"viewer":{"login":"octocat"},"repository":{"pullRequest":{"id":"PR_kwDOAA","reviews":{"nodes":[{"id":"PRR_1","state":"COMMENTED","author":{"login":"octocat"}},{"id":"PRR_pending","state":"PENDING","author":{"login":"octocat"}},{"id":"PRR_other","state":"PENDING","author":{"login":"someone-else"}}]}}}}}`),
	}
	subject := NewClientWithRunner(runner)

	actual, found, actualErr := subject.GetPendingPullRequestReviewID("acme/widgets", 42)

	then_noError(t, actualErr)
	if !found {
		t.Fatal("expected to find a pending review")
	}
	if actual != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", actual)
	}
	then_commandIs(t, runner, "gh", []string{"api", "graphql", "-f", "query=" + pendingPullRequestReviewQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"})
}

func TestStartPendingPullRequestReview_GivenExistingViewerPendingReview_WhenStarting_ThenItReusesTheExistingReviewID(t *testing.T) {
	runner := &fakeRunner{
		stdout: []byte(`{"data":{"viewer":{"login":"octocat"},"repository":{"pullRequest":{"id":"PR_kwDOAA","reviews":{"nodes":[{"id":"PRR_pending","state":"PENDING","author":{"login":"octocat"}}]}}}}}`),
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.StartPendingPullRequestReview("acme/widgets", 42)

	then_noError(t, actualErr)
	if actual != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", actual)
	}
	then_commandsAre(t, runner, []fakeCommandCall{{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pendingPullRequestReviewQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}}})
}

func TestStartPendingPullRequestReview_GivenNoViewerPendingReview_WhenStarting_ThenItCreatesOneViaGraphQL(t *testing.T) {
	runner := &fakeRunner{
		responses: []fakeCommandResponse{
			{stdout: []byte(`{"data":{"viewer":{"login":"octocat"},"repository":{"pullRequest":{"id":"PR_kwDOAA","reviews":{"nodes":[{"id":"PRR_1","state":"COMMENTED","author":{"login":"octocat"}}]}}}}}`)},
			{stdout: []byte(`{"data":{"addPullRequestReview":{"pullRequestReview":{"id":"PRR_new"}}}}`)},
		},
	}
	subject := NewClientWithRunner(runner)

	actual, actualErr := subject.StartPendingPullRequestReview("acme/widgets", 42)

	then_noError(t, actualErr)
	if actual != "PRR_new" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_new", actual)
	}
	then_commandsAre(t, runner, []fakeCommandCall{
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + pendingPullRequestReviewQuery, "-F", "owner=acme", "-F", "name=widgets", "-F", "number=42"}},
		{name: "gh", args: []string{"api", "graphql", "-f", "query=" + addPendingPullRequestReviewMutation, "-F", "pullRequestId=PR_kwDOAA"}},
	})
}

func TestStartPendingPullRequestReview_GivenMissingPullRequestIdentity_WhenStarting_ThenItReturnsAValidationError(t *testing.T) {
	subject := NewClientWithRunner(&fakeRunner{})

	_, actualErr := subject.StartPendingPullRequestReview(" ", 0)

	if !errors.Is(actualErr, ErrMissingPullRequestIdentity) {
		t.Fatalf("expected error %v, actual %v", ErrMissingPullRequestIdentity, actualErr)
	}
}

func TestGetPendingPullRequestReviewID_GivenWhitespacePayload_WhenListingReviews_ThenItNormalizesTheResponse(t *testing.T) {
	runner := &fakeRunner{
		stdout: []byte(`{"data":{"viewer":{"login":"  octocat  "},"repository":{"pullRequest":{"id":"  PR_kwDOAA  ","reviews":{"nodes":[{"id":"  PRR_pending  ","state":"  PENDING  ","author":{"login":"  octocat  "}}]}}}}}`),
	}
	subject := NewClientWithRunner(runner)

	actual, found, actualErr := subject.GetPendingPullRequestReviewID("acme/widgets", 42)

	then_noError(t, actualErr)
	if !found {
		t.Fatal("expected to find a pending review")
	}
	if actual != "PRR_pending" {
		t.Fatalf("expected normalized pending review id %q, actual %q", "PRR_pending", actual)
	}
}

func TestSplitRepositoryOwnerAndName_GivenRepositoryNameWithOwner_WhenSplitting_ThenItReturnsTheOwnerAndName(t *testing.T) {
	actualOwner, actualName, actualErr := splitRepositoryOwnerAndName("acme/widgets")

	then_noError(t, actualErr)
	if !reflect.DeepEqual([]string{actualOwner, actualName}, []string{"acme", "widgets"}) {
		t.Fatalf("expected owner and name %v, actual %v", []string{"acme", "widgets"}, []string{actualOwner, actualName})
	}
}
