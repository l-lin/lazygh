package tui

import "testing"

func TestCharacterMotionTargetReadModel_GivenAStaleDetailSelection_WhenResolvingTheDetailRunes_ThenItUsesTheSnapshotCursor(t *testing.T) {
	document := newDetailDocument("Alpha", 80)
	subject := characterMotionTargetReadModel{
		detail:      characterMotionTargetSelection{document: document, cursor: detailPosition{line: 99, column: 0}},
		detailKnown: true,
	}

	actual := subject.detailRunes()

	if !containsCharacterMotionTargetRune(actual, 'A') {
		t.Fatalf("expected detail target runes to include %q, actual %q", 'A', string(actual))
	}
}

func TestCharacterMotionTargetReadModel_GivenAStaleBuildPopupSelection_WhenResolvingThePopupRunes_ThenItUsesTheSnapshotCursor(t *testing.T) {
	document := newDetailDocument("Build", 80)
	subject := characterMotionTargetReadModel{
		buildPopup:      characterMotionTargetSelection{document: document, cursor: detailPosition{line: 99, column: 0}},
		buildPopupKnown: true,
	}

	actual := subject.buildPopupRunes()

	if !containsCharacterMotionTargetRune(actual, 'B') {
		t.Fatalf("expected popup target runes to include %q, actual %q", 'B', string(actual))
	}
}

func containsCharacterMotionTargetRune(values []rune, target rune) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
