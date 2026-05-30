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

func (program *Program) updateSystemLinkOpenerCommand(command []string) bool {
	if program == nil {
		return false
	}
	actual, ok := program.linkOpener.(*systemLinkOpener)
	if !ok {
		return false
	}
	actual.command = append([]string(nil), command...)
	return true
}
