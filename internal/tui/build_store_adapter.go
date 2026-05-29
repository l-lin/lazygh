package tui

func (program *Program) updateBuildStore(transition func(buildStore) buildStore) {
	if program == nil || program.buildStore == nil {
		return
	}

	updatedStore := transition(*program.buildStore)
	program.buildStore = &updatedStore
}

func (program *Program) startPullRequestBuildRunLoad(command string) {
	program.updateBuildStore(func(store buildStore) buildStore {
		return store.withBuildRunLoadStarted(command)
	})
}

func (program *Program) startPullRequestBuildRunJobLogLoad(command string) {
	program.updateBuildStore(func(store buildStore) buildStore {
		return store.withJobLogLoadStarted(command)
	})
}

func (program *Program) clearPullRequestBuildRunLoad() {
	program.updateBuildStore(func(store buildStore) buildStore {
		return store.withLoadCleared()
	})
}

func (program *Program) closePullRequestBuildRunPopupState() {
	program.updateBuildStore(func(store buildStore) buildStore {
		return store.withPopupClosed()
	})
}
