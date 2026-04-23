package tui

func (program *Program) pullRequestYankHelpEntry(scope string) helpEntry {
	return helpEntry{
		Key:         program.helpKeysOrFallback("y", keybindingActionID{scope: scope, action: "copy_pull_request_url"}),
		Description: "Copy PR URL",
	}
}
