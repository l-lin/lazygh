package githubcli

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestRefreshSharing_GivenDetailAndDiffReservations_WhenEitherLoadsFirst_ThenBothUseOnePaginatedThreadResult(t *testing.T) {
	for _, detailFirst := range []bool{true, false} {
		name := "diff-first"
		if detailFirst {
			name = "detail-first"
		}
		t.Run(name, func(t *testing.T) {
			threadPage := `{"data":{"repository":{"pullRequest":{"id":"PR_42","reviewThreads":{"pageInfo":{"hasNextPage":true,"endCursor":"next"},"nodes":[{"id":"thread-1","path":"main.go","comments":{"pageInfo":{"hasNextPage":true,"endCursor":"reply"},"nodes":[{"id":"comment-1","body":"first"}]}}]}}}}}`
			lastThreadPage := `{"data":{"repository":{"pullRequest":{"id":"PR_42","reviewThreads":{"nodes":[{"id":"thread-2","path":"main.go"}]}}}}}`
			runner := &fakeRunner{responses: []fakeCommandResponse{
				{commandContains: "pr view", stdout: []byte(`{"number":42,"body":"description"}`)},
				{commandContains: "/comments?", stdout: []byte(`[]`)},
				{commandContains: "Accept: application/vnd.github.v3.diff", stdout: []byte("diff")},
				{commandContains: "/files?", stdout: []byte(`[[{"filename":"main.go"}]]`)},
				{commandContains: "includeThreads=false", stdout: []byte(`{"data":{"repository":{"pullRequest":{"id":"PR_42"}}}}`)},
				{commandContains: "threadID=thread-1", stdout: []byte(`{"data":{"node":{"comments":{"nodes":[{"id":"comment-2","body":"reply"}]}}}}`)},
				{commandContains: "api graphql", stdout: []byte(threadPage)},
				{commandContains: "api graphql", stdout: []byte(lastThreadPage)},
			}}
			subject := NewPullRequestDetailServiceWithRunner(runner)
			releaseDetail := subject.AcquirePullRequestRefresh("acme/widgets", 42)
			defer releaseDetail()
			releaseDiff := subject.AcquirePullRequestRefresh("acme/widgets", 42)
			defer releaseDiff()
			var actualDetail PullRequestDetail
			var actualDiff PullRequestDiff
			var detailErr, diffErr error
			if detailFirst {
				actualDetail, detailErr = subject.GetPullRequestDetail("acme/widgets", 42)
				actualDiff, diffErr = subject.GetPullRequestDiff("acme/widgets", 42)
			} else {
				actualDiff, diffErr = subject.GetPullRequestDiff("acme/widgets", 42)
				actualDetail, detailErr = subject.GetPullRequestDetail("acme/widgets", 42)
			}
			then_noError(t, detailErr)
			then_noError(t, diffErr)
			expected := actualDetail.InlineCommentThreads
			actual := actualDiff.Threads
			if len(actual) != 2 || len(actual[0].Comments) != 2 || !reflect.DeepEqual(actual, expected) {
				t.Fatalf("expected complete shared threads %+v, actual %+v", expected, actual)
			}
			threadReads := 0
			for _, call := range runner.calls {
				command := strings.Join(call.args, " ")
				if strings.Contains(command, "reviewThreads(first:") && !strings.Contains(command, "includeThreads=false") {
					threadReads++
				}
			}
			if actual, expected := threadReads, 2; actual != expected {
				t.Fatalf("expected %d paginated thread reads, actual %d: %+v", expected, actual, runner.calls)
			}
			actualDiff.Threads[0].Comments[0].Body = "changed"
			if actualDetail.InlineCommentThreads[0].Comments[0].Body != "first" {
				t.Fatal("expected independently owned shared result slices")
			}
		})
	}
}

func TestCodeownersCache_GivenMissingOrFailedBlob_WhenFetchingAgain_ThenItCachesAbsenceButRetriesFailure(t *testing.T) {
	for _, failFirst := range []bool{false, true} {
		name := "missing"
		if failFirst {
			name = "failed"
		}
		t.Run(name, func(t *testing.T) {
			loadErr := errors.New("offline")
			response := fakeCommandResponse{stdout: []byte(`{"data":{"repository":{"dotgithub":null,"root":null,"docs":null}}}`)}
			if failFirst {
				response.err = loadErr
			}
			runner := &fakeRunner{responses: []fakeCommandResponse{response, {stdout: []byte(`{"data":{"repository":{"dotgithub":{"text":"* @acme/Core"}}}}`)}}}
			subject := NewPullRequestDetailServiceWithRunner(runner)
			_, _, actualErr := subject.cachedPullRequestCodeownersBlob("acme", "widgets", "base-1")
			if failFirst && !errors.Is(actualErr, loadErr) {
				t.Fatalf("expected load error, actual %v", actualErr)
			}
			_, actualFound, actualErr := subject.cachedPullRequestCodeownersBlob("acme", "widgets", "base-1")
			then_noError(t, actualErr)
			expectedCalls := 1
			if failFirst {
				expectedCalls = 2
			}
			if actualFound != failFirst || len(runner.calls) != expectedCalls {
				t.Fatalf("expected found=%v and %d calls, actual %v and %d", failFirst, expectedCalls, actualFound, len(runner.calls))
			}
		})
	}
}

func TestCodeownersCache_GivenConcurrentReadsAndAnotherRepository_WhenFetching_ThenItCoalescesOnlyMatchingRepositoryAndRevision(t *testing.T) {
	runner := &fakeRunner{stdout: []byte(`{"data":{"repository":{"dotgithub":{"text":"* @acme/Core"}}}}`)}
	subject := NewPullRequestDetailServiceWithRunner(runner)
	var workers sync.WaitGroup
	for range 10 {
		workers.Go(func() {
			_, found, err := subject.cachedPullRequestCodeownersBlob("acme", "widgets", "base-1")
			if err != nil || !found {
				t.Errorf("expected blob, actual %v, %v", found, err)
			}
		})
	}
	workers.Wait()
	_, _, actualErr := subject.cachedPullRequestCodeownersBlob("acme", "other", "base-1")
	then_noError(t, actualErr)
	if actual, expected := len(runner.calls), 2; actual != expected {
		t.Fatalf("expected %d calls, actual %d", expected, actual)
	}
}

func TestRefreshThreads_GivenFailedCycle_WhenStartingAnotherCycle_ThenItRetries(t *testing.T) {
	subject := NewPullRequestDetailServiceWithRunner(nil)
	release := subject.AcquirePullRequestRefresh("acme/widgets", 42)
	loadErr := errors.New("offline")
	_, actualErr := subject.pullRequestRefresh("acme/widgets", 42).loadThreads(func() ([]PullRequestReviewThread, error) { return nil, loadErr })
	if !errors.Is(actualErr, loadErr) {
		t.Fatalf("expected load failure, actual %v", actualErr)
	}
	release()
	release = subject.AcquirePullRequestRefresh("acme/widgets", 42)
	defer release()
	actual, actualErr := subject.pullRequestRefresh("acme/widgets", 42).loadThreads(func() ([]PullRequestReviewThread, error) { return []PullRequestReviewThread{{ID: "fresh"}}, nil })
	then_noError(t, actualErr)
	if len(actual) != 1 || actual[0].ID != "fresh" {
		t.Fatalf("expected fresh threads, actual %+v", actual)
	}
}
