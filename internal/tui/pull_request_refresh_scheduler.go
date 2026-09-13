package tui

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/jesseduffield/gocui"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

const pastedPullRequestRefreshInterval = 10 * time.Minute

type pullRequestRefreshScheduleEntry struct {
	tab      PullRequestTab
	interval time.Duration
	nextDue  time.Time
}

type pullRequestRefreshSchedule struct {
	searches      []pullRequestRefreshScheduleEntry
	pastedNextDue time.Time
	generation    uint64
}

type pullRequestRefreshDue struct {
	tabs          []PullRequestTab
	refreshPasted bool
}

func newPullRequestRefreshSchedule(searches []appconfig.PullRequestSearch, now time.Time, generation uint64) pullRequestRefreshSchedule {
	entries := make([]pullRequestRefreshScheduleEntry, 0, len(searches))
	for index, search := range searches {
		if search.Refresh <= 0 {
			continue
		}
		entries = append(entries, pullRequestRefreshScheduleEntry{
			tab:      PullRequestTab(index),
			interval: search.Refresh,
			nextDue:  now.Add(search.Refresh),
		})
	}
	return pullRequestRefreshSchedule{
		searches:      entries,
		pastedNextDue: now.Add(pastedPullRequestRefreshInterval),
		generation:    generation,
	}
}

func (schedule *pullRequestRefreshSchedule) advance(now time.Time) (pullRequestRefreshDue, time.Time) {
	if schedule == nil {
		return pullRequestRefreshDue{}, time.Time{}
	}

	due := pullRequestRefreshDue{}
	for index := range schedule.searches {
		entry := &schedule.searches[index]
		if entry.nextDue.After(now) {
			continue
		}
		due.tabs = append(due.tabs, entry.tab)
		entry.nextDue = advancePullRequestRefreshDueTime(entry.nextDue, entry.interval, now)
	}
	if !schedule.pastedNextDue.After(now) {
		due.refreshPasted = true
		schedule.pastedNextDue = advancePullRequestRefreshDueTime(schedule.pastedNextDue, pastedPullRequestRefreshInterval, now)
	}

	return due, schedule.nextDue()
}

func advancePullRequestRefreshDueTime(nextDue time.Time, interval time.Duration, now time.Time) time.Time {
	elapsed := now.Sub(nextDue)
	remaining := elapsed % interval
	if remaining == 0 {
		return now.Add(interval)
	}
	return now.Add(interval - remaining)
}

func (schedule pullRequestRefreshSchedule) nextDue() time.Time {
	nextDue := schedule.pastedNextDue
	for _, entry := range schedule.searches {
		if nextDue.IsZero() || entry.nextDue.Before(nextDue) {
			nextDue = entry.nextDue
		}
	}
	return nextDue
}

type pullRequestRefreshTimer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(time.Duration) bool
}

type pullRequestRefreshTimerFactory func(time.Duration) pullRequestRefreshTimer

type realPullRequestRefreshTimer struct {
	timer *time.Timer
}

func (timer realPullRequestRefreshTimer) C() <-chan time.Time {
	return timer.timer.C
}

func (timer realPullRequestRefreshTimer) Stop() bool {
	return timer.timer.Stop()
}

func (timer realPullRequestRefreshTimer) Reset(delay time.Duration) bool {
	return timer.timer.Reset(delay)
}

type pullRequestRefreshScheduler struct {
	now           func() time.Time
	newTimer      pullRequestRefreshTimerFactory
	dispatch      func(Msg) bool
	configureCh   chan pullRequestRefreshSchedule
	acknowledgeCh chan uint64
	stopCh        chan struct{}
	stoppedCh     chan struct{}
	started       atomic.Bool
	stopped       atomic.Bool
	stopOnce      sync.Once
}

func newPullRequestRefreshScheduler(
	now func() time.Time,
	newTimer pullRequestRefreshTimerFactory,
	dispatch func(Msg) bool,
) *pullRequestRefreshScheduler {
	if now == nil {
		now = time.Now
	}
	if newTimer == nil {
		newTimer = func(delay time.Duration) pullRequestRefreshTimer {
			return realPullRequestRefreshTimer{timer: time.NewTimer(delay)}
		}
	}
	return &pullRequestRefreshScheduler{
		now:           now,
		newTimer:      newTimer,
		dispatch:      dispatch,
		configureCh:   make(chan pullRequestRefreshSchedule),
		acknowledgeCh: make(chan uint64),
		stopCh:        make(chan struct{}),
		stoppedCh:     make(chan struct{}),
	}
}

