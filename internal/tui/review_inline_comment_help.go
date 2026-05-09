package tui

func (program *Program) reviewInlineCommentHelpEntry() helpEntry {
	return helpEntry{
		Key:         program.helpKeysOrFallback("c", keybindingActionID{scope: keymapScopePullRequests, action: "comment_on_pull_request"}),
		Description: "Add inline comment",
	}
}
