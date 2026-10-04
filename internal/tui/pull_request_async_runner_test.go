package tui

import (
	"reflect"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestPullRequestAsyncRunner_GivenBackgroundBacklog_WhenActiveContentArrives_ThenItUsesReservedCapacity(t *testing.T) {
	// GIVEN
	backend := &capturingAsyncRunner{}
	subject := newPullRequestAsyncRunner(backend)
	subject.setActivePullRequest("visible")
	for range 10 {
		subject.goPullRequest("background", func() {})
	}
	// WHEN
	subject.goPullRequest("visible", func() {})
	subject.goPullRequest("visible", func() {})
	// THEN
	actual := len(backend.runs)
	expected := 4
	if actual != expected {
		t.Fatalf("expected %d dispatched workers, actual %d", expected, actual)
	}
	backend.runs[2]()
	if actual := len(backend.runs); actual != expected {
		t.Fatalf("expected background work to leave visible capacity free, actual %d dispatched workers", actual)
	}
}

func TestPullRequestAsyncRunner_GivenQueuedContent_WhenSelectionChanges_ThenItPromotesBothLoadsAheadOfBackground(t *testing.T) {
	// GIVEN
	backend := &capturingAsyncRunner{}
	subject := newPullRequestAsyncRunner(backend)
	var actual []string
	for _, key := range []string{"running-1", "running-2", "background", "selected", "selected"} {
		subject.goPullRequest(key, func() { actual = append(actual, key) })
	}
	// WHEN
	subject.setActivePullRequest("selected")
	backend.runs[2]()
	backend.runs[3]()
	// THEN
	expected := []string{"selected", "selected"}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %v, actual %v", expected, actual)
	}
	if actual := len(backend.runs); actual != 4 {
		t.Fatalf("expected four dispatched jobs with background still queued, actual %d", actual)
	}
}

func TestPullRequestAsyncRunner_GivenFullPool_WhenWorkCompletes_ThenItDrainsQueueWithoutExceedingLimit(t *testing.T) {
	// GIVEN
	backend := &capturingAsyncRunner{}
	subject := newPullRequestAsyncRunner(backend)
	subject.setActivePullRequest("visible")
	completed := 0
	for range 12 {
		subject.goPullRequest("visible", func() { completed++ })
	}
	// WHEN
	for index := 0; index < len(backend.runs); index++ {
		if actual := len(backend.runs) - index; actual > 4 {
			t.Fatalf("expected at most four active workers, actual %d", actual)
		}
		backend.runs[index]()
	}
	// THEN
	actual := completed
	expected := 12
	if actual != expected {
		t.Fatalf("expected %d completed jobs, actual %d", expected, actual)
	}
}

func TestPullRequestAsyncRunner_GivenSaturatedReads_WhenUnrelatedAsyncWorkArrives_ThenItDoesNotQueueTimersOrMutations(t *testing.T) {
	// GIVEN
	backend := &capturingAsyncRunner{}
	subject := newPullRequestAsyncRunner(backend)
	for range 10 {
		subject.goPullRequest("background", func() {})
	}
	// WHEN
	subject.Go(func() {})
	// THEN
	actual := len(backend.runs)
	expected := 3
	if actual != expected {
		t.Fatalf("expected %d dispatched jobs, actual %d", expected, actual)
	}
}

func TestPullRequestAsyncRunner_GivenBlockedBackgroundRequests_WhenLoadingVisibleContent_ThenItCompletesWithoutWaiting(t *testing.T) {
	// GIVEN
	subject := newPullRequestAsyncRunner(goroutineAsyncRunner{})
	subject.setActivePullRequest("visible")
	releaseBackground := make(chan struct{})
	defer close(releaseBackground)
	for range 100 {
		subject.goPullRequest("background", func() { <-releaseBackground })
	}
	completed := make(chan struct{}, 2)
	startedAt := time.Now()
	// WHEN
	for range 2 {
		subject.goPullRequest("visible", func() { completed <- struct{}{} })
	}
	// THEN
	deadline := time.NewTimer(time.Second)
	defer deadline.Stop()
	for range 2 {
		select {
		case <-completed:
		case <-deadline.C:
			t.Fatal("expected visible content to complete while background requests remain blocked")
		}
	}
	t.Logf("visible description and diff completed in %s with 100 blocked background requests", time.Since(startedAt))
}

