package tui

func (program *Program) updateDetailStore(transition func(detailStore) detailStore) {
	if program == nil || program.detailStore == nil {
		return
	}

	updatedStore := transition(*program.detailStore)
	program.detailStore = &updatedStore
}

func (program *Program) setPullRequestDetailStatusOperationID(key string, operationID uint64) {
	program.updateDetailStore(func(store detailStore) detailStore {
		if store.pullRequestDetailStatusOperationIDs == nil {
			store.pullRequestDetailStatusOperationIDs = map[string]uint64{}
		}
		store.pullRequestDetailStatusOperationIDs[key] = operationID
		return store
	})
}

func (program *Program) pullRequestDetailStatusOperationID(key string) uint64 {
	if program == nil || program.detailStore == nil {
		return 0
	}
	return program.detailStore.pullRequestDetailStatusOperationIDs[key]
}

func (program *Program) setIssueDetailStatusOperationID(key string, operationID uint64) {
	program.updateDetailStore(func(store detailStore) detailStore {
		if store.issueDetailStatusOperationIDs == nil {
			store.issueDetailStatusOperationIDs = map[string]uint64{}
		}
		store.issueDetailStatusOperationIDs[key] = operationID
		return store
	})
}

func (program *Program) issueDetailStatusOperationID(key string) uint64 {
	if program == nil || program.detailStore == nil {
		return 0
	}
	return program.detailStore.issueDetailStatusOperationIDs[key]
}

func (program *Program) setReleaseDetailStatusOperationID(key string, operationID uint64) {
	program.updateDetailStore(func(store detailStore) detailStore {
		if store.releaseDetailStatusOperationIDs == nil {
			store.releaseDetailStatusOperationIDs = map[string]uint64{}
		}
		store.releaseDetailStatusOperationIDs[key] = operationID
		return store
	})
}

func (program *Program) releaseDetailStatusOperationID(key string) uint64 {
	if program == nil || program.detailStore == nil {
		return 0
	}
	return program.detailStore.releaseDetailStatusOperationIDs[key]
}
