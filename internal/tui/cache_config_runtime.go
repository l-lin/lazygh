package tui

import (
	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
)

type cacheConfigRuntime struct {
	openPersistentCache       func(string) (persistentPullRequestCache, error)
	openNotificationDoneStore func(string) (notificationDoneStore, error)
	closePersistentCache      func(persistentPullRequestCache) error
}

func newCacheConfigRuntime() cacheConfigRuntime {
	return cacheConfigRuntime{
		openPersistentCache:       openPersistentPullRequestCache,
		openNotificationDoneStore: openPersistentNotificationDoneStore,
		closePersistentCache:      closePersistentPullRequestCache,
	}
}

func executeApplyCacheConfigRuntime(runtime cacheConfigRuntime, previous persistentPullRequestCache, config appconfig.CacheConfig) (cacheConfigAppliedState, error) {
	if config.Path == "" {
		if runtime.closePersistentCache != nil {
			_ = runtime.closePersistentCache(previous)
		}
		return cacheConfigAppliedState{notificationDoneStore: noopNotificationDoneStore{}}, nil
	}

	store, actualErr := runtime.openPersistentCache(config.Path)
	if actualErr != nil {
		return cacheConfigAppliedState{}, actualErr
	}
	doneStore, actualErr := runtime.openNotificationDoneStore(persistcache.NotificationDoneStorePath(config.Path))
	if actualErr != nil {
		if runtime.closePersistentCache != nil {
			_ = runtime.closePersistentCache(store)
		}
		return cacheConfigAppliedState{}, actualErr
	}
	if runtime.closePersistentCache != nil {
		_ = runtime.closePersistentCache(previous)
	}
	return cacheConfigAppliedState{pullRequestCache: store, notificationDoneStore: doneStore}, nil
}

func openPersistentPullRequestCache(path string) (persistentPullRequestCache, error) {
	return persistcache.Open(path)
}

func openPersistentNotificationDoneStore(path string) (notificationDoneStore, error) {
	store, actualErr := persistcache.OpenNotificationDoneStore(path)
	if actualErr != nil {
		return nil, actualErr
	}
	return notificationDoneStoreAdapter{store: store}, nil
}

func closePersistentPullRequestCache(cache persistentPullRequestCache) error {
	if cache == nil {
		return nil
	}
	return cache.Close()
}
