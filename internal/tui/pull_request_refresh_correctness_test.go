package tui

import (
	"errors"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func given_refreshSummary() githubdomain.PullRequest {
	return githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-10-04T12:00:00Z"}
}

func TestScheduledRefresh_GivenUnreadCurrentCaches_WhenSelectingTargets_ThenItSkipsFetching(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	subject.pullRequestDetailCache[key] = pullRequestDetailResult{sourceUpdatedAt: summary.UpdatedAt}
	subject.pullRequestDiffCache[key] = pullRequestDiffResult{sourceUpdatedAt: summary.UpdatedAt, fileTeamOwnersAttempted: true}
	actual := subject.pullRequestsNeedingScheduledRefresh([]githubdomain.PullRequest{summary})
	expected := 0
	if len(actual) != expected {
		t.Fatalf("expected %d targets, actual %v", expected, actual)
	}
}

func TestScheduledRefresh_GivenSeenMissingCaches_WhenSelectingTargets_ThenItFetches(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen("acme/widgets#42", summary.UpdatedAt)
	})
	actual := subject.pullRequestsNeedingScheduledRefresh([]githubdomain.PullRequest{summary})
	expected := 1
	if len(actual) != expected {
		t.Fatalf("expected %d targets, actual %v", expected, actual)
	}
}

func TestRefresh_GivenInitialFailures_WhenBothLoadsSucceed_ThenItClearsErrorsAndRetryState(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	subject.applyPullRequestDetailLoaded(MsgPullRequestDetailLoaded{Summary: summary, Err: errors.New("offline")})
	subject.applyPullRequestDiffLoaded(MsgPullRequestDiffLoaded{Summary: summary, Err: errors.New("offline")})
	actualDetail := subject.pullRequestDetailCache[key]
	actualDiff := subject.pullRequestDiffCache[key]
	if actualDetail.sourceUpdatedAt != "" || actualDiff.sourceUpdatedAt != "" || actualDetail.retry.ready(subject.currentTime()) || actualDiff.retry.ready(subject.currentTime()) {
		t.Fatalf("expected failed initial loads to have no fetched version and a retry delay, actual detail=%+v diff=%+v", actualDetail, actualDiff)
	}
	subject.applyPullRequestDetailLoaded(MsgPullRequestDetailLoaded{Summary: summary, Detail: githubdomain.PullRequestDetail{Body: "fresh"}})
	subject.applyPullRequestDiffLoaded(MsgPullRequestDiffLoaded{Summary: summary})
	actualDetail = subject.pullRequestDetailCache[key]
	actualDiff = subject.pullRequestDiffCache[key]
	expected := summary.UpdatedAt
	if actualDetail.sourceUpdatedAt != expected || actualDiff.sourceUpdatedAt != expected || actualDetail.err != nil || actualDiff.err != nil || actualDetail.retry.failureCount != 0 || actualDiff.retry.failureCount != 0 {
		t.Fatalf("expected successful version %s and cleared retries, actual detail=%+v diff=%+v", expected, actualDetail, actualDiff)
	}
}

func TestRefresh_GivenCachedContent_WhenBothLoadsFail_ThenItKeepsSuccessfulVersionsAndRetriesAfterBackoff(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	subject.timingState.now = func() time.Time { return now }
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	expected := "2026-10-04T11:00:00Z"
	subject.pullRequestDetailCache[key] = pullRequestDetailResult{detail: githubdomain.PullRequestDetail{Body: "cached"}, sourceUpdatedAt: expected}
	subject.pullRequestDiffCache[key] = pullRequestDiffResult{sourceUpdatedAt: expected}
	subject.applyPullRequestDetailLoaded(MsgPullRequestDetailLoaded{Summary: summary, Err: errors.New("offline")})
	subject.applyPullRequestDiffLoaded(MsgPullRequestDiffLoaded{Summary: summary, Err: errors.New("offline")})
	actualDetail := subject.pullRequestDetailCache[key]
	actualDiff := subject.pullRequestDiffCache[key]
	if actualDetail.sourceUpdatedAt != expected || actualDiff.sourceUpdatedAt != expected || !actualDetail.needsRefresh || !actualDiff.needsRefresh || actualDetail.detail.Body != "cached" {
		t.Fatalf("expected retained stale content at %s, actual detail=%+v diff=%+v", expected, actualDetail, actualDiff)
	}
	actual := subject.pullRequestsNeedingScheduledRefresh([]githubdomain.PullRequest{summary})
	if len(actual) != 0 {
		t.Fatalf("expected backoff to suppress retry, actual %v", actual)
	}
	now = now.Add(31 * time.Second)
	actual = subject.pullRequestsNeedingScheduledRefresh([]githubdomain.PullRequest{summary})
	if len(actual) != 1 {
		t.Fatalf("expected retry after backoff, actual %v", actual)
	}
}
