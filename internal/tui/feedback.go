package tui

func (program *Program) setFeedback(_ Focus, message string) {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withFeedback(message)
	})
}

func (program *Program) clearFeedbackMessage() {
	program.updateStatusStore(func(store statusStore) statusStore {
		return store.withoutFeedback()
	})
}
