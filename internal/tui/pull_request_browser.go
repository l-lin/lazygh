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
	if err := program.dispatch(gui, MsgOpenPullRequestInBrowserRequested{Target: target}); err != nil {
		return actionsPopupActionResult{err: err}
	}
	return actionsPopupActionResult{}
}

func openPullRequestInBrowserCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "view", fmt.Sprintf("%d", number), "-R", repository, "--web")
}