func (scheduler *pullRequestRefreshScheduler) Start(schedule pullRequestRefreshSchedule) (stop func()) {
	if scheduler == nil {
		return func() {}
	}
	stop = scheduler.stop
	if !scheduler.started.CompareAndSwap(false, true) {
		return stop
	}

	go scheduler.run(clonePullRequestRefreshSchedule(schedule))
	return stop
}

func (scheduler *pullRequestRefreshScheduler) Configure(schedule pullRequestRefreshSchedule) {
	if scheduler == nil || !scheduler.started.Load() || scheduler.stopped.Load() {
		return
	}
	select {
	case scheduler.configureCh <- clonePullRequestRefreshSchedule(schedule):
	case <-scheduler.stopCh:
	}
}

func (scheduler *pullRequestRefreshScheduler) Acknowledge(generation uint64) {
	if scheduler == nil || !scheduler.started.Load() || scheduler.stopped.Load() {
		return
	}
	select {
	case scheduler.acknowledgeCh <- generation:
	case <-scheduler.stopCh:
	}
}

func (scheduler *pullRequestRefreshScheduler) stop() {
	if scheduler == nil || !scheduler.started.Load() {
		return
	}
	scheduler.stopOnce.Do(func() {
		close(scheduler.stopCh)
	})
	<-scheduler.stoppedCh
}

func (scheduler *pullRequestRefreshScheduler) run(schedule pullRequestRefreshSchedule) {
	defer func() {
		scheduler.stopped.Store(true)
		close(scheduler.stoppedCh)
	}()

	timer := scheduler.newTimer(scheduler.timerDelay(schedule))
	if timer == nil {
		<-scheduler.stopCh
		return
	}
	defer stopPullRequestRefreshTimer(timer)

	pending := false
	for {
		select {
		case <-scheduler.stopCh:
			return
		case nextSchedule := <-scheduler.configureCh:
			stopPullRequestRefreshTimer(timer)
			schedule = clonePullRequestRefreshSchedule(nextSchedule)
			pending = false
			timer.Reset(scheduler.timerDelay(schedule))
		case generation := <-scheduler.acknowledgeCh:
			if generation != schedule.generation || !pending {
				continue
			}
			pending = false
			now := scheduler.now()
			_, nextDue := schedule.advance(now)
			stopPullRequestRefreshTimer(timer)
			timer.Reset(durationUntil(nextDue, now))
		case <-timer.C():
			now := scheduler.now()
			due, nextDue := schedule.advance(now)
			if pending {
				continue
			}
			if len(due.tabs) == 0 && !due.refreshPasted {
				timer.Reset(durationUntil(nextDue, now))
				continue
			}

			pending = true
			message := MsgScheduledPullRequestRefreshDue{
				Tabs:          append([]PullRequestTab(nil), due.tabs...),
				RefreshPasted: due.refreshPasted,
				Generation:    schedule.generation,
			}
			if scheduler.dispatch == nil || !scheduler.dispatch(message) {
				pending = false
				timer.Reset(durationUntil(nextDue, now))
			}
		}
	}
}

func (scheduler *pullRequestRefreshScheduler) timerDelay(schedule pullRequestRefreshSchedule) time.Duration {
	return durationUntil(schedule.nextDue(), scheduler.now())
}

func durationUntil(due time.Time, now time.Time) time.Duration {
	if due.IsZero() {
		return 0
	}
	if delay := due.Sub(now); delay > 0 {
		return delay
	}
	return 0
}

func stopPullRequestRefreshTimer(timer pullRequestRefreshTimer) {
	if timer == nil || timer.Stop() {
		return
	}
	select {
	case <-timer.C():
	default:
	}
}

func clonePullRequestRefreshSchedule(schedule pullRequestRefreshSchedule) pullRequestRefreshSchedule {
	schedule.searches = append([]pullRequestRefreshScheduleEntry(nil), schedule.searches...)
	return schedule
}

type configurePullRequestRefreshSchedulerCmd struct {
	searches   []appconfig.PullRequestSearch
	generation uint64
}

func (command configurePullRequestRefreshSchedulerCmd) execute(program *Program, _ *gocui.Gui) {
	if program == nil || program.pullRequestRefreshScheduler == nil {
		return
	}
	scheduler := program.pullRequestRefreshScheduler
	schedule := newPullRequestRefreshSchedule(command.searches, scheduler.now(), command.generation)
	scheduler.Configure(schedule)
}

type acknowledgePullRequestRefreshDueCmd struct {
	generation uint64
}

func (command acknowledgePullRequestRefreshDueCmd) execute(program *Program, _ *gocui.Gui) {
	if program == nil || program.pullRequestRefreshScheduler == nil {
		return
	}
	program.pullRequestRefreshScheduler.Acknowledge(command.generation)
}
