package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

type actionsPopupAction struct {
	id       string
	group    string
	title    string
	icon     string
	keywords []string
	execute  func(*gocui.Gui) actionsPopupActionResult
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

type actionsPopupActionResult struct {
	closePopup      bool
	err             error
	feedbackMessage string
	feedbackTarget  Focus
}

func actionsPopupExecuteErr(execute func(*gocui.Gui) error) func(*gocui.Gui) actionsPopupActionResult {
	return func(gui *gocui.Gui) actionsPopupActionResult {
		return actionsPopupActionResultFromError(execute(gui))
	}
}

func actionsPopupActionResultFromError(err error) actionsPopupActionResult {
	if err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{}
}

var errActionsPopupActionUnavailable = errors.New("action is unavailable")
