package tui

import (
	"path/filepath"
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	githubdomain "github.com/l-lin/lazygh/internal/github"
)

type closingPersistentPullRequestCache struct {
	fakePersistentPullRequestCache
	closeCalls int
}

func (cache *closingPersistentPullRequestCache) Close() error {
	cache.closeCalls++
	return nil
}

type recordingNotificationDoneStore struct{}

func (recordingNotificationDoneStore) FilterNotifications(notifications []githubdomain.Notification) []githubdomain.Notification {
	return append([]githubdomain.Notification(nil), notifications...)
}

func (recordingNotificationDoneStore) HideNotifications([]githubdomain.Notification) error {
	return nil
}

type fakeRuntimeLinkOpener struct{}

func (fakeRuntimeLinkOpener) Open(string) error {
	return nil
}

func TestUpdate_GivenMsgLinksConfigApplied_WhenApplying_ThenItReplacesNonSystemOpenersWithAConfiguredSystemOpener(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.linkOpener = fakeRuntimeLinkOpener{}
	given_config := appconfig.LinksConfig{OpenCommand: []string{"custom-open", "--background"}}

	Update(subject, MsgLinksConfigApplied{Config: given_config})

	actual, ok := subject.linkOpener.(*systemLinkOpener)
	if !ok {
		t.Fatalf("expected a system link opener, actual %T", subject.linkOpener)
	}
	expectedCommand := []string{"custom-open", "--background"}
	if !reflect.DeepEqual(actual.command, expectedCommand) {
		t.Fatalf("expected system opener command %v, actual %v", expectedCommand, actual.command)
	}
}

func TestApplyCacheConfig_GivenExistingCacheAndEmptyPath_WhenApplying_ThenItClosesTheOldCacheAndResetsTheRuntimeStores(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	oldCache := &closingPersistentPullRequestCache{}
	subject.pullRequestCache = oldCache
	subject.notificationDoneStore = recordingNotificationDoneStore{}

	actualErr := subject.ApplyCacheConfig(appconfig.CacheConfig{})

	then_noError(t, actualErr)
	if actual := oldCache.closeCalls; actual != 1 {
		t.Fatalf("expected the previous cache to be closed once, actual %d", actual)
	}
	if subject.pullRequestCache != nil {
		t.Fatalf("expected the persistent cache to be cleared, actual %T", subject.pullRequestCache)
	}
	if _, ok := subject.notificationDoneStore.(noopNotificationDoneStore); !ok {
		t.Fatalf("expected the notification done store to reset to noop, actual %T", subject.notificationDoneStore)
	}
}

func TestApplyCacheConfig_GivenAValidPath_WhenApplying_ThenItOpensThePersistentStoresThroughTheRuntimeSurface(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	cachePath := filepath.Join(t.TempDir(), "lazygh", "cache.sqlite3")

	actualErr := subject.ApplyCacheConfig(appconfig.CacheConfig{Path: cachePath})
	then_noError(t, actualErr)
	if subject.pullRequestCache == nil {
		t.Fatal("expected a persistent pull request cache to be opened")
	}
	defer func() {
		_ = subject.pullRequestCache.Close()
	}()
	if _, ok := subject.notificationDoneStore.(notificationDoneStoreAdapter); !ok {
		t.Fatalf("expected the notification done store adapter to be installed, actual %T", subject.notificationDoneStore)
	}
}
