package tui

import (
	"errors"
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
		id:       "clear-cache",
		title:    clearCacheActionTitle,
		icon:     iconDelete,
		keywords: []string{"cache", "clear", "cleanup", "wipe", "reset"},
		execute:  program.executeClearCacheAction,
	}
}

func (program *Program) executeClearCacheAction(_ *gocui.Gui) actionsPopupActionResult {
	if program.pullRequestCache == nil {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if strings.TrimSpace(program.actionsPopupPendingConfirmationActionID) != clearCacheActionTitle {
		program.actionsPopupPendingConfirmationActionID = clearCacheActionTitle
		program.actionsPopupErrorMessage = ""
		return actionsPopupActionResult{}
	}

	program.clearActionsPopupPendingConfirmation()
	if err := program.clearPullRequestCaches(); err != nil {
		return actionsPopupActionResult{err: err}
	}
	program.setFeedback(program.model.Focus(), clearCacheSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}

func (program *Program) clearPullRequestCaches() error {
	if program.pullRequestCache == nil {
		return errors.New("persistent cache is unavailable")
	}
	if err := program.pullRequestCache.Clear(); err != nil {
		return err
	}

	program.pullRequestDetailCache = map[string]pullRequestDetailResult{}
	program.pullRequestDetailLoadInFlight = map[string]bool{}
	program.pullRequestDiffCache = map[string]pullRequestDiffResult{}
	program.pullRequestDiffLoadInFlight = map[string]bool{}
	program.invalidatePullRequestDetailDocumentCache()
	program.invalidateReviewDiffRenderCache()
	program.resetPullRequestSearchState()
	program.model.SetPullRequestTabs(pullRequestTabSeedsForSearches(program.pullRequestSearches))
	return nil
}

func (program *Program) clearActionsPopupPendingConfirmation() {
	program.actionsPopupPendingConfirmationActionID = ""
}

func (program *Program) actionsPopupConfirmationMessage() string {
	if strings.TrimSpace(program.actionsPopupPendingConfirmationActionID) == clearCacheActionTitle {
		return clearCacheConfirmationPromptMessage
	}
	return ""
}
