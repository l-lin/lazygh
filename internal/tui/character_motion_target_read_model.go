package tui

type characterMotionTargetSelection struct {
	document detailDocument
	cursor   detailPosition
}

type characterMotionTargetReadModel struct {
	detail          characterMotionTargetSelection
	detailKnown     bool
	buildPopup      characterMotionTargetSelection
	buildPopupKnown bool
}

func (model characterMotionTargetReadModel) detailRunes() []rune {
	if !model.detailKnown {
		return nil
	}
	return currentCharacterMotionTargetRunesAtCursor(model.detail.document, model.detail.cursor)
}

func (model characterMotionTargetReadModel) buildPopupRunes() []rune {
	if !model.buildPopupKnown {
		return nil
	}
	return currentCharacterMotionTargetRunesAtCursor(model.buildPopup.document, model.buildPopup.cursor)
}
