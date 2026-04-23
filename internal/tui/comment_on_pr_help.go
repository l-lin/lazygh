package tui

func (program *Program) pullRequestCommentHelpEntry(scope string) helpEntry {
	return helpEntry{
		Key:         program.helpKeysOrFallback("c", keybindingActionID{scope: scope, action: "comment_on_pull_request"}),
		Description: "Comment on PR",
	}
}
