package tui

import "strings"

const (
	clearCacheActionTitle               = "Clear cache"
	clearCacheConfirmationPromptMessage = "Press Enter again to clear cache"
	clearCacheSuccessMessage            = "Cache cleared."
)

func (program *Program) clearCacheActionsPopupAction() actionsPopupAction {
	var requested Msg = MsgClearCacheRequested{}
	if program.pullRequestCache == nil {
		requested = actionsPopupErrorRequested(errActionsPopupActionUnavailable)
	}
	return actionsPopupAction{
		id:        "clear-cache",
		title:     clearCacheActionTitle,
		icon:      iconDelete,
		requested: requested,
	}
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
