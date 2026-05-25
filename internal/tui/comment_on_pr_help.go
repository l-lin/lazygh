package tui

func pullRequestCommentHelpEntry(resolver keybindingLabelResolver, scope string) helpEntry {
	return helpEntry{
		Key:         resolver.helpKeysOrFallback("c", keybindingActionID{scope: scope, action: "comment_on_pull_request"}),
		Description: "Comment on PR",
	}
}

func detailPullRequestCommentHelpEntry(resolver keybindingLabelResolver, browserChangesInlineCommentShortcutActive bool) helpEntry {
	if browserChangesInlineCommentShortcutActive {
		return helpEntry{
			Key:         resolver.helpKeysOrFallback("c", keybindingActionID{scope: keymapScopePullRequests, action: "comment_on_pull_request"}),
			Description: "Add inline comment",
		}
	}
	return pullRequestCommentHelpEntry(resolver, keymapScopePullRequests)
}

func inlineCommentReplyHelpEntry(resolver keybindingLabelResolver) helpEntry {
	return helpEntry{
		Key:         resolver.helpKeysOrFallback("r", keybindingActionID{scope: keymapScopePullRequests, action: "reply_to_inline_comment"}),
		Description: "Reply to inline comment",
	}
}
