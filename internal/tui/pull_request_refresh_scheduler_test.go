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

func (timer *fakePullRequestRefreshTimer) C() <-chan time.Time {
	return timer.channel
}

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

	actual := newPullRequestRefreshSchedule(nil, givenNow, 42)

	if actual.generation != 42 {
		t.Fatalf("expected schedule generation %d, actual %d", 42, actual.generation)
	}
}

func TestPullRequestRefreshSchedule_GivenSeveralIntervals_WhenTheEarliestDueTimeArrives_ThenItReturnsOnlyDueTabs(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	subject := newPullRequestRefreshSchedule([]appconfig.PullRequestSearch{
		{Label: "fast", Refresh: 5 * time.Minute},
		{Label: "slow", Refresh: 10 * time.Minute},
		{Label: "manual"},
	}, givenNow, 0)

	actual, nextDue := subject.advance(givenNow.Add(5 * time.Minute))

	if !reflect.DeepEqual(actual.tabs, []PullRequestTab{MyPullRequestsTab}) || actual.refreshPasted {
		t.Fatalf("expected only the first search to be due, actual %+v", actual)
	}
	if expected := givenNow.Add(10 * time.Minute); !nextDue.Equal(expected) {
		t.Fatalf("expected next due time %v, actual %v", expected, nextDue)
	}
}

func TestPullRequestRefreshSchedule_GivenLateTime_WhenAdvancing_ThenItSkipsMissedOccurrencesWithoutBursting(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	subject := newPullRequestRefreshSchedule([]appconfig.PullRequestSearch{{Label: "fast", Refresh: time.Minute}}, givenNow, 0)

	actual, nextDue := subject.advance(givenNow.Add(4*time.Minute + 30*time.Second))

	if !reflect.DeepEqual(actual.tabs, []PullRequestTab{MyPullRequestsTab}) {
		t.Fatalf("expected one coalesced due tab, actual %+v", actual.tabs)
	}
	expectedNextDue := givenNow.Add(5 * time.Minute)
	if !nextDue.Equal(expectedNextDue) {
		t.Fatalf("expected the next phase-aligned due time %v, actual %v", expectedNextDue, nextDue)
	}
}

func TestPullRequestRefreshSchedule_GivenPastedRefresh_WhenAdvancingEveryTenMinutes_ThenItReturnsThePastedTarget(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	subject := newPullRequestRefreshSchedule(nil, givenNow, 0)

	actual, nextDue := subject.advance(givenNow.Add(10 * time.Minute))

	if !actual.refreshPasted || len(actual.tabs) != 0 {
		t.Fatalf("expected only the pasted target to be due, actual %+v", actual)
	}
	if expected := givenNow.Add(20 * time.Minute); !nextDue.Equal(expected) {
		t.Fatalf("expected pasted next due time %v, actual %v", expected, nextDue)
	}
}

func TestPullRequestRefreshSchedule_GivenATinyIntervalAndLongLateness_WhenAdvancing_ThenItSkipsMissedOccurrencesWithoutIteratingPerOccurrence(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	subject := newPullRequestRefreshSchedule([]appconfig.PullRequestSearch{{Label: "tiny", Refresh: time.Nanosecond}}, givenNow, 0)

	actual, nextDue := subject.advance(givenNow.Add(24 * time.Hour))

	if !reflect.DeepEqual(actual.tabs, []PullRequestTab{MyPullRequestsTab}) {
		t.Fatalf("expected one tiny-interval due tab, actual %+v", actual.tabs)
	}
	if !nextDue.After(givenNow.Add(24 * time.Hour)) {
		t.Fatalf("expected the next due time after now, actual %v", nextDue)
	}
}

func TestPullRequestRefreshScheduler_GivenAConfiguredSchedule_WhenStopped_ThenItStopsTheTimerAndDoesNotDispatchFurtherTicks(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "fast", Refresh: time.Minute}}, givenNow, 1))

	clock = givenNow.Add(time.Minute)
	timer.fire(clock)
	<-dispatched
	stop()

	clock = givenNow.Add(2 * time.Minute)
	timer.fire(clock)
	select {
	case actual := <-dispatched:
		t.Fatalf("expected no dispatch after stop, actual %T", actual)
	default:
	}
	if timer.stopCount == 0 {
		t.Fatal("expected the scheduler to stop the timer")
	}
}

