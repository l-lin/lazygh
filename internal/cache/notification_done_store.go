package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const notificationDoneStoreFileName = "notification-done.json"

type NotificationDoneStore struct {
	path             string
	mutex            sync.Mutex
	hiddenByThreadID map[string]string
}

type notificationDoneStoreEntry struct {
	ThreadID  string `json:"thread_id"`
	UpdatedAt string `json:"updated_at"`
}

func NotificationDoneStorePath(cachePath string) string {
	trimmedCachePath := strings.TrimSpace(cachePath)
	if trimmedCachePath == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(trimmedCachePath), notificationDoneStoreFileName)
}

func OpenNotificationDoneStore(path string) (*NotificationDoneStore, error) {
	store := &NotificationDoneStore{
		path:             strings.TrimSpace(path),
		hiddenByThreadID: map[string]string{},
	}
	if store.path == "" {
		return store, nil
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (store *NotificationDoneStore) HideNotifications(notifications []githubcli.Notification) error {
	if store == nil {
		return nil
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	changed := false
	for _, notification := range notifications {
		threadID := strings.TrimSpace(notification.ID)
		if threadID == "" {
			continue
		}
		updatedAt := strings.TrimSpace(notification.UpdatedAt)
		if existingUpdatedAt, ok := store.hiddenByThreadID[threadID]; ok && compareNotificationUpdatedAt(existingUpdatedAt, updatedAt) >= 0 {
			continue
		}
		store.hiddenByThreadID[threadID] = updatedAt
		changed = true
	}
	if !changed {
		return nil
	}
	return store.persistLocked()
}

func (store *NotificationDoneStore) FilterNotifications(notifications []githubcli.Notification) []githubcli.Notification {
	if store == nil {
		return append([]githubcli.Notification(nil), notifications...)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	filtered := make([]githubcli.Notification, 0, len(notifications))
	pruned := false
	for _, notification := range notifications {
		threadID := strings.TrimSpace(notification.ID)
		hiddenUpdatedAt, ok := store.hiddenByThreadID[threadID]
		if !ok || threadID == "" {
			filtered = append(filtered, notification)
			continue
		}
		if compareNotificationUpdatedAt(strings.TrimSpace(notification.UpdatedAt), hiddenUpdatedAt) > 0 {
			delete(store.hiddenByThreadID, threadID)
			pruned = true
			filtered = append(filtered, notification)
			continue
		}
	}
	if pruned {
		_ = store.persistLocked()
	}
	return filtered
}

func (store *NotificationDoneStore) load() error {
	bytes, err := os.ReadFile(store.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(bytes) == 0 {
		return nil
	}

	var entries []notificationDoneStoreEntry
	if err := json.Unmarshal(bytes, &entries); err != nil {
		return err
	}
	for _, entry := range entries {
		threadID := strings.TrimSpace(entry.ThreadID)
		if threadID == "" {
			continue
		}
		store.hiddenByThreadID[threadID] = strings.TrimSpace(entry.UpdatedAt)
	}
	return nil
}

func (store *NotificationDoneStore) persistLocked() error {
	if store.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(store.path), 0o755); err != nil {
		return err
	}

	entries := make([]notificationDoneStoreEntry, 0, len(store.hiddenByThreadID))
	for threadID, updatedAt := range store.hiddenByThreadID {
		entries = append(entries, notificationDoneStoreEntry{ThreadID: threadID, UpdatedAt: updatedAt})
	}
	sort.Slice(entries, func(left int, right int) bool {
		return entries[left].ThreadID < entries[right].ThreadID
	})

	payload, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	temporaryPath := store.path + ".tmp"
	if err := os.WriteFile(temporaryPath, payload, 0o644); err != nil {
		return err
	}
	return os.Rename(temporaryPath, store.path)
}

func compareNotificationUpdatedAt(left string, right string) int {
	trimmedLeft := strings.TrimSpace(left)
	trimmedRight := strings.TrimSpace(right)
	switch {
	case trimmedLeft == trimmedRight:
		return 0
	case trimmedLeft == "":
		return -1
	case trimmedRight == "":
		return 1
	default:
		return strings.Compare(trimmedLeft, trimmedRight)
	}
}
