package tui

func pullRequestYankHelpEntry(resolver keybindingLabelResolver, scope string) helpEntry {
	return helpEntry{
		Key:         resolver.helpKeysOrFallback("alt+y", keybindingActionID{scope: scope, action: "copy_pull_request_url"}),
		Description: "Copy PR URL",
	}
}
