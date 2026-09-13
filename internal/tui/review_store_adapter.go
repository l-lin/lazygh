package tui

func (program *Program) updateReviewStore(transition func(reviewStore) reviewStore) {
	if program == nil || program.reviewStore == nil {
		return
	}

	updatedStore := transition(*program.reviewStore)
	program.reviewStore = &updatedStore
}

func (program *Program) setPullRequestDiffStatusOperationID(key string, operationID uint64) {
	program.updateReviewStore(func(store reviewStore) reviewStore {
		if store.pullRequestDiffStatusOperationIDs == nil {
			store.pullRequestDiffStatusOperationIDs = map[string]uint64{}
		}
		store.pullRequestDiffStatusOperationIDs[key] = operationID
		return store
	})
}

func (program *Program) pullRequestDiffStatusOperationID(key string) uint64 {
	if program == nil || program.reviewStore == nil {
		return 0
	}
	return program.reviewStore.pullRequestDiffStatusOperationIDs[key]
}
