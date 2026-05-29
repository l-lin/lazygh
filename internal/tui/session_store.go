package tui

import "strings"

func (store sessionStore) withLoadPlanned() sessionStore {
	store.connectedUserLoadStarted = true
	return store
}

func (store sessionStore) withConnectedUser(login string, name string) sessionStore {
	store.connectedUserLogin = strings.TrimSpace(login)
	store.connectedUserName = strings.TrimSpace(name)
	return store
}
