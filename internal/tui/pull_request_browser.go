package tui

import (
	"errors"
	"fmt"

	"github.com/jesseduffield/gocui"
)

const pullRequestBrowserOpenSuccessMessage = "PR opened in browser"

func (program *Program) executeOpenPullRequestInBrowserAction(gui *gocui.Gui) error {
	target, ok := program.selectedPullRequestActionTarget()
	if !ok {
		return errActionsPopupActionUnavailable
	}
	if !program.hasPullRequestMutations() {
		return errors.New("github loader is unavailable")
	}
	return program.dispatch(gui, MsgOpenPullRequestInBrowserRequested{Target: target})
}

func openPullRequestInBrowserCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "view", fmt.Sprintf("%d", number), "-R", repository, "--web")
}
