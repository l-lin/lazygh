package tui

func reviewInlineCommentHelpEntry(resolver keybindingLabelResolver) helpEntry {
	return helpEntry{
		Key:         resolver.helpKeysOrFallback("c", keybindingActionID{scope: keymapScopePullRequests, action: "comment_on_pull_request"}),
		Description: "Add inline comment",
	}
}
