package tui

func (program *Program) updateNotificationStore(transition func(notificationStore) notificationStore) {
	if program == nil || program.notificationStore == nil {
		return
	}

	updatedStore := transition(*program.notificationStore)
	program.notificationStore = &updatedStore
}

func (program *Program) planNotificationsLoad() {
	program.updateNotificationStore(func(store notificationStore) notificationStore {
		return store.withLoadPlanned()
	})
}

func (program *Program) startNotificationMutationLoading(message string) {
	program.updateNotificationStore(func(store notificationStore) notificationStore {
		return store.withMutationStarted(message)
	})
}

func (program *Program) finishNotificationsLoading() {
	program.updateNotificationStore(func(store notificationStore) notificationStore {
		return store.withLoadingFinished()
	})
}

func (program *Program) resetNotificationLoadState() {
	program.updateNotificationStore(func(store notificationStore) notificationStore {
		return store.withLoadStateReset()
	})
}
