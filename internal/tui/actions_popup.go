package tui

import (
	"errors"
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	actionsPopupStartReviewIcon            = ""
	actionsPopupYankPullRequestURLIcon     = "󰆏"
	actionsPopupOpenPullRequestBrowserIcon = ""
	actionsPopupOpenLinkIcon               = ""
	actionsPopupRefreshPullRequestIcon     = ""
	actionsPopupReviewApproveIcon          = "󰆀"
	actionsPopupReviewCommentIcon          = "󰆂"
	actionsPopupReviewRequestChangesIcon   = "󰅾"
	actionsPopupCommentOnPullRequestIcon   = "󰆆"
	actionsPopupEditPullRequestIcon        = ""
)

type actionsPopupAction struct {
	id       string
	title    string
	icon     string
	keywords []string
	execute  func(*gocui.Gui) actionsPopupActionResult
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
