package tui

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

type fakePullRequestRefreshTimer struct {
	mu          sync.Mutex
	channel     chan time.Time
	active      bool
	stopCount   int
	resetDelays []time.Duration
	resetSignal chan time.Duration
}

func newFakePullRequestRefreshTimer(delay time.Duration) *fakePullRequestRefreshTimer {
	timer := &fakePullRequestRefreshTimer{channel: make(chan time.Time, 10), resetSignal: make(chan time.Duration, 10)}
	timer.Reset(delay)
	return timer
}

func (timer *fakePullRequestRefreshTimer) C() <-chan time.Time { return timer.channel }

func (timer *fakePullRequestRefreshTimer) Stop() bool {
	timer.mu.Lock()
	defer timer.mu.Unlock()
	wasActive := timer.active
	timer.active = false
	timer.stopCount++
	return wasActive
}

func (timer *fakePullRequestRefreshTimer) Reset(delay time.Duration) bool {
	timer.mu.Lock()
	defer timer.mu.Unlock()
	wasActive := timer.active
	timer.active = true
	timer.resetDelays = append(timer.resetDelays, delay)
	timer.resetSignal <- delay
	return wasActive
}

func (timer *fakePullRequestRefreshTimer) fire(at time.Time) {
	timer.mu.Lock()
	timer.active = false
	timer.mu.Unlock()
	timer.channel <- at
}

func (timer *fakePullRequestRefreshTimer) lastResetDelay() time.Duration {
	timer.mu.Lock()
	defer timer.mu.Unlock()
	if len(timer.resetDelays) == 0 {
		return 0
	}
	return timer.resetDelays[len(timer.resetDelays)-1]
}

func TestPullRequestRefreshSchedule_GivenAnInitialGeneration_WhenCreated_ThenItPreservesTheGeneration(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	actual := newPullRequestRefreshSchedule(appconfig.PullRequestConfig{}, givenNow, 42)
	if actual.generation != 42 {
		t.Fatalf("expected schedule generation %d, actual %d", 42, actual.generation)
	}
}

func TestPullRequestRefreshSchedule_GivenGlobalIntervalAndOptOuts_WhenCreated_ThenItIncludesOnlyEligibleTabsAndPastedPolicy(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	actual := newPullRequestRefreshSchedule(appconfig.PullRequestConfig{
		Refresh: 5 * time.Minute,
		Searches: []appconfig.PullRequestSearch{
			{Label: "enabled", AutoRefresh: true},
			{Label: "disabled", AutoRefresh: false},
		},
		PastedPRs: appconfig.PastedPullRequestConfig{AutoRefresh: true},
	}, givenNow, 1)
	if !reflect.DeepEqual(actual.tabs, []PullRequestTab{MyPullRequestsTab}) || !actual.refreshPasted {
		t.Fatalf("expected one eligible tab and pasted refresh, actual %+v", actual)
	}
	if expected := givenNow.Add(5 * time.Minute); !actual.nextDue.Equal(expected) {
		t.Fatalf("expected first due time %v, actual %v", expected, actual.nextDue)
	}
}

func TestPullRequestRefreshSchedule_GivenManualMode_WhenCreated_ThenItHasNoDueTime(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	actual := newPullRequestRefreshSchedule(appconfig.PullRequestConfig{
		Searches:  []appconfig.PullRequestSearch{{Label: "enabled", AutoRefresh: true}},
		PastedPRs: appconfig.PastedPullRequestConfig{AutoRefresh: true},
	}, givenNow, 1)
	if !actual.nextDue.IsZero() {
		t.Fatalf("expected manual mode to have no due time, actual %v", actual.nextDue)
	}
}

func TestPullRequestRefreshSchedule_GivenLateTime_WhenAdvancing_ThenItSkipsMissedOccurrencesWithoutDrifting(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	subject := newPullRequestRefreshSchedule(appconfig.PullRequestConfig{
		Refresh:  time.Minute,
		Searches: []appconfig.PullRequestSearch{{Label: "fast", AutoRefresh: true}},
	}, givenNow, 0)

	actualTabs, actualPasted, actualNextDue := subject.advance(givenNow.Add(4*time.Minute + 30*time.Second))
	if !reflect.DeepEqual(actualTabs, []PullRequestTab{MyPullRequestsTab}) || actualPasted {
		t.Fatalf("expected one coalesced due tab, actual %v, pasted=%v", actualTabs, actualPasted)
	}
	if expected := givenNow.Add(5 * time.Minute); !actualNextDue.Equal(expected) {
		t.Fatalf("expected the next fixed phase %v, actual %v", expected, actualNextDue)
	}
}

