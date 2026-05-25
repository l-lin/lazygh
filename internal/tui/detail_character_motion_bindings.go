package tui

func (program *Program) activeDetailCharacterMotionTargetBindingSpecs() []keybindingSpec {
	if !program.detailState.viewState.hasPendingCharacterMotion() {
		return nil
	}

	actualView := program.resolveView(program.gui, nil, viewDetailName)
	document := program.currentDetailDocument(actualView)
	program.syncDetailViewState(document, viewPageSize(actualView))
	bindings := characterMotionTargetRunes(detailDocumentLineAt(document, program.detailState.viewState.cursor.line))
	if len(bindings) == 0 {
		return nil
	}

	specs := make([]keybindingSpec, 0, len(bindings))
	for _, target := range bindings {
		specs = append(specs, keybindingSpec{viewName: viewDetailName, key: target, handler: program.detailCharacterMotionTargetHandler(target)})
	}
	return specs
}

func (program *Program) activePullRequestBuildRunPopupCharacterMotionTargetBindingSpecs() []keybindingSpec {
	popup := program.pullRequestBuildRunPopup
	if popup == nil || !popup.viewState.hasPendingCharacterMotion() {
		return nil
	}

	actualView := program.resolveView(program.gui, nil, viewPullRequestBuildInfoName)
	document := program.currentPullRequestBuildRunPopupDocument(actualView)
	program.syncPullRequestBuildRunPopupViewState(document, viewPageSize(actualView))
	bindings := characterMotionTargetRunes(detailDocumentLineAt(document, popup.viewState.cursor.line))
	if len(bindings) == 0 {
		return nil
	}

	specs := make([]keybindingSpec, 0, len(bindings))
	for _, target := range bindings {
		specs = append(specs, keybindingSpec{viewName: viewPullRequestBuildInfoName, key: target, handler: program.pullRequestBuildRunPopupCharacterMotionTargetHandler(target)})
	}
	return specs
}
