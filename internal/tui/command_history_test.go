package tui

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCommandHistoryStore_GivenCommandsInAnyArrivalOrder_WhenAppendingThem_ThenItReturnsChronologicalEntries(t *testing.T) {
	subject := *newCommandHistoryStore()
	later := time.Date(2026, time.September, 13, 12, 35, 2, 0, time.Local)
	earlier := later.Add(-6 * time.Second)

	subject = subject.withCommandStarted("gh pr view 42", later)
	subject = subject.withCommandStarted("gh api graphql", earlier)

	actual := subject.entriesSnapshot()
	expected := []commandHistoryEntry{
		{command: "gh api graphql", startedAt: earlier},
		{command: "gh pr view 42", startedAt: later},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected chronological entries %+v, actual %+v", expected, actual)
	}
}

func TestCommandHistoryStore_GivenCommandsWithEqualTimestamps_WhenAppendingThem_ThenItPreservesInsertionOrder(t *testing.T) {
	subject := *newCommandHistoryStore()
	startedAt := time.Date(2026, time.September, 13, 12, 35, 2, 0, time.Local)

	subject = subject.withCommandStarted("gh api first", startedAt)
	subject = subject.withCommandStarted("gh api second", startedAt)

	actual := subject.entriesSnapshot()
	if len(actual) != 2 || actual[0].command != "gh api first" || actual[1].command != "gh api second" {
		t.Fatalf("expected stable insertion order, actual %+v", actual)
	}
}

func TestCommandHistoryStore_GivenMoreThanTheRetentionLimit_WhenAppendingCommands_ThenItRetainsTheLatest100Entries(t *testing.T) {
	subject := *newCommandHistoryStore()
	startedAt := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.Local)
	for index := 0; index < maxCommandHistoryEntries+1; index++ {
		subject = subject.withCommandStarted("gh api "+strconv.Itoa(index), startedAt.Add(time.Duration(index)*time.Second))
	}

	actual := subject.entriesSnapshot()
	if len(actual) != maxCommandHistoryEntries {
		t.Fatalf("expected exactly %d entries, actual %d", maxCommandHistoryEntries, len(actual))
	}
	if actual[0].command != "gh api 1" || actual[len(actual)-1].command != "gh api 100" {
		t.Fatalf("expected the latest 100 commands, first %q last %q", actual[0].command, actual[len(actual)-1].command)
	}
}

func TestCommandHistoryStore_GivenAnExistingStore_WhenAppendingACommand_ThenItReturnsAnIndependentSnapshot(t *testing.T) {
	startedAt := time.Date(2026, time.September, 13, 12, 35, 2, 0, time.Local)
	subject := newCommandHistoryStore().withCommandStarted("gh api first", startedAt)
	before := subject.entriesSnapshot()

	updated := subject.withCommandStarted("gh api second", startedAt.Add(time.Second))
	before[0].command = "mutated snapshot"
	actual := subject.entriesSnapshot()

	if len(actual) != 1 || actual[0].command != "gh api first" {
		t.Fatalf("expected the original store to remain unchanged, actual %+v", actual)
	}
	if len(updated.entriesSnapshot()) != 2 {
		t.Fatalf("expected updated store to contain two entries, actual %+v", updated.entriesSnapshot())
	}
}

func TestCommandHistoryStore_GivenBlankCommand_WhenAppendingIt_ThenItIgnoresTheCommand(t *testing.T) {
	actual := newCommandHistoryStore().withCommandStarted("  ", time.Now()).entriesSnapshot()

	if len(actual) != 0 {
		t.Fatalf("expected blank command to be ignored, actual %+v", actual)
	}
}

func TestFormatCommandHistoryEntry_GivenACommandEntry_WhenFormattingIt_ThenItUsesLocalTimeAndTheFullGHCommand(t *testing.T) {
	entry := commandHistoryEntry{
		command:   "gh pr view 42 -R acme/widgets --web",
		startedAt: time.Date(2026, time.September, 13, 12, 34, 56, 0, time.Local),
	}

	actual := formatCommandHistoryEntry(entry)

	if actual != "12:34:56 gh pr view 42 -R acme/widgets --web" {
		t.Fatalf("expected formatted history entry %q, actual %q", "12:34:56 gh pr view 42 -R acme/widgets --web", actual)
	}
}

func TestRenderCommandHistory_GivenEntries_WhenRenderingThem_ThenItJoinsTimestampedCommandsInOrder(t *testing.T) {
	entries := []commandHistoryEntry{
		{command: "gh api graphql", startedAt: time.Date(2026, time.September, 13, 12, 34, 56, 0, time.Local)},
		{command: "gh pr view 42", startedAt: time.Date(2026, time.September, 13, 12, 35, 2, 0, time.Local)},
	}

	actual := renderCommandHistory(entries)
	expected := "12:34:56 gh api graphql\n12:35:02 gh pr view 42"
	if actual != expected {
		t.Fatalf("expected rendered history %q, actual %q", expected, actual)
	}
}

func TestRenderCommandHistory_GivenNoEntries_WhenRenderingThem_ThenItReturnsAnEmptyBody(t *testing.T) {
	if actual := renderCommandHistory(nil); actual != "" {
		t.Fatalf("expected an empty history body, actual %q", actual)
	}
}

func TestUpdate_GivenGHCommandStartedMessages_WhenReducingThem_ThenItStoresChronologicalHistoryOnTheUIThread(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	later := time.Date(2026, time.September, 13, 12, 35, 2, 0, time.Local)

	Update(subject, MsgGHCommandStarted{Command: "gh pr view 42", StartedAt: later})
	Update(subject, MsgGHCommandStarted{Command: "gh api graphql", StartedAt: later.Add(-6 * time.Second)})

	actual := subject.renderCommandHistoryPopupBody()
	expected := "12:34:56 gh api graphql\n12:35:02 gh pr view 42"
	if actual != expected {
		t.Fatalf("expected reduced history body %q, actual %q", expected, actual)
	}
}

func TestProgram_GivenNoCommandHistory_WhenRenderingTheHistoryPopupBody_ThenItReturnsAnEmptyBody(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	if actual := strings.TrimSpace(subject.renderCommandHistoryPopupBody()); actual != "" {
		t.Fatalf("expected an empty history popup body, actual %q", actual)
	}
}
