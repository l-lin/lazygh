package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

type actionsPopupAction struct {
	id        string
	group     string
	title     string
	icon      string
	keywords  []string
	requested Msg
}

func (program *Program) closeActionsPopupIfVisible(gui *gocui.Gui) error {
	if program == nil || program.model == nil || !program.model.ActionsPopupVisible() {
		return nil
	}
	return program.dispatch(gui, MsgCloseActionsPopup{})
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
