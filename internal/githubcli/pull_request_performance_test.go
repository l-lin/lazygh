package githubcli

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestDiffAssembler_GivenIndependentReads_WhenUnifiedDiffIsBlocked_ThenFilesLoadConcurrently(t *testing.T) {
	filesStarted := make(chan struct{})
	assembler := PullRequestDiffAssembler{
		LoadUnifiedDiff: func(string, int) (string, error) {
			select {
			case <-filesStarted:
				return "diff", nil
			case <-time.After(time.Second):
				return "", errors.New("files did not start")
			}
		},
		LoadDiffFiles: func(string, int) ([]PullRequestDiffFile, error) {
			close(filesStarted)
			return []PullRequestDiffFile{{Path: "main.go"}}, nil
		},
	}
	actual, actualErr := assembler.Assemble("acme/widgets", 42)
	then_noError(t, actualErr)
	expected := "diff"
	if actual.UnifiedDiff != expected {
		t.Fatalf("expected %q, actual %+v", expected, actual)
	}
}

func TestDetailAssembler_GivenIndependentEnrichment_WhenCommentsAreBlocked_ThenComparisonAndBuildLinksLoadConcurrently(t *testing.T) {
	comparisonStarted := make(chan struct{})
	buildsStarted := make(chan struct{})
	assembler := PullRequestDetailAssembler{
		LoadBaseDetail: func(string, int) (PullRequestDetail, error) {
			return PullRequestDetail{StatusCheckRollup: []PullRequestStatusCheck{{Name: "CI"}}}, nil
		},
		ListInlineComments: func(string, int) ([]PullRequestInlineComment, error) {
			for _, started := range []chan struct{}{comparisonStarted, buildsStarted} {
				select {
				case <-started:
				case <-time.After(time.Second):
					return nil, errors.New("independent enrichment did not start")
				}
			}
			return nil, nil
		},
		LoadOutOfDateWithBase: func(string, PullRequestDetail) (bool, error) { close(comparisonStarted); return true, nil },
		HydrateBuildLinks: func(string, int, []PullRequestStatusCheck) []PullRequestStatusCheck {
			close(buildsStarted)
			return []PullRequestStatusCheck{{Name: "CI", Link: "build"}}
		},
	}
	actual, actualErr := assembler.Assemble("acme/widgets", 42)
	then_noError(t, actualErr)
	if !actual.OutOfDateWithBase || actual.StatusCheckRollup[0].Link != "build" {
		t.Fatalf("expected both enrichments, actual %+v", actual)
	}
}

func TestCodeownersCache_GivenSharedBaseRevision_WhenFetchingTwoPRsThenChangedBase_ThenItReusesOnlyImmutableRevision(t *testing.T) {
	runner := &fakeRunner{responses: []fakeCommandResponse{
		{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main","baseRefOid":"base-1"}}}}`)},
		{stdout: []byte(`{"data":{"repository":{"dotgithub":{"text":"* @acme/Core"}}}}`)},
		{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main","baseRefOid":"base-1"}}}}`)},
		{stdout: []byte(`{"data":{"repository":{"pullRequest":{"baseRefName":"main","baseRefOid":"base-2"}}}}`)},
		{stdout: []byte(`{"data":{"repository":{"dotgithub":{"text":"* @acme/New"}}}}`)},
	}}
	subject := NewPullRequestDetailServiceWithRunner(runner)
	for _, testCase := range []struct {
		number int
		team   string
	}{{42, "Core"}, {43, "Core"}, {42, "New"}} {
		actual, actualErr := subject.GetPullRequestFileTeamOwners("acme/widgets", testCase.number, []string{"main.go"})
		then_noError(t, actualErr)
		expected := map[string][]string{"main.go": {testCase.team}}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expected %+v, actual %+v", expected, actual)
		}
	}
	if actual, expected := len(runner.calls), 5; actual != expected {
		t.Fatalf("expected %d calls, actual %d", expected, actual)
	}
}

func TestRefreshThreads_GivenTwoQueuedReads_WhenSharingThenReleasingCycle_ThenItFetchesOnceAndNextCycleRefetches(t *testing.T) {
	subject := NewPullRequestDetailServiceWithRunner(nil)
	releaseDetail := subject.AcquirePullRequestRefresh("acme/widgets", 42)
	releaseDiff := subject.AcquirePullRequestRefresh("acme/widgets", 42)
	scope := subject.pullRequestRefresh("acme/widgets", 42)
	var workers sync.WaitGroup
	actual := 0
	var mutex sync.Mutex
	for range 2 {
		workers.Go(func() {
			threads, err := scope.loadThreads(func() ([]PullRequestReviewThread, error) {
				mutex.Lock()
				actual++
				mutex.Unlock()
				return []PullRequestReviewThread{{ID: "thread"}}, nil
			})
			if err != nil || len(threads) != 1 {
				t.Errorf("expected shared threads, actual %+v, %v", threads, err)
			}
		})
	}
	workers.Wait()
	if expected := 1; actual != expected {
		t.Fatalf("expected %d fetch, actual %d", expected, actual)
	}
	releaseDetail()
	if subject.pullRequestRefresh("acme/widgets", 42) != scope {
		t.Fatal("expected scope while diff still owns it")
	}
	releaseDiff()
	releaseNext := subject.AcquirePullRequestRefresh("acme/widgets", 42)
	defer releaseNext()
	nextScope := subject.pullRequestRefresh("acme/widgets", 42)
	if nextScope == scope {
		t.Fatal("expected new scope after completion")
	}
	_, _ = nextScope.loadThreads(func() ([]PullRequestReviewThread, error) { actual++; return nil, nil })
	if expected := 2; actual != expected {
		t.Fatalf("expected %d fetches, actual %d", expected, actual)
	}
}
