package tui

import (
	"fmt"

	"github.com/jesseduffield/gocui"
)

const pullRequestBrowserOpenSuccessMessage = "PR opened in browser"

func openPullRequestInBrowserCommand(repository string, number int) string {
	return formatStatusLineCommand("gh", "pr", "view", fmt.Sprintf("%d", number), "-R", repository, "--web")
}

func (program *Program) openPullRequestInBrowserShortcut(gui *gocui.Gui, _ *gocui.View) error {
	return program.dispatch(gui, MsgOpenPullRequestInBrowserShortcutRequested{})
}

func (program *Program) pullRequestBrowserShortcutAvailable() bool {
	if program == nil {
		return false
	}
	_, ok := program.selectedPullRequestActionTarget()
	return ok
}