func TestPullRequestCommands_GivenQueuedBackgroundLoads_WhenSelectingTheirPR_ThenItPromotesExistingWorkWithoutDuplicatingLoads(t *testing.T) {
	// GIVEN
	active := given_refreshSummary()
	background := active
	background.Number++
	subject := NewProgram()
	subject.detailQueries = &fakePullRequestDetailLoader{}
	subject.navigationState.reviewSession = reviewSessionState{active: true, summary: active}
	backend := &capturingAsyncRunner{}
	subject.asyncRunner = newPullRequestAsyncRunner(backend)
	gui := given_headlessGui(t)
	defer gui.Close()
	for _, number := range []int{100, 101} {
		blockingSummary := active
		blockingSummary.Number = number
		(loadPullRequestDetailCmd{summary: blockingSummary}).execute(subject, gui)
	}
	scheduledPullRequestRefreshCmd{
		detailSummaries: []githubdomain.PullRequest{background},
		diffSummaries:   []githubdomain.PullRequest{background},
	}.execute(subject, gui)
	if actual := len(backend.runs); actual != 2 {
		t.Fatalf("expected two background workers before navigation, actual %d", actual)
	}
	// WHEN
	subject.navigationState.reviewSession.summary = background
	if err := subject.afterStateChange(nil); err != nil {
		t.Fatal(err)
	}
	// THEN
	actual := len(backend.runs)
	expected := 4
	if actual != expected {
		t.Fatalf("expected existing detail and diff jobs to be promoted, actual %d dispatched workers", actual)
	}
	if actual := subject.selectedPullRequestDetailLoadPlan(); len(actual.commands) != 0 {
		t.Fatalf("expected no duplicate detail fetch, actual %+v", actual)
	}
	if actual := subject.selectedPullRequestDiffLoadPlan(); len(actual.commands) != 0 {
		t.Fatalf("expected no duplicate diff fetch, actual %+v", actual)
	}
}

func TestPullRequestCommands_GivenScheduledListsAndProbes_WhenContentIsQueued_ThenItSharesTheReadWorkerLimit(t *testing.T) {
	// GIVEN
	subject := NewProgram()
	subject.detailQueries = &fakePullRequestDetailLoader{}
	subject.pullRequestListQueries = &fakePullRequestListQueries{}
	backend := &capturingAsyncRunner{}
	subject.asyncRunner = newPullRequestAsyncRunner(backend)
	gui := given_headlessGui(t)
	defer gui.Close()
	summary := given_refreshSummary()
	scheduledSummary := summary
	scheduledSummary.Number++
	// WHEN
	(loadPullRequestsCmd{Source: pullRequestLoadSourceScheduled}).execute(subject, gui)
	(scheduledPullRequestProbeCmd{summaries: []githubdomain.PullRequest{scheduledSummary}}).execute(subject, gui)
	(loadPullRequestDetailCmd{summary: scheduledSummary}).execute(subject, gui)
	(loadPullRequestDiffCmd{summary: scheduledSummary}).execute(subject, gui)
	// THEN
	actual := len(backend.runs)
	expected := 2
	if actual != expected {
		t.Fatalf("expected list and probe to occupy the two background workers, actual %d dispatched jobs", actual)
	}
}

func TestScheduledRefreshPlan_GivenActivePRAfterBackgroundPR_WhenPlanning_ThenItLoadsActiveDescriptionAndDiffFirst(t *testing.T) {
	// GIVEN
	background := given_refreshSummary()
	active := background
	active.Number++
	// WHEN
	actual := planScheduledPullRequestRefresh(scheduledPullRequestRefreshPlanInput{
		detailSummaries:  []githubdomain.PullRequest{background, active},
		diffSummaries:    []githubdomain.PullRequest{background, active},
		hasDetailQueries: true,
		activeKey:        pullRequestDetailKey(active.Repository, active.Number),
	})
	// THEN
	expected := []Cmd{loadPullRequestDetailCmd{summary: active}, loadPullRequestDiffCmd{summary: active}, loadPullRequestDetailCmd{summary: background}, loadPullRequestDiffCmd{summary: background}}
	if !reflect.DeepEqual(actual.commands, expected) {
		t.Fatalf("expected active description and diff first: %+v, actual %+v", expected, actual.commands)
	}
}
