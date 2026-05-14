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

func (action actionsPopupAction) withGroup(group string) actionsPopupAction {
	action.group = strings.TrimSpace(group)
	return action
}

func (action actionsPopupAction) withKeywords(keywords ...string) actionsPopupAction {
	action.keywords = append([]string(nil), keywords...)
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

var errActionsPopupActionUnavailable = errors.New("action is unavailable")
