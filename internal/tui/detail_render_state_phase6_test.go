package tui

import (
	"os"
	"strings"
	"testing"
)

func TestViewDerivation_GivenRenderGo_WhenInspecting_ThenItDoesNotMutateDetailShellState(t *testing.T) {
	contents, actualErr := os.ReadFile("render.go")
	then_noError(t, actualErr)
	actualSource := string(contents)

	for _, forbiddenSnippet := range []string{
		"program.detailWrapWidth =",
		"program.syncDetailViewState(",
	} {
		if strings.Contains(actualSource, forbiddenSnippet) {
			t.Fatalf("expected render.go to stay free of %q, actual source:\n%s", forbiddenSnippet, actualSource)
		}
	}
}

func TestDetailViewRenderState_GivenDetailView_WhenPreparing_ThenItUpdatesWrapWidthAndSyncsDetailState(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()

	view, actualErr := gui.SetView(viewDetailName, 0, 0, 79, 19, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	subject.detailWrapWidth = defaultDetailWrapWidth
	subject.detailViewState.cursor = detailPosition{line: 999}
	subject.lastDetailIdentity = "stale"

	subject.syncDetailViewRenderState(view)

	expectedWrapWidth := effectiveMarkdownWidth(view.InnerWidth())
	if actual := subject.detailWrapWidth; actual != expectedWrapWidth {
		t.Fatalf("expected detail wrap width %d, actual %d", expectedWrapWidth, actual)
	}
	if actual := subject.lastDetailIdentity; actual == "stale" {
		t.Fatalf("expected detail identity to sync away from %q", "stale")
	}
	if actual := subject.detailViewState.cursor.line; actual >= 999 {
		t.Fatalf("expected the detail cursor to be clamped during render-state prep, actual line %d", actual)
	}
}
