package tui

func (program *Program) pullRequestCommentHelpEntry(scope string) helpEntry {
	return helpEntry{
		Key:         program.helpKeysOrFallback("c", keybindingActionID{scope: scope, action: "comment_on_pull_request"}),
		Description: "Comment on PR",
	}
}

func (program *Program) detailPullRequestCommentHelpEntry() helpEntry {
	if program.browserChangesInlineCommentShortcutActive() {
		return helpEntry{
			Key:         program.helpKeysOrFallback("c", keybindingActionID{scope: keymapScopePullRequests, action: "comment_on_pull_request"}),
			Description: "Add inline comment",
		}
	}
	return program.pullRequestCommentHelpEntry(keymapScopePullRequests)
}

func (program *Program) inlineCommentReplyHelpEntry() helpEntry {
	return helpEntry{
		Key:         program.helpKeysOrFallback("r", keybindingActionID{scope: keymapScopePullRequests, action: "reply_to_inline_comment"}),
		Description: "Reply to inline comment",
	}
}
