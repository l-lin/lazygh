package tui

import "strings"

func (store statusStore) withFeedback(message string) statusStore {
	store.feedbackMessage = strings.TrimSpace(message)
	return store
}

func (store statusStore) withoutFeedback() statusStore {
	store.feedbackMessage = ""
	return store
}

func (store statusStore) withGHCommandLoadingStarted(command string) statusStore {
	store.feedbackMessage = ""
	store.ghCommandLoadingMessage = formatRunningCommandStatus(command)
	return store
}

func (store statusStore) withGHCommandLoadingCleared() statusStore {
	store.ghCommandLoadingMessage = ""
	return store
}

func (store statusStore) withStoryReviewLoadingStarted() statusStore {
	store.feedbackMessage = ""
	store.storyReviewLoading = true
	return store
}

func (store statusStore) withStoryReviewLoadingFinished() statusStore {
	store.storyReviewLoading = false
	return store
}
