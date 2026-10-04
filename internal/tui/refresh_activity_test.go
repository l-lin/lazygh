package tui

import (
	"errors"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestRefreshActivity_GivenQueuedWork_WhenItRuns_ThenSeparatesQueueWaitFromExecutionAndRetainsOwnership(t *testing.T) {
	for _, fails := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fails], func(t *testing.T) {
			// GIVEN
			currentTime := time.Date(2026, 10, 4, 12, 0, 0, 0, time.Local)
			now := func() time.Time { return currentTime }
			var queuedRun func()
			var actual []githubdomain.CommandActivity
			report := func(msg Msg) { actual = append(actual, msg.(MsgCommandActivity).Activity) }
			run := func() error {
				currentTime = currentTime.Add(2 * time.Second)
				if fails {
					return errors.New("read failed")
				}
				return nil
			}
			// WHEN
			runObservedCommandAsync(func(task func()) { queuedRun = task }, report, now, "refresh diff acme/widgets#42", "scheduled batch=7", run)
			if len(actual) != 0 {
				t.Fatalf("expected no execution before queue dispatch, actual %+v", actual)
			}
			currentTime = currentTime.Add(3 * time.Second)
			queuedRun()
			// THEN
			expected := githubdomain.CommandSucceeded
			if fails {
				expected = githubdomain.CommandFailed
			}
			if len(actual) != 2 || actual[1].Outcome != expected || actual[0].Outcome != githubdomain.CommandRunning {
				t.Fatalf("expected paired lifecycle with outcome %v, actual %+v", expected, actual)
			}
			if actual[1].QueueWait != 3*time.Second || actual[1].FinishedAt.Sub(actual[1].StartedAt) != 2*time.Second || actual[1].Owner != "scheduled batch=7" || actual[1].ID != actual[0].ID {
				t.Fatalf("expected separate queue and execution timings with stable ownership, actual %+v", actual)
			}
		})
	}
}

func TestRefreshActivity_GivenSaturatedReadQueue_WhenWorkerAndBackendDispatchAreDelayed_ThenMeasuresBothAsQueueWait(t *testing.T) {
	// GIVEN
	backend := &capturingAsyncRunner{}
	queue := newPullRequestAsyncRunner(backend)
	for range 2 {
		queue.goPullRequest("other", func() {})
	}
	currentTime := time.Now()
	var actual []githubdomain.CommandActivity
	runObservedCommandAsync(func(run func()) { queue.goPullRequest("target", run) }, func(msg Msg) { actual = append(actual, msg.(MsgCommandActivity).Activity) }, func() time.Time { return currentTime }, "refresh detail acme/widgets#42", "background", func() error { currentTime = currentTime.Add(time.Second); return nil })
	if len(backend.runs) != 2 || len(actual) != 0 {
		t.Fatal("expected observed work to remain queued")
	}
	// WHEN
	currentTime = currentTime.Add(3 * time.Second)
	backend.runs[0]()
	currentTime = currentTime.Add(2 * time.Second)
	backend.runs[2]()
	// THEN
	expected := 5 * time.Second
	if len(actual) != 2 || actual[1].QueueWait != expected || actual[1].FinishedAt.Sub(actual[1].StartedAt) != time.Second {
		t.Fatalf("expected %s queue wait and 1s execution, actual %+v", expected, actual)
	}
}

func TestRefreshActivity_GivenSynchronousBackend_WhenRunningTwice_ThenEachExecutionHasAnIndependentIdentity(t *testing.T) {
	// GIVEN
	var actual []githubdomain.CommandActivity
	report := func(msg Msg) { actual = append(actual, msg.(MsgCommandActivity).Activity) }
	// WHEN
	for range 2 {
		runObservedCommandAsync(nil, report, time.Now, "refresh detail acme/widgets#42", "manual", func() error { return nil })
	}
	// THEN
	if len(actual) != 4 || actual[0].ID == actual[2].ID {
		t.Fatalf("expected independent execution IDs, actual %+v", actual)
	}
}

func TestPullRequestCommands_GivenRefreshReads_WhenTheyComplete_ThenRecordsTargetOutcomeAndSubmissionOwnership(t *testing.T) {
	summary := given_refreshSummary()
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	for _, scenario := range []struct {
		name            string
		command         Cmd
		expectedCommand string
		expectedOwner   string
		fails           bool
	}{
		{"manual detail", loadPullRequestDetailCmd{summary: summary}, "refresh detail " + key, "manual", false},
		{"scheduled diff", loadPullRequestDiffCmd{summary: summary, ScheduledRefreshBatchID: 7}, "refresh diff " + key, "scheduled batch=7", false},
		{"failed probe", scheduledPullRequestProbeCmd{summaries: []githubdomain.PullRequest{summary}, batchID: 7}, "refresh probe " + key, "scheduled batch=7", true},
		{"scheduled list", loadPullRequestsCmd{ScheduledRefreshBatchID: 7, Source: pullRequestLoadSourceScheduled}, "refresh list tab=0", "scheduled batch=7", false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			// GIVEN
			subject := NewProgram()
			subject.scheduledPullRequestRefreshBatches = map[uint64]scheduledPullRequestRefreshBatch{7: {pending: 10, planning: true}}
			queries := &fakePullRequestDetailLoader{}
			if scenario.fails {
				queries.detailErrors = map[string]error{key: errors.New("remote failure")}
			}
			subject.detailQueries = queries
			subject.pullRequestListQueries = &fakePullRequestListQueries{}
			subject.markManualPullRequestDetailRefresh(summary)
			backend := &capturingAsyncRunner{}
			subject.asyncRunner = newPullRequestAsyncRunner(backend)
			subject.uiUpdater = immediateUIUpdater{}
			gui := given_headlessGui(t)
			defer gui.Close()
			// WHEN
			scenario.command.execute(subject, gui)
			subject.consumeManualPullRequestDetailRefresh(key)
			if len(backend.runs) == 0 {
				t.Fatal("expected a queued read")
			}
			backend.runs[0]()
			// THEN
			var actual githubdomain.CommandActivity
			for _, entry := range subject.entriesSnapshot() {
				if entry.command == scenario.expectedCommand {
					actual = entry.activity
				}
			}
			expected := githubdomain.CommandSucceeded
			if scenario.fails {
				expected = githubdomain.CommandFailed
			}
			if actual.Outcome != expected || actual.Owner != scenario.expectedOwner || actual.ID == "" || actual.QueueWait < 0 || actual.FinishedAt.Before(actual.StartedAt) {
				t.Fatalf("expected %q with owner %q and outcome %v, actual %+v", scenario.expectedCommand, scenario.expectedOwner, expected, actual)
			}
		})
	}
}

func TestRefreshCommandOwner_GivenRefreshSources_WhenSnapshottingOwnership_ThenDistinguishesScheduledManualVisibleAndBackground(t *testing.T) {
	// GIVEN
	subject := NewProgramWithModel(given_model())
	for _, scenario := range []struct {
		batchID  uint64
		manual   bool
		expected string
	}{{7, false, "scheduled batch=7"}, {7, true, "scheduled batch=7"}, {0, true, "manual"}, {0, false, "background"}} {
		// WHEN
		actual := subject.refreshCommandOwner(scenario.batchID, scenario.manual, "other/repo#1")
		// THEN
		if actual != scenario.expected {
			t.Fatalf("expected %q, actual %q", scenario.expected, actual)
		}
	}
	actual := subject.refreshCommandOwner(0, false, subject.activePullRequestLoadKey())
	expected := "interactive"
	if actual != expected {
		t.Fatalf("expected %q, actual %q", expected, actual)
	}
}
