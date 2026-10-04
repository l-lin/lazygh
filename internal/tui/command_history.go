package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const maxCommandHistoryEntries = 100

type commandHistoryEntry struct {
	command   string
	startedAt time.Time
	activity  githubdomain.CommandActivity
}

type commandHistoryStore struct {
	entries []commandHistoryEntry
}

func newCommandHistoryStore() *commandHistoryStore {
	return &commandHistoryStore{}
}

func (store commandHistoryStore) withCommandStarted(command string, startedAt time.Time) commandHistoryStore {
	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		return store
	}

	return store.withEntry(commandHistoryEntry{command: trimmedCommand, startedAt: startedAt})
}

func (store commandHistoryStore) withEntry(entry commandHistoryEntry) commandHistoryStore {
	entries := make([]commandHistoryEntry, len(store.entries)+1)
	copy(entries, store.entries)
	entries[len(store.entries)] = entry
	sort.SliceStable(entries, func(left int, right int) bool {
		return entries[left].startedAt.Before(entries[right].startedAt)
	})
	if len(entries) > maxCommandHistoryEntries {
		entries = entries[len(entries)-maxCommandHistoryEntries:]
	}

	store.entries = entries
	return store
}

func (store commandHistoryStore) withActivity(activity githubdomain.CommandActivity) commandHistoryStore {
	if strings.TrimSpace(activity.Command) == "" || activity.ID == "" {
		return store
	}
	entries := store.entriesSnapshot()
	for index, entry := range entries {
		if entry.activity.ID != activity.ID {
			continue
		}
		// Worker delivery order must not turn a completed execution back into a running one.
		if entry.activity.Outcome != githubdomain.CommandRunning {
			return store
		}
		entries[index].activity = activity
		store.entries = entries
		return store
	}
	return store.withEntry(commandHistoryEntry{
		command: strings.TrimSpace(activity.Command), startedAt: activity.StartedAt, activity: activity,
	})
}

func (store commandHistoryStore) entriesSnapshot() []commandHistoryEntry {
	return append([]commandHistoryEntry(nil), store.entries...)
}

func formatCommandHistoryEntry(entry commandHistoryEntry) string {
	line := entry.startedAt.Format("15:04:05") + " " + entry.command
	if entry.activity.ID == "" {
		return line
	}
	status := "running"
	switch entry.activity.Outcome {
	case githubdomain.CommandSucceeded:
		status = "success"
	case githubdomain.CommandFailed:
		status = "failure"
	}
	fields := []string{status}
	if entry.activity.Outcome != githubdomain.CommandRunning {
		fields = append(fields, entry.activity.FinishedAt.Sub(entry.activity.StartedAt).Round(time.Millisecond).String())
	}
	if entry.activity.Owner != "" {
		fields = append(fields, fmt.Sprintf("queue=%s", entry.activity.QueueWait.Round(time.Millisecond)), entry.activity.Owner)
	}
	return line + " [" + strings.Join(fields, " · ") + "]"
}

func renderCommandHistory(entries []commandHistoryEntry) string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, formatCommandHistoryEntry(entry))
	}
	return strings.Join(lines, "\n")
}

func (program *Program) renderCommandHistoryPopupBody() string {
	if program == nil || program.commandHistoryStore == nil {
		return ""
	}
	return renderCommandHistory(program.commandHistoryStore.entriesSnapshot())
}
