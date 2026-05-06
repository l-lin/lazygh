package tui

import (
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestLatestPullRequestReviews_GivenMultipleReviewsPerAuthor_WhenSelectingLatest_ThenItUsesSubmittedAtAndInputOrderAsTieBreaker(t *testing.T) {
	actual := latestPullRequestReviews([]githubcli.PullRequestReview{
		{Author: &githubcli.PullRequestCommentAuthor{Login: "alice"}, State: "APPROVED", SubmittedAt: "2026-05-05T10:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "bob"}, State: "CHANGES_REQUESTED", SubmittedAt: "2026-05-05T09:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "alice"}, State: "CHANGES_REQUESTED", SubmittedAt: "2026-05-05T11:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "bob"}, State: "APPROVED", SubmittedAt: "2026-05-05T09:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "carol"}, State: "APPROVED", SubmittedAt: "not-a-time"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "carol"}, State: "COMMENTED", SubmittedAt: "still-not-a-time"},
		{State: "APPROVED", SubmittedAt: "2026-05-05T12:00:00Z"},
	})

	if actual["alice"].State != "CHANGES_REQUESTED" {
		t.Fatalf("expected alice latest state %q, actual %q", "CHANGES_REQUESTED", actual["alice"].State)
	}
	if actual["bob"].State != "APPROVED" {
		t.Fatalf("expected bob latest state %q, actual %q", "APPROVED", actual["bob"].State)
	}
	if actual["carol"].State != "COMMENTED" {
		t.Fatalf("expected carol latest state %q, actual %q", "COMMENTED", actual["carol"].State)
	}
	if _, ok := actual[""]; ok {
		t.Fatal("expected reviews without authors to be ignored")
	}
}

func TestApprovedPullRequestReviewerLogins_GivenMixedLatestReviewStates_WhenFiltering_ThenItReturnsSortedApprovers(t *testing.T) {
	actual := approvedPullRequestReviewerLogins([]githubcli.PullRequestReview{
		{Author: &githubcli.PullRequestCommentAuthor{Login: "zoe"}, State: "APPROVED", SubmittedAt: "2026-05-05T08:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "amy"}, State: "APPROVED", SubmittedAt: "2026-05-05T07:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "zoe"}, State: "COMMENTED", SubmittedAt: "2026-05-05T09:00:00Z"},
		{Author: &githubcli.PullRequestCommentAuthor{Login: "bob"}, State: "APPROVED", SubmittedAt: "2026-05-05T10:00:00Z"},
	})

	expected := []string{"amy", "bob"}
	if len(actual) != len(expected) {
		t.Fatalf("expected approvers %v, actual %v", expected, actual)
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("expected approvers %v, actual %v", expected, actual)
		}
	}
}
