package tui

func (program *Program) updateDetailStore(transition func(detailStore) detailStore) {
	if program == nil || program.detailStore == nil {
		return
	}

	updatedStore := transition(*program.detailStore)
	program.detailStore = &updatedStore
}
