package tui

import (
	"sort"
	"strings"
	"time"
)

const maxCommandHistoryEntries = 100

type commandHistoryEntry struct {
	command   string
	startedAt time.Time
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

	entries := make([]commandHistoryEntry, len(store.entries)+1)
	copy(entries, store.entries)
	entries[len(store.entries)] = commandHistoryEntry{command: trimmedCommand, startedAt: startedAt}
	sort.SliceStable(entries, func(left int, right int) bool {
		return entries[left].startedAt.Before(entries[right].startedAt)
	})
	if len(entries) > maxCommandHistoryEntries {
		entries = entries[len(entries)-maxCommandHistoryEntries:]
	}

	store.entries = entries
	return store
}

func (store commandHistoryStore) entriesSnapshot() []commandHistoryEntry {
	return append([]commandHistoryEntry(nil), store.entries...)
}

func formatCommandHistoryEntry(entry commandHistoryEntry) string {
	return entry.startedAt.Format("15:04:05") + " " + entry.command
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
