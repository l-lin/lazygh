package tui

func (program *Program) updateSessionStore(transition func(sessionStore) sessionStore) {
	if program == nil || program.sessionStore == nil {
		return
	}

	updatedStore := transition(*program.sessionStore)
	program.sessionStore = &updatedStore
}

func (program *Program) planConnectedUserLoad() {
	program.updateSessionStore(func(store sessionStore) sessionStore {
		return store.withLoadPlanned()
	})
}

func (program *Program) setConnectedUser(login string, name string) bool {
	if program == nil || program.sessionStore == nil {
		return false
	}

	previousLogin := program.connectedUserLogin
	program.updateSessionStore(func(store sessionStore) sessionStore {
		return store.withConnectedUser(login, name)
	})
	return previousLogin != program.connectedUserLogin
}