func TestPullRequestRefreshScheduler_GivenManualSchedule_WhenStarted_ThenItDoesNotCreateATimer(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	created := 0
	subject := newPullRequestRefreshScheduler(func() time.Time { return givenNow }, func(time.Duration) pullRequestRefreshTimer {
		created++
		return newFakePullRequestRefreshTimer(time.Minute)
	}, nil)
	stop := subject.Start(newPullRequestRefreshSchedule(appconfig.PullRequestConfig{}, givenNow, 1))
	stop()
	if created != 0 {
		t.Fatalf("expected no timer in manual mode, created %d", created)
	}
}

func TestPullRequestRefreshScheduler_GivenAConfiguredSchedule_WhenTheFirstIntervalElapses_ThenItDispatchesTheTriggerTime(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	var clockMu sync.RWMutex
	clock := givenNow
	now := func() time.Time {
		clockMu.RLock()
		defer clockMu.RUnlock()
		return clock
	}
	setClock := func(value time.Time) {
		clockMu.Lock()
		defer clockMu.Unlock()
		clock = value
	}
	timer := newFakePullRequestRefreshTimer(time.Minute)
	dispatched := make(chan Msg, 1)
	subject := newPullRequestRefreshScheduler(now, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(newPullRequestRefreshSchedule(appconfig.PullRequestConfig{Refresh: time.Minute, Searches: []appconfig.PullRequestSearch{{AutoRefresh: true}}}, givenNow, 3))
	defer stop()

	triggeredAt := givenNow.Add(time.Minute)
	setClock(triggeredAt)
	timer.fire(triggeredAt)
	message := <-dispatched
	actual, ok := message.(MsgScheduledPullRequestRefreshDue)
	if !ok || !actual.TriggeredAt.Equal(triggeredAt) || actual.Generation != 3 {
		t.Fatalf("expected a timestamped due message, actual %#v", message)
	}
}

func TestPullRequestRefreshScheduler_GivenAPendingDueMessage_WhenTheTimerFiresAgain_ThenItCoalescesTheOccurrence(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	var clockMu sync.RWMutex
	clock := givenNow
	now := func() time.Time {
		clockMu.RLock()
		defer clockMu.RUnlock()
		return clock
	}
	setClock := func(value time.Time) {
		clockMu.Lock()
		defer clockMu.Unlock()
		clock = value
	}
	timer := newFakePullRequestRefreshTimer(time.Minute)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(now, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(newPullRequestRefreshSchedule(appconfig.PullRequestConfig{Refresh: time.Minute, Searches: []appconfig.PullRequestSearch{{AutoRefresh: true}}}, givenNow, 4))
	defer stop()

	firstDue := givenNow.Add(time.Minute)
	setClock(firstDue)
	timer.fire(firstDue)
	<-dispatched
	secondDue := givenNow.Add(2 * time.Minute)
	setClock(secondDue)
	timer.fire(secondDue)
	select {
	case actual := <-dispatched:
		t.Fatalf("expected the occurrence to be coalesced, actual %T", actual)
	default:
	}
}

func TestPullRequestRefreshScheduler_GivenAQueuedDueMessage_WhenReconfigured_ThenItStartsAFreshGeneration(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	var clockMu sync.RWMutex
	clock := givenNow
	now := func() time.Time {
		clockMu.RLock()
		defer clockMu.RUnlock()
		return clock
	}
	setClock := func(value time.Time) {
		clockMu.Lock()
		defer clockMu.Unlock()
		clock = value
	}
	timer := newFakePullRequestRefreshTimer(time.Minute)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(now, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(newPullRequestRefreshSchedule(appconfig.PullRequestConfig{Refresh: time.Minute, Searches: []appconfig.PullRequestSearch{{AutoRefresh: true}}}, givenNow, 6))
	defer stop()
	firstDue := givenNow.Add(time.Minute)
	setClock(firstDue)
	timer.fire(firstDue)
	<-dispatched

	reconfiguredAt := givenNow.Add(10 * time.Second)
	setClock(reconfiguredAt)
	subject.Configure(newPullRequestRefreshSchedule(appconfig.PullRequestConfig{Refresh: time.Minute, Searches: []appconfig.PullRequestSearch{{AutoRefresh: true}}}, reconfiguredAt, 7))
	waitForFakePullRequestRefreshTimerDelay(t, timer, time.Minute)
	subject.Acknowledge(6)
	secondDue := givenNow.Add(70 * time.Second)
	setClock(secondDue)
	timer.fire(secondDue)
	select {
	case message := <-dispatched:
		actual, ok := message.(MsgScheduledPullRequestRefreshDue)
		if !ok || actual.Generation != 7 {
			t.Fatalf("expected only the new generation, actual %#v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("expected the new generation to dispatch")
	}
}

func waitForFakePullRequestRefreshTimerDelay(t *testing.T, timer *fakePullRequestRefreshTimer, expected time.Duration) {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		select {
		case actual := <-timer.resetSignal:
			if actual == expected {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for timer delay %v, actual %v", expected, timer.lastResetDelay())
		default:
			runtime.Gosched()
		}
	}
}
