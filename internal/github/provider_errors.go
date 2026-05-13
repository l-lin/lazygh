package github

import "errors"

var (
	ErrUnavailable        = errors.New("gh is unavailable")
	ErrUnauthenticated    = errors.New("gh is not authenticated")
	ErrEmptyConnectedUser = errors.New("empty connected user response")
)
