package tui

import "github.com/jesseduffield/gocui"

const (
	yankBranchSuccessMessage     = iconCopy + " Branch copied"
	yankBranchUnavailableMessage = iconUnavailable + " No PR branch"
)

func (program *Program) copyPullRequestBranch(gui *gocui.Gui, view *gocui.View) error {
	return program.dispatch(gui, MsgCopyPullRequestBranchRequested{})
}

func pullRequestBranchYankHelpEntry(resolver keybindingLabelResolver, scope string) helpEntry {
	return helpEntry{
		Key:         resolver.helpKeysOrFallback("alt+y", keybindingActionID{scope: scope, action: "copy_pull_request_branch"}),
		Description: "Copy PR branch",
	}
}
