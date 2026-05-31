package tui

func currentCharacterMotionTargetSelectionFor(document detailDocument, state detailViewState) characterMotionTargetSelection {
	state.sync(document, 1)
	return characterMotionTargetSelection{document: document, cursor: state.cursor}
}

func (program *Program) currentCharacterMotionTargetReadModel() characterMotionTargetReadModel {
	if program == nil {
		return characterMotionTargetReadModel{}
	}

	model := characterMotionTargetReadModel{}
	if program.detailState.viewState.hasPendingCharacterMotion() {
		model.detail = currentCharacterMotionTargetSelectionFor(program.currentDetailDocument(nil), program.detailState.viewState)
		model.detailKnown = true
	}
	if program.pullRequestBuildRunPopup != nil && program.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
		model.buildPopup = currentCharacterMotionTargetSelectionFor(program.currentPullRequestBuildRunPopupDocument(nil), program.pullRequestBuildRunPopup.viewState)
		model.buildPopupKnown = true
	}
	return model
}

func (program *Program) currentDetailCharacterMotionTargetRunes() []rune {
	return program.currentCharacterMotionTargetReadModel().detailRunes()
}

func (program *Program) currentPullRequestBuildRunPopupCharacterMotionTargetRunes() []rune {
	return program.currentCharacterMotionTargetReadModel().buildPopupRunes()
}
