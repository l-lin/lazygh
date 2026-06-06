package tui

import (
	"errors"
	"strings"
)

type actionsPopupAction struct {
	id        string
	group     string
	title     string
	icon      string
	keywords  []string
	requested Msg
}

func (action actionsPopupAction) withGroup(group string) actionsPopupAction {
	action.group = strings.TrimSpace(group)
	return action
}

func (action actionsPopupAction) label() string {
	if strings.TrimSpace(action.icon) == "" {
		return action.title
	}
	return action.icon + " " + action.title
}

func actionsPopupErrorRequested(err error) Msg {
	return MsgActionsPopupActionErrorHandled{Err: err}
}

var errActionsPopupActionUnavailable = errors.New("action is unavailable")
