package tui

import "testing"

func TestKeybindingSpecs_GivenPendingCharacterMotionAndStaleCursor_WhenListingBindings_ThenItDoesNotMutateStoredState(t *testing.T) {
	t.Run("detail", func(t *testing.T) {
		model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha"}}})
		model.OpenDetail()
		subject := NewProgramWithModel(model)
		subject.detailState.viewState.cursor = detailPosition{line: 99, column: 0}
		subject.detailState.viewState.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)

		_ = subject.registeredKeybindingSpecs()

		if actual := subject.detailState.viewState.cursor.line; actual != 99 {
			t.Fatalf("expected detail cursor line %d to stay unchanged while deriving bindings, actual %d", 99, actual)
		}
	})

	t.Run("build popup", func(t *testing.T) {
		subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
		subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "CI / test", body: "Alpha"})
		subject.pullRequestBuildRunPopup.viewState.cursor = detailPosition{line: 99, column: 0}
		subject.pullRequestBuildRunPopup.viewState.armCharacterMotion(detailCharacterMotionDirectionForward, detailCharacterMotionMatch)

		_ = subject.registeredKeybindingSpecs()

		if actual := subject.pullRequestBuildRunPopup.viewState.cursor.line; actual != 99 {
			t.Fatalf("expected popup cursor line %d to stay unchanged while deriving bindings, actual %d", 99, actual)
		}
	})
}
