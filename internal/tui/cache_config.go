package tui

import (
	persistcache "github.com/l-lin/lazygh/internal/cache"
	appconfig "github.com/l-lin/lazygh/internal/config"
)

func (program *Program) ApplyCacheConfig(config appconfig.CacheConfig) error {
	if program == nil {
		return nil
	}

	applied, actualErr := openCacheConfigAppliedState(config)
	if actualErr != nil {
		return actualErr
	}
	if previous := program.pullRequestCache; previous != nil {
		_ = previous.Close()
	}
	return program.dispatchRuntimeMessage(applied.message())
}

type cacheConfigAppliedState struct {
	pullRequestCache      persistentPullRequestCache
	notificationDoneStore notificationDoneStore
}

func (state cacheConfigAppliedState) message() MsgCacheConfigApplied {
	return MsgCacheConfigApplied{PullRequestCache: state.pullRequestCache, NotificationDoneStore: state.notificationDoneStore}
}

func openCacheConfigAppliedState(config appconfig.CacheConfig) (cacheConfigAppliedState, error) {
	if config.Path == "" {
		return cacheConfigAppliedState{notificationDoneStore: noopNotificationDoneStore{}}, nil
	}

	store, actualErr := persistcache.Open(config.Path)
	if actualErr != nil {
		return cacheConfigAppliedState{}, actualErr
	}
	doneStore, actualErr := persistcache.OpenNotificationDoneStore(persistcache.NotificationDoneStorePath(config.Path))
	if actualErr != nil {
		_ = store.Close()
		return cacheConfigAppliedState{}, actualErr
	}
	return cacheConfigAppliedState{pullRequestCache: store, notificationDoneStore: notificationDoneStoreAdapter{store: doneStore}}, nil
}
