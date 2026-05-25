package tui

import "strings"

type keybindingLabelResolver struct {
	actions map[keybindingActionID]resolvedKeybindingAction
}

func newKeybindingLabelResolver(actions []resolvedKeybindingAction) keybindingLabelResolver {
	resolvedActions := make(map[keybindingActionID]resolvedKeybindingAction, len(actions))
	for _, action := range actions {
		resolvedActions[action.action.id] = action
	}
	return keybindingLabelResolver{actions: resolvedActions}
}

func (resolver keybindingLabelResolver) helpKeysOrFallback(fallback string, actionIDs ...keybindingActionID) string {
	if len(actionIDs) == 0 {
		return formatKeyTextForDisplay(fallback)
	}

	actualLabels, ok, hasOverride := resolver.resolvedKeyLabels(actionIDs...)
	if !ok || !hasOverride || len(actualLabels) == 0 {
		return formatKeyTextForDisplay(fallback)
	}

	return strings.Join(formattedKeySequenceLabelsForDisplay(actualLabels), "/")
}

func (resolver keybindingLabelResolver) resolvedKeyLabelsText(actionIDs ...keybindingActionID) string {
	actualLabels, ok, _ := resolver.resolvedKeyLabels(actionIDs...)
	if !ok || len(actualLabels) == 0 {
		return ""
	}
	return strings.Join(formattedKeySequenceLabelsForDisplay(actualLabels), "/")
}

func (resolver keybindingLabelResolver) resolvedKeyLabels(actionIDs ...keybindingActionID) ([]string, bool, bool) {
	if len(actionIDs) == 0 {
		return nil, false, false
	}

	actualLabels := make([]string, 0)
	hasOverride := false
	for _, actionID := range actionIDs {
		action, ok := resolver.actions[actionID]
		if !ok {
			return nil, false, false
		}
		if action.overridden {
			hasOverride = true
		}
		for _, binding := range action.bindings {
			actualLabels = append(actualLabels, binding.label)
		}
	}

	return actualLabels, true, hasOverride
}
