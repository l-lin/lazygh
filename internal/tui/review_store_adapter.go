package tui

func (program *Program) updateReviewStore(transition func(reviewStore) reviewStore) {
	if program == nil || program.reviewStore == nil {
		return
	}

	updatedStore := transition(*program.reviewStore)
	program.reviewStore = &updatedStore
}
