package tui

import (
	"errors"

	"github.com/jesseduffield/gocui"
)

const pullRequestBrowserOpenSuccessMessage = "PR opened in browser"

func (program *Program) executeOpenPullRequestInBrowserAction(_ *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if !program.hasPullRequestMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}
	if err := program.pullRequestMutations.OpenPullRequestInBrowser(target.repository, target.number); err != nil {
		return actionsPopupActionResult{err: err}
	}

	program.setFeedback(program.model.Focus(), pullRequestBrowserOpenSuccessMessage)
	return actionsPopupActionResult{closePopup: true}
}
