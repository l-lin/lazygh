package tui

import (
	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
)

func (program *Program) ApplyCacheConfig(config appconfig.CacheConfig) error {
	if program.pullRequestCache != nil {
		_ = program.pullRequestCache.Close()
		program.pullRequestCache = nil
	}
	program.notificationDoneStore = noopNotificationDoneStore{}
	if config.Path == "" {
		return nil
	}

	store, actualErr := persistcache.Open(config.Path)
	if actualErr != nil {
		return actualErr
	}
	doneStore, actualErr := persistcache.OpenNotificationDoneStore(persistcache.NotificationDoneStorePath(config.Path))
	if actualErr != nil {
		_ = store.Close()
		return actualErr
	}

	program.pullRequestCache = store
	program.notificationDoneStore = notificationDoneStoreAdapter{store: doneStore}
	return nil
}
