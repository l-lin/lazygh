package tui

import (
	"errors"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestScheduledRefresh_GivenAnActiveReviewOutsideSearches_WhenTheScheduleIsDue_ThenItProbesTheReview(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	subject.runtimeConfig.pullRequestConfig.Refresh = time.Minute
	subject.navigationState.reviewSession = reviewSessionState{active: true, summary: summary}
	actual := subject.applyScheduledPullRequestRefreshDue(MsgScheduledPullRequestRefreshDue{})
	expected := 2
	if len(actual) != expected {
		t.Fatalf("expected %d commands, actual %v", expected, actual)
	}
	probe, ok := actual[0].(scheduledPullRequestProbeCmd)
	if !ok || len(probe.summaries) != 1 || !samePullRequestIdentity(probe.summaries[0], summary) {
		t.Fatalf("expected active review probe, actual %+v", actual)
	}
}

func TestScheduledRefresh_GivenASeenPastedPullRequest_WhenAProbeFindsChanges_ThenItUpdatesMetadataAndFetchesStaleContentWithoutMarkingSeen(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestsLoaded([]githubdomain.PullRequest{summary})
	})
	subject.updatePullRequestListStore(func(store pullRequestListStore) pullRequestListStore {
		return store.withPullRequestSeen("acme/widgets#42", summary.UpdatedAt)
	})
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{sourceUpdatedAt: summary.UpdatedAt}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{sourceUpdatedAt: summary.UpdatedAt}
	latest := summary
	latest.UpdatedAt = "2026-10-04T13:00:00Z"
	latest.Title = "Remote title"
	batchID := subject.beginScheduledPullRequestRefreshBatch(1, subject.currentTime())
	subject.addScheduledPullRequestRefreshWork(batchID, 1)
	subject.finishScheduledPullRequestRefreshPlanning(batchID)
	actual := Update(subject, MsgScheduledPullRequestProbed{Summary: latest, BatchID: batchID})
	if len(actual) != 1 {
		t.Fatalf("expected refresh handoff, actual %v", actual)
	}
	fanout, ok := actual[0].(scheduledPullRequestRefreshCmd)
	expected := latest.UpdatedAt
	if !ok || len(fanout.detailSummaries) != 1 || len(fanout.diffSummaries) != 1 || fanout.diffSummaries[0].UpdatedAt != expected {
		t.Fatalf("expected new version %s in fetch targets, actual %+v", expected, actual)
	}
	if subject.pastedPullRequests.summaries()[0].UpdatedAt != expected || !subject.pullRequestFreshnessFor("acme/widgets#42").isUnread(expected) {
		t.Fatal("expected updated pasted metadata to remain unread")
	}
	if subject.scheduledPullRequestRefreshBatchCancelled(batchID) {
		t.Fatal("expected batch to remain active through refresh handoff")
	}
}

func TestScheduledRefresh_GivenAnUnchangedProbe_WhenApplyingMetadata_ThenItSkipsCurrentCaches(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{sourceUpdatedAt: summary.UpdatedAt}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{sourceUpdatedAt: summary.UpdatedAt, fileTeamOwnersAttempted: true}
	actual := subject.applyScheduledPullRequestProbed(MsgScheduledPullRequestProbed{Summary: summary})
	fanout := actual[0].(scheduledPullRequestRefreshCmd)
	expected := 0
	if len(fanout.detailSummaries) != expected || len(fanout.diffSummaries) != expected {
		t.Fatalf("expected no stale targets, actual %+v", fanout)
	}
}

func TestScheduledRefresh_GivenDuplicateProbeTargets_WhenExecuting_ThenItQueriesRemotelyOnce(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	loader := &fakePullRequestDetailLoader{}
	subject.detailQueries = loader
	runner := &capturingAsyncRunner{}
	subject.asyncRunner = runner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	scheduledPullRequestProbeCmd{summaries: []githubdomain.PullRequest{summary, summary}}.execute(subject, gui)
	actual := len(runner.runs)
	expected := 1
	if actual != expected {
		t.Fatalf("expected %d remote query, actual %d", expected, actual)
	}
}

func TestRefreshRetry_GivenRepeatedFailures_WhenComputingBackoff_ThenItCapsAtFiveMinutes(t *testing.T) {
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	actual := pullRequestRefreshRetry{}
	expected := []time.Duration{30 * time.Second, time.Minute, 2 * time.Minute, 4 * time.Minute, 5 * time.Minute, 5 * time.Minute}
	for _, delay := range expected {
		actual = actual.afterFailure(now)
		if actual.retryAfter.Sub(now) != delay || actual.ready(now) || !actual.ready(actual.retryAfter) {
			t.Fatalf("expected backoff %v, actual %+v", delay, actual)
		}
	}
}

