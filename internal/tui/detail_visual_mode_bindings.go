package tui

import "github.com/jesseduffield/gocui"

func (program *Program) activeDetailVisualModeOtherEndBindingSpecs() []keybindingSpec {
	if program == nil || !program.detailState.viewState.mode.isVisual() || program.detailState.viewState.hasPendingCharacterMotion() {
		return nil
	}

	return program.visualModeOtherEndBindingSpecs(
		viewDetailName,
		keybindingActionID{scope: keymapScopeCursor, action: "move_cursor_to_other_end"},
		program.moveDetailCursorToOtherEnd,
	)
}

func (program *Program) activePullRequestBuildRunPopupVisualModeOtherEndBindingSpecs() []keybindingSpec {
	if program == nil || program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.mode.isVisual() || program.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
		return nil
	}

	return program.visualModeOtherEndBindingSpecs(
		viewPullRequestBuildInfoName,
		keybindingActionID{scope: keymapScopePullRequestBuildInfo, action: "move_cursor_to_other_end"},
		program.movePullRequestBuildRunPopupCursorToOtherEnd,
	)
}

func (program *Program) visualModeOtherEndBindingSpecs(viewName string, actionID keybindingActionID, handler func(*gocui.Gui, *gocui.View) error) []keybindingSpec {
	specs := make([]keybindingSpec, 0, 1)
	for _, binding := range program.resolvedBindingsForActionID(actionID) {
		if len(binding.keys) != 1 {
			continue
		}
		specs = append(specs, keybindingSpec{viewName: viewName, key: binding.keys[0].value, mod: binding.keys[0].mod, handler: handler})
	}
	return specs
}
