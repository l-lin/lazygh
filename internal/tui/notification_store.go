package tui

import "strings"

func (store notificationStore) withLoadPlanned() notificationStore {
	store.notificationsLoadStarted = true
	store.notificationsLoading = true
	store.notificationsLoadingDetailMessage = notificationsLoadingDetail
	return store
}

func (store notificationStore) withMutationStarted(message string) notificationStore {
	store.notificationsLoading = true
	store.notificationsLoadingDetailMessage = strings.TrimSpace(message)
	return store
}

func (store notificationStore) withLoadingFinished() notificationStore {
	store.notificationsLoading = false
	store.notificationsLoadingDetailMessage = ""
	return store
}

func (store notificationStore) withLoadStateReset() notificationStore {
	store.notificationsLoadStarted = false
	store.notificationsLoading = false
	store.notificationsLoadingDetailMessage = ""
	return store
}

func (store notificationStore) withNotificationDoneStore(doneStore notificationDoneStore) notificationStore {
	if doneStore == nil {
		doneStore = noopNotificationDoneStore{}
	}
	store.notificationDoneStore = doneStore
	return store
}
