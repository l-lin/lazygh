package tui

import (
	"path/filepath"
	"reflect"
	"testing"

	persistcache "github.com/l-lin/lazygh/internal/cache"
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

func TestExecuteApplyCacheConfigRuntime_GivenOpenedStoresAndAPreviousCache_WhenApplying_ThenItOpensTheNewStoresBeforeClosingThePreviousCache(t *testing.T) {
	previousCache := &closingPersistentPullRequestCache{}
	openedCache := &closingPersistentPullRequestCache{}
	callLog := []string{}
	config := appconfig.CacheConfig{Path: filepath.Join(t.TempDir(), "lazygh", "cache.sqlite3")}

	actual, actualErr := executeApplyCacheConfigRuntime(cacheConfigRuntime{
		openPersistentCache: func(path string) (persistentPullRequestCache, error) {
			callLog = append(callLog, "open-cache:"+path)
			return openedCache, nil
		},
		openNotificationDoneStore: func(path string) (notificationDoneStore, error) {
			callLog = append(callLog, "open-done:"+path)
			return recordingNotificationDoneStore{}, nil
		},
		closePersistentCache: func(cache persistentPullRequestCache) error {
			callLog = append(callLog, "close-previous")
			return cache.Close()
		},
	}, previousCache, config)

	then_noError(t, actualErr)
	if actual.pullRequestCache != openedCache {
		t.Fatalf("expected the opened cache %p, actual %p", openedCache, actual.pullRequestCache)
	}
	if _, ok := actual.notificationDoneStore.(recordingNotificationDoneStore); !ok {
		t.Fatalf("expected the runtime notification done store, actual %T", actual.notificationDoneStore)
	}
	expectedCallLog := []string{"open-cache:" + config.Path, "open-done:" + persistcache.NotificationDoneStorePath(config.Path), "close-previous"}
	if !reflect.DeepEqual(callLog, expectedCallLog) {
		t.Fatalf("expected cache-config runtime call log %v, actual %v", expectedCallLog, callLog)
	}
	if actual := previousCache.closeCalls; actual != 1 {
		t.Fatalf("expected the previous cache to close once, actual %d", actual)
	}
	if actual := openedCache.closeCalls; actual != 0 {
		t.Fatalf("expected the newly opened cache to stay open, actual close calls %d", actual)
	}
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

func TestUpdate_GivenMsgLinksConfigAppliedAndExistingSystemOpener_WhenApplying_ThenItReplacesTheOpenerWithAnUpdatedCopy(t *testing.T) {
	startHook := func(string, ...string) error { return nil }
	original := &systemLinkOpener{command: []string{"open", "--foreground"}, start: startHook}
	subject := NewProgramWithModel(given_model())
	subject.linkOpener = original
	given_config := appconfig.LinksConfig{OpenCommand: []string{"custom-open", "--background"}}

	Update(subject, MsgLinksConfigApplied{Config: given_config})

	actual, ok := subject.linkOpener.(*systemLinkOpener)
	if !ok {
		t.Fatalf("expected a system link opener, actual %T", subject.linkOpener)
	}
	if actual == original {
		t.Fatal("expected links config apply to replace the system opener instead of mutating it in place")
	}
	expectedCommand := []string{"custom-open", "--background"}
	if !reflect.DeepEqual(actual.command, expectedCommand) {
		t.Fatalf("expected system opener command %v, actual %v", expectedCommand, actual.command)
	}
	if !reflect.DeepEqual(original.command, []string{"open", "--foreground"}) {
		t.Fatalf("expected the original command %v to stay untouched, actual %v", []string{"open", "--foreground"}, original.command)
	}
	if actual.start == nil || reflect.ValueOf(actual.start).Pointer() != reflect.ValueOf(startHook).Pointer() {
		t.Fatal("expected the updated system opener to keep the existing start hook")
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

func TestApplyCacheConfig_GivenPersistentCacheContainsPastedPullRequests_WhenApplying_ThenItHydratesThePastedTabInSavedOrder(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	cachePath := filepath.Join(t.TempDir(), "lazygh", "cache.sqlite3")
	store, actualErr := persistcache.Open(cachePath)
	then_noError(t, actualErr)
	defer func() {
		then_noError(t, store.Close())
	}()
	then_noError(t, store.SavePullRequests(pastedPullRequestsPersistentSearch(), []githubdomain.PullRequest{
		{
			Title:      "Newest pasted PR",
			Number:     77,
			Repository: githubdomain.Repository{NameWithOwner: "acme/rocket"},
			URL:        "https://github.com/acme/rocket/pull/77",
			Body:       "Body 77",
			State:      "OPEN",
		},
		{
			Title:      "Earlier pasted PR",
			Number:     13,
			Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/13",
			Body:       "Body 13",
			State:      "OPEN",
		},
	}))
	subject.ApplyPullRequestSearches(nil)

	actualErr = subject.ApplyCacheConfig(appconfig.CacheConfig{Path: cachePath})
	then_noError(t, actualErr)

	expectedLabels := []string{"My PRs", "My reviews", "Requested", "Pasted (2)"}
	if actual := subject.pullRequestsTabLabels(); !reflect.DeepEqual(actual, expectedLabels) {
		t.Fatalf("expected pull request tab labels %v, actual %v", expectedLabels, actual)
	}
	actualRows := subject.model.PullRequestRows(PullRequestTab(len(expectedLabels) - 1))
	if len(actualRows) != 2 {
		t.Fatalf("expected two pasted pull request rows, actual %+v", actualRows)
	}
	if actualRows[0].Summary == nil || actualRows[0].Summary.Title != "Newest pasted PR" || actualRows[0].Summary.Number != 77 {
		t.Fatalf("expected the newest cached pasted pull request first, actual %+v", actualRows[0])
	}
	if actualRows[1].Summary == nil || actualRows[1].Summary.Title != "Earlier pasted PR" || actualRows[1].Summary.Number != 13 {
		t.Fatalf("expected the earlier cached pasted pull request second, actual %+v", actualRows[1])
	}
}
