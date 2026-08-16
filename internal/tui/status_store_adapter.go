package tui

func (program *Program) updateStatusStore(transition func(statusStore) statusStore) {
	if program == nil || program.statusStore == nil {
		return
	}

	updatedStore := transition(*program.statusStore)
	program.statusStore = &updatedStore
}

func (program *Program) startStoryReviewLoading(command string) {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withStoryReviewLoadingStarted(command)
	})
}

func (program *Program) finishStoryReviewLoading() {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withStoryReviewLoadingFinished()
	})
}
