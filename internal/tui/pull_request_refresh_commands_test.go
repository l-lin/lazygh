package tui

import (
	"errors"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type given_refreshQueries struct {
	given_progressQueries
	reservations int
	releases     int
}

func (queries *given_refreshQueries) AcquirePullRequestRefresh(string, int) func() {
	queries.reservations++
	return func() { queries.releases++ }
}

func TestRefreshCommands_GivenQueuedDetailAndDiff_WhenTheyCompleteOrFail_ThenTheyReserveBeforeQueueingAndReleaseAfterLoading(t *testing.T) {
	for _, fail := range []bool{false, true} {
		name := "success"
		if fail {
			name = "failure"
		}
		t.Run(name, func(t *testing.T) {
			subject := NewProgram()
			summary := given_refreshSummary()
			runner := &capturingAsyncRunner{}
			subject.asyncRunner = runner
			var loadErr error
			if fail {
				loadErr = errors.New("offline")
			}
			queries := &given_refreshQueries{}
			queries.when_detailProgress = func(func(githubdomain.PullRequestDetail)) (githubdomain.PullRequestDetail, error) {
				if queries.reservations != 2 || queries.releases != 0 {
					t.Fatalf("expected both queued reads reserved during detail, actual %+v", queries)
				}
				return githubdomain.PullRequestDetail{PendingReviewStateKnown: true}, loadErr
			}
			queries.when_diffProgress = func(func(githubdomain.PullRequestDiff)) (githubdomain.PullRequestDiff, error) {
				if queries.reservations != 2 || queries.releases != 1 {
					t.Fatalf("expected diff reservation retained, actual %+v", queries)
				}
				return githubdomain.PullRequestDiff{FileTeamOwnersAttempted: true}, loadErr
			}
			subject.detailQueries = queries
			loadPullRequestDetailCmd{summary: summary}.execute(subject, nil)
			loadPullRequestDiffCmd{summary: summary}.execute(subject, nil)
			if queries.reservations != 2 || queries.releases != 0 || len(runner.runs) != 2 {
				t.Fatalf("expected reservations before queued loads, actual %+v", queries)
			}
			for _, run := range runner.runs {
				run()
			}
			if actual, expected := queries.releases, 2; actual != expected {
				t.Fatalf("expected %d releases, actual %d", expected, actual)
			}
		})
	}
}
