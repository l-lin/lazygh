package tui

import (
	"errors"
	"fmt"

	"github.com/jesseduffield/gocui"
)

const pullRequestBrowserOpenSuccessMessage = "PR opened in browser"

func (program *Program) executeOpenPullRequestInBrowserAction(gui *gocui.Gui) actionsPopupActionResult {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return actionsPopupActionResult{err: errActionsPopupActionUnavailable}
	}
	if !program.hasPullRequestMutations() {
		return actionsPopupActionResult{err: errors.New("github loader is unavailable")}
	}

	return program.startActionsPopupAsyncGHCommand(gui, openPullRequestInBrowserCommand(target.repository, target.number), func() error {
		return program.pullRequestMutations.OpenPullRequestInBrowser(target.repository, target.number)
	}, func() {
		program.setFeedback(program.model.Focus(), pullRequestBrowserOpenSuccessMessage)
	})
}

func openPullRequestInBrowserCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "view", fmt.Sprintf("%d", number), "-R", repository, "--web")
}