func TestScheduledRefresh_GivenAProbeFailure_WhenTheNextCycleRuns_ThenItWaitsForBackoffAndResetsAfterSuccess(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	now := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	subject.timingState.now = func() time.Time { return now }
	subject.applyScheduledPullRequestProbed(MsgScheduledPullRequestProbed{Summary: summary, Err: errors.New("offline")})
	subject.detailQueries = &fakePullRequestDetailLoader{}
	runner := &capturingAsyncRunner{}
	subject.asyncRunner = runner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	command := scheduledPullRequestProbeCmd{summaries: []githubdomain.PullRequest{summary}}
	command.execute(subject, gui)
	if actual := len(runner.runs); actual != 0 {
		t.Fatalf("expected no probe during backoff, actual %d", actual)
	}
	now = now.Add(30 * time.Second)
	command.execute(subject, gui)
	if actual := len(runner.runs); actual != 1 {
		t.Fatalf("expected probe after backoff, actual %d", actual)
	}
	subject.applyScheduledPullRequestProbed(MsgScheduledPullRequestProbed{Summary: summary})
	if actual := len(subject.pullRequestProbeRetries); actual != 0 {
		t.Fatalf("expected successful probe to clear retries, actual %d", actual)
	}
}

func TestScheduledRefresh_GivenOnlyAStaleDiff_WhenExecuting_ThenItDoesNotReloadCurrentDetail(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	subject.detailQueries = &fakePullRequestDetailLoader{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{sourceUpdatedAt: summary.UpdatedAt}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{sourceUpdatedAt: "old"}
	runner := &capturingAsyncRunner{}
	subject.asyncRunner = runner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	scheduledPullRequestRefreshCmd{detailSummaries: []githubdomain.PullRequest{summary}, diffSummaries: []githubdomain.PullRequest{summary}}.execute(subject, gui)
	actual := len(runner.runs)
	expected := 1
	if actual != expected || subject.pullRequestDetailLoadInFlight["acme/widgets#42"] || !subject.pullRequestDiffLoadInFlight["acme/widgets#42"] {
		t.Fatalf("expected only diff load, actual %d loads", actual)
	}
}

func TestScheduledRefresh_GivenASkippedListAndActiveReview_WhenPlanningCompletes_ThenItWaitsForTheProbe(t *testing.T) {
	subject := NewProgram()
	summary := given_refreshSummary()
	subject.runtimeConfig.pullRequestConfig.Refresh = time.Minute
	subject.navigationState.reviewSession = reviewSessionState{active: true, summary: summary}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{sourceUpdatedAt: summary.UpdatedAt}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{sourceUpdatedAt: summary.UpdatedAt, fileTeamOwnersAttempted: true}
	subject.detailQueries = &fakePullRequestDetailLoader{}
	runner := &capturingAsyncRunner{}
	subject.asyncRunner = runner
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	commands := subject.applyScheduledPullRequestRefreshDue(MsgScheduledPullRequestRefreshDue{Tabs: []PullRequestTab{MyPullRequestsTab}})
	subject.executeCmds(gui, commands)
	actual := len(subject.scheduledPullRequestRefreshBatches)
	expected := 1
	if actual != expected || len(runner.runs) != expected {
		t.Fatalf("expected one pending probe and batch, actual batches=%d probes=%d", actual, len(runner.runs))
	}
	for _, batch := range subject.scheduledPullRequestRefreshBatches {
		if batch.planning || batch.pending != expected {
			t.Fatalf("expected only probe slot after planning, actual %+v", batch)
		}
	}
}

func TestScheduledRefresh_GivenNoStaleContent_WhenTheFanoutFinishes_ThenItAcknowledgesTheScheduler(t *testing.T) {
	subject := NewProgram()
	subject.detailQueries = &fakePullRequestDetailLoader{}
	scheduler := newPullRequestRefreshScheduler(nil, nil, nil)
	scheduler.started.Store(true)
	scheduler.acknowledgeCh = make(chan uint64, 1)
	subject.pullRequestRefreshScheduler = scheduler
	batchID := subject.beginScheduledPullRequestRefreshBatch(7, subject.currentTime())
	subject.addScheduledPullRequestRefreshWork(batchID, 1)
	subject.finishScheduledPullRequestRefreshPlanning(batchID)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	scheduledPullRequestRefreshCmd{batchID: batchID}.execute(subject, gui)
	select {
	case actual := <-scheduler.acknowledgeCh:
		expected := uint64(7)
		if actual != expected {
			t.Fatalf("expected generation %d, actual %d", expected, actual)
		}
	default:
		t.Fatal("expected scheduler acknowledgement after empty fanout")
	}
}

func TestRefresh_GivenAStaleSummary_WhenDetailReportsANewerVersion_ThenItCachesTheFetchedVersion(t *testing.T) {
	summary := given_refreshSummary()
	subject := NewProgram()
	expected := "2026-10-04T13:00:00Z"
	detail := githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{UpdatedAt: expected})
	subject.applyPullRequestDetailLoaded(MsgPullRequestDetailLoaded{Summary: summary, Detail: detail})
	actual := subject.pullRequestDetailCache["acme/widgets#42"].sourceUpdatedAt
	if actual != expected {
		t.Fatalf("expected fetched version %s, actual %s", expected, actual)
	}
}
