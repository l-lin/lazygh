package tui

import (
	"strings"

	"github.com/jesseduffield/gocui"
)

const (
	clearCacheActionTitle               = "Clear cache"
	clearCacheConfirmationPromptMessage = "Press Enter again to clear cache"
	clearCacheSuccessMessage            = "Cache cleared."
)

func (program *Program) clearCacheActionsPopupAction() actionsPopupAction {
	return actionsPopupAction{
		id:      "clear-cache",
		title:   clearCacheActionTitle,
		icon:    iconDelete,
		execute: program.executeClearCacheAction,
	}
}

func (program *Program) executeClearCacheAction(gui *gocui.Gui) actionsPopupActionResult {
	if program.pullRequestCache == nil {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if err := program.dispatch(gui, MsgClearCacheRequested{}); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{}
}

func (program *Program) actionsPopupConfirmationMessage() string {
	switch strings.TrimSpace(program.actionsPopupWidget.pendingConfirmationActionID) {
	case clearCacheActionTitle:
		return clearCacheConfirmationPromptMessage
	case squashMergePullRequestActionTitle:
		return squashMergePullRequestConfirmationPromptMessage
	default:
		return ""
	}
}
