package tui

import (
	"errors"
	"strings"
)

type transientErrorPopupActionError struct {
	err error
}

func newTransientErrorPopupActionError(err error) error {
	if err == nil {
		return nil
	}

	var popupErr transientErrorPopupActionError
	if errors.As(err, &popupErr) {
		return err
	}

	normalizedErr := normalizeGHCommandError(err)
	if normalizedErr == nil {
		return nil
	}
	return transientErrorPopupActionError{err: normalizedErr}
}

func (err transientErrorPopupActionError) Error() string {
	if err.err == nil {
		return ""
	}
	return err.err.Error()
}

func (err transientErrorPopupActionError) Unwrap() error {
	return err.err
}

func transientErrorPopupActionMessage(err error) (string, bool) {
	var popupErr transientErrorPopupActionError
	if !errors.As(err, &popupErr) || popupErr.err == nil {
		return "", false
	}

	message := strings.TrimSpace(popupErr.err.Error())
	return message, message != ""
}

func normalizeGHCommandError(err error) error {
	if err == nil {
		return nil
	}

	message := strings.TrimSpace(err.Error())
	if strings.HasPrefix(message, "run `") {
		if separatorIndex := strings.Index(message, ":"); separatorIndex >= 0 {
			message = strings.TrimSpace(message[separatorIndex+1:])
		}
	}
	if strings.HasPrefix(message, "exit status ") {
		if separatorIndex := strings.Index(message, ":"); separatorIndex >= 0 {
			message = strings.TrimSpace(message[separatorIndex+1:])
		}
	}
	if message == "" {
		return err
	}
	return errors.New(message)
}
