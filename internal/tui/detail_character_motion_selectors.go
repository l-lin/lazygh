package tui

func currentCharacterMotionTargetRunesAtCursor(document detailDocument, cursor detailPosition) []rune {
	clampedCursor := document.clampPosition(cursor)
	return characterMotionTargetRunes(detailDocumentLineAt(document, clampedCursor.line))
}
