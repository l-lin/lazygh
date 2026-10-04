package tui

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestCommandHistoryStore_GivenOutOfOrderLifecycleEvents_WhenReducingThem_ThenRetainsOneCompletedEntryPerIdentity(t *testing.T) {
	// GIVEN
	startedAt := time.Date(2026, 10, 4, 12, 0, 0, 0, time.Local)
	running := githubdomain.CommandActivity{ID: "gh-1", Command: "gh api user", StartedAt: startedAt, Outcome: githubdomain.CommandRunning}
	completed := running
	completed.FinishedAt = startedAt.Add(123 * time.Millisecond)
	completed.Outcome = githubdomain.CommandSucceeded
	for _, events := range [][]githubdomain.CommandActivity{{running, completed}, {completed, running}, {running, completed, completed}} {
		subject := NewProgramWithModel(given_model())
		// WHEN
		for _, activity := range events {
			Update(subject, MsgCommandActivity{Activity: activity})
		}
		// THEN
		actual := subject.renderCommandHistoryPopupBody()
		expected := "12:00:00 gh api user [success · 123ms]"
		if actual != expected {
			t.Fatalf("expected %q, actual %q", expected, actual)
		}
	}
}

func TestCommandHistoryStore_GivenCompletedRefresh_WhenRendering_ThenShowsDurationQueueWaitAndOwnership(t *testing.T) {
	// GIVEN
	startedAt := time.Date(2026, 10, 4, 12, 0, 0, 0, time.Local)
	activity := githubdomain.CommandActivity{ID: "refresh-1", Command: "refresh diff acme/widgets#42", StartedAt: startedAt, FinishedAt: startedAt.Add(2 * time.Second), Outcome: githubdomain.CommandFailed, QueueWait: 350 * time.Millisecond, Owner: "scheduled batch=7"}
	subject := newCommandHistoryStore().withActivity(activity)
	before := subject.entriesSnapshot()
	// WHEN
	updated := subject.withActivity(githubdomain.CommandActivity{ID: "refresh-2", Command: "refresh detail acme/widgets#42", StartedAt: startedAt.Add(time.Second), Outcome: githubdomain.CommandRunning, Owner: "interactive"})
	actual := renderCommandHistory(updated.entriesSnapshot())
	// THEN
	expected := "12:00:00 refresh diff acme/widgets#42 [failure · 2s · queue=350ms · scheduled batch=7]\n12:00:01 refresh detail acme/widgets#42 [running · queue=0s · interactive]"
	if actual != expected {
		t.Fatalf("expected %q, actual %q", expected, actual)
	}
	if !reflect.DeepEqual(subject.entriesSnapshot(), before) {
		t.Fatal("expected original snapshot to stay unchanged")
	}
}

func TestCommandHistoryStore_GivenIdenticalLegacyAndObservedCommands_WhenAppendingActivity_ThenOnlyTheObservedEntryGetsLifecycleData(t *testing.T) {
	// GIVEN
	startedAt := time.Now()
	subject := newCommandHistoryStore().withCommandStarted("gh api user", startedAt)
	activity := githubdomain.CommandActivity{ID: "gh-1", Command: "gh api user", StartedAt: startedAt, Outcome: githubdomain.CommandRunning}
	// WHEN
	subject = subject.withActivity(activity)
	// THEN
	actual := subject.entriesSnapshot()
	if len(actual) != 2 || actual[0].activity.ID != "" || actual[1].activity.ID != activity.ID {
		t.Fatalf("expected only the appended entry to get telemetry, actual %+v", actual)
	}
}

func TestCommandHistoryStore_GivenEvictedExecution_WhenCompletionArrives_ThenKeepsTheRetentionLimitAndNewerEntries(t *testing.T) {
	// GIVEN
	startedAt := time.Now()
	subject := *newCommandHistoryStore()
	first := githubdomain.CommandActivity{ID: "gh-0", Command: "gh api first", StartedAt: startedAt, Outcome: githubdomain.CommandRunning}
	subject = subject.withActivity(first)
	for index := 1; index <= maxCommandHistoryEntries; index++ {
		subject = subject.withActivity(githubdomain.CommandActivity{ID: fmt.Sprintf("gh-%d", index), Command: "gh api user", StartedAt: startedAt.Add(time.Duration(index) * time.Second), Outcome: githubdomain.CommandRunning})
	}
	// WHEN
	first.Outcome = githubdomain.CommandSucceeded
	first.FinishedAt = startedAt.Add(time.Hour)
	subject = subject.withActivity(first)
	// THEN
	actual := subject.entriesSnapshot()
	expected := maxCommandHistoryEntries
	if len(actual) != expected || actual[0].activity.ID != "gh-1" {
		t.Fatalf("expected latest %d entries, actual %+v", expected, actual)
	}
}
