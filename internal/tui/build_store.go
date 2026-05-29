package tui

import "strings"

func (store buildStore) withBuildRunLoadStarted(command string) buildStore {
	store.pullRequestBuildRunPopup = nil
	return store.withLoadStarted(command)
}

func (store buildStore) withJobLogLoadStarted(command string) buildStore {
	return store.withLoadStarted(command)
}

func (store buildStore) withLoadStarted(command string) buildStore {
	store.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: strings.TrimSpace(command)}
	return store
}

func (store buildStore) withLoadCleared() buildStore {
	store.pullRequestBuildRunLoad = nil
	return store
}

func (store buildStore) withPopupOpened(content pullRequestBuildRunPopupContent) buildStore {
	store.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(content)
	return store
}

func (store buildStore) withPopupClosed() buildStore {
	if popup := store.pullRequestBuildRunPopup; popup != nil && popup.previousPopup != nil {
		store.pullRequestBuildRunPopup = popup.previousPopup
		return store
	}
	store.pullRequestBuildRunPopup = nil
	return store
}