func TestPullRequestRefreshScheduler_GivenAReconfiguredSchedule_WhenConfiguredAgain_ThenItResetsDueTimesFromNow(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(Msg) bool { return true })
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "old", Refresh: time.Minute}}, givenNow, 1))
	defer stop()

	clock = givenNow.Add(20 * time.Second)
	subject.Configure(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "new", Refresh: 5 * time.Minute}}, clock, 2))
	waitForFakePullRequestRefreshTimerDelay(t, timer, 5*time.Minute)

	if expected := 5 * time.Minute; timer.lastResetDelay() != expected {
		t.Fatalf("expected the reconfigured timer delay %v, actual %v", expected, timer.lastResetDelay())
	}
}

func TestPullRequestRefreshScheduler_GivenAPendingDueMessage_WhenTheTimerFiresAgain_ThenItCoalescesTheOccurrence(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "fast", Refresh: time.Minute}}, givenNow, 3))
	defer stop()

	clock = givenNow.Add(time.Minute)
	timer.fire(clock)
	<-dispatched
	clock = givenNow.Add(2 * time.Minute)
	timer.fire(clock)

	select {
	case actual := <-dispatched:
		t.Fatalf("expected the pending occurrence to be coalesced, actual %T", actual)
	default:
	}
}

func TestPullRequestRefreshScheduler_GivenAPendingDueMessage_WhenAcknowledged_ThenItResumesAtTheNextFixedCadence(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "fast", Refresh: time.Minute}}, givenNow, 4))
	defer stop()

	clock = givenNow.Add(time.Minute)
	timer.fire(clock)
	<-dispatched
	clock = givenNow.Add(2*time.Minute + 30*time.Second)
	subject.Acknowledge(4)
	waitForFakePullRequestRefreshTimerDelay(t, timer, 30*time.Second)

	if expected := 30 * time.Second; timer.lastResetDelay() != expected {
		t.Fatalf("expected the next phase-aligned delay %v, actual %v", expected, timer.lastResetDelay())
	}
}

func TestPullRequestRefreshScheduler_GivenAStaleAcknowledgement_WhenAcknowledged_ThenItLeavesTheCurrentGenerationPending(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "fast", Refresh: time.Minute}}, givenNow, 5))
	defer stop()

	clock = givenNow.Add(time.Minute)
	timer.fire(clock)
	<-dispatched
	subject.Acknowledge(4)
	clock = givenNow.Add(2 * time.Minute)
	timer.fire(clock)

	select {
	case actual := <-dispatched:
		t.Fatalf("expected stale acknowledgement to keep the message pending, actual %T", actual)
	default:
	}
}

func TestPullRequestRefreshScheduler_GivenAQueuedDueMessage_WhenReconfigured_ThenItInvalidatesTheOldGeneration(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	dispatched := make(chan Msg, 2)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "old", Refresh: time.Minute}}, givenNow, 6))
	defer stop()

	clock = givenNow.Add(time.Minute)
	timer.fire(clock)
	<-dispatched
	clock = givenNow.Add(10 * time.Second)
	subject.Configure(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "new", Refresh: time.Minute}}, clock, 7))
	waitForFakePullRequestRefreshTimerDelay(t, timer, time.Minute)
	subject.Acknowledge(6)
	clock = givenNow.Add(70 * time.Second)
	timer.fire(clock)
	select {
	case message := <-dispatched:
		actual, ok := message.(MsgScheduledPullRequestRefreshDue)
		if !ok || actual.Generation != 7 {
			t.Fatalf("expected only the new generation to dispatch, actual %#v", message)
		}
	case <-time.After(time.Second):
		t.Fatal("expected the new generation to dispatch")
	}
}

func TestPullRequestRefreshScheduler_GivenAQueuedDueMessage_WhenStopped_ThenStopReturnsWithoutWaitingForAcknowledgement(t *testing.T) {
	givenNow := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := givenNow
	timer := newFakePullRequestRefreshTimer(0)
	dispatched := make(chan Msg, 1)
	subject := newPullRequestRefreshScheduler(func() time.Time { return clock }, func(time.Duration) pullRequestRefreshTimer { return timer }, func(message Msg) bool {
		dispatched <- message
		return true
	})
	stop := subject.Start(scheduleWithGeneration([]appconfig.PullRequestSearch{{Label: "fast", Refresh: time.Minute}}, givenNow, 8))
	clock = givenNow.Add(time.Minute)
	timer.fire(clock)
	<-dispatched

	stop()
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

func scheduleWithGeneration(searches []appconfig.PullRequestSearch, now time.Time, generation uint64) pullRequestRefreshSchedule {
	return newPullRequestRefreshSchedule(searches, now, generation)
}
