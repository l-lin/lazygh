package tui

func (program *Program) reviewInlineCommentHelpEntry() helpEntry {
	return helpEntry{
		Key:         program.helpKeysOrFallback("c", keybindingActionID{scope: keymapScopeDetail, action: "comment_on_pull_request"}),
		Description: "Add inline comment",
	}
}
