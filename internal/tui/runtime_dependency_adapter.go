package tui

func (program *Program) setPersistentPullRequestCache(cache persistentPullRequestCache) {
	if program == nil || program.persistentCacheStore == nil {
		return
	}
	program.persistentCacheStore.pullRequestCache = cache
}

func (program *Program) setNotificationDoneStore(doneStore notificationDoneStore) {
	program.updateNotificationStore(func(store notificationStore) notificationStore {
		return store.withNotificationDoneStore(doneStore)
	})
}

func (program *Program) setLinkOpener(opener linkOpener) {
	if program == nil {
		return
	}
	program.linkOpener = opener
}

func configuredSystemLinkOpener(current linkOpener, command []string) linkOpener {
	actual, ok := current.(*systemLinkOpener)
	if !ok {
		return newSystemLinkOpener(command)
	}
	return actual.withCommand(command)
}
