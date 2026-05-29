package tui

func pullRequestBrowserHelpEntry(resolver keybindingLabelResolver, scope string) helpEntry {
	return helpEntry{
		Key:         resolver.helpKeysOrFallback("alt+b", keybindingActionID{scope: scope, action: "open_pull_request_in_browser"}),
		Description: "Open PR in browser",
	}
}
