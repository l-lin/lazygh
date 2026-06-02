package tui

func (program *Program) activeDetailCharacterMotionTargetBindingSpecs() []keybindingSpec {
	bindings := program.currentDetailCharacterMotionTargetRunes()
	if len(bindings) == 0 {
		return nil
	}

	specs := make([]keybindingSpec, 0, len(bindings))
	for _, target := range bindings {
		specs = append(specs, keybindingSpec{viewName: viewDetailName, key: keybindingValueForRune(target), handler: program.detailCharacterMotionTargetHandler(target)})
	}
	return specs
}

func (program *Program) activePullRequestBuildRunPopupCharacterMotionTargetBindingSpecs() []keybindingSpec {
	bindings := program.currentPullRequestBuildRunPopupCharacterMotionTargetRunes()
	if len(bindings) == 0 {
		return nil
	}

	specs := make([]keybindingSpec, 0, len(bindings))
	for _, target := range bindings {
		specs = append(specs, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: keybindingValueForRune(target), handler: program.pullRequestBuildRunPopupCharacterMotionTargetHandler(target)})
	}
	return specs
}
