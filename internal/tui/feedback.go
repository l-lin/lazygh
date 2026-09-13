package tui

func (program *Program) setFeedback(_ Focus, message string) {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withFeedback(message)
	})
}

func (program *Program) setFailureFeedback(_ Focus, action string, operationErr error) {
	program.recordStatusLineFailure(action, operationErr)
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withFailureFeedback(formatStatusLineFailure(action, operationErr))
	})
}

func (program *Program) clearFeedbackMessage() {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withoutFeedback()
	})
}
