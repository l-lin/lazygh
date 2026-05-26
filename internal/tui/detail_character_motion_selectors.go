package tui

func currentCharacterMotionTargetRunesAtCursor(document detailDocument, cursor detailPosition) []rune {
	clampedCursor := document.clampPosition(cursor)
	return characterMotionTargetRunes(detailDocumentLineAt(document, clampedCursor.line))
}

func (program *Program) currentDetailCharacterMotionTargetRunes() []rune {
	if program == nil || !program.detailState.viewState.hasPendingCharacterMotion() {
		return nil
	}
	return currentCharacterMotionTargetRunesAtCursor(program.currentDetailDocument(nil), program.detailState.viewState.cursor)
}

func (program *Program) currentPullRequestBuildRunPopupCharacterMotionTargetRunes() []rune {
	if program == nil || program.pullRequestBuildRunPopup == nil || !program.pullRequestBuildRunPopup.viewState.hasPendingCharacterMotion() {
		return nil
	}
	return currentCharacterMotionTargetRunesAtCursor(program.currentPullRequestBuildRunPopupDocument(nil), program.pullRequestBuildRunPopup.viewState.cursor)
}
