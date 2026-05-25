package tui

func (program *Program) helpKeysOrFallback(fallback string, actionIDs ...keybindingActionID) string {
	return program.keybindingLabelResolver().helpKeysOrFallback(fallback, actionIDs...)
}

func (program *Program) resolvedKeyLabelsText(actionIDs ...keybindingActionID) string {
	return program.keybindingLabelResolver().resolvedKeyLabelsText(actionIDs...)
}

func (program *Program) resolvedKeyLabels(actionIDs ...keybindingActionID) ([]string, bool, bool) {
	return program.keybindingLabelResolver().resolvedKeyLabels(actionIDs...)
}

func (program *Program) keybindingLabelResolver() keybindingLabelResolver {
	if program == nil {
		return keybindingLabelResolver{}
	}
	return newKeybindingLabelResolver(program.resolvedKeybindingActions())
}
