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
		"program.detailState.wrapWidth =",
		"program.syncDetailViewState(",
	} {
		if strings.Contains(actualSource, forbiddenSnippet) {
			t.Fatalf("expected render.go to stay free of %q, actual source:\n%s", forbiddenSnippet, actualSource)
		}
	}
}

func TestDetailViewRenderState_GivenDetailView_WhenPreparing_ThenItLeavesDurableDetailStateUntouched(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()

	view, actualErr := gui.SetView(viewDetailName, 0, 0, 79, 19, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	subject.detailState.wrapWidth = defaultDetailWrapWidth
	subject.detailState.viewState.cursor = detailPosition{line: 999}
	subject.detailState.lastIdentity = "stale"

	subject.prepareViewRenderState(viewDetailName, view)

	if actual := subject.detailState.wrapWidth; actual != defaultDetailWrapWidth {
		t.Fatalf("expected durable detail wrap width %d, actual %d", defaultDetailWrapWidth, actual)
	}
	if actual := subject.detailState.lastIdentity; actual != "stale" {
		t.Fatalf("expected durable detail identity %q, actual %q", "stale", actual)
	}
	if actual := subject.detailState.viewState.cursor.line; actual != 999 {
		t.Fatalf("expected the durable detail cursor line %d, actual %d", 999, actual)
	}
}

func TestDetailViewRenderState_GivenDetailView_WhenSyncingViewShellState_ThenItUpdatesWrapWidthAndClampsDetailStateBeforeRender(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGuiWithSize(t, 120, 40)
	defer gui.Close()

	view, actualErr := gui.SetView(viewDetailName, 0, 0, 79, 19, 0)
	if actualErr != nil && !isUnknownViewError(actualErr) {
		then_noError(t, actualErr)
	}
	subject.detailState.wrapWidth = defaultDetailWrapWidth
	subject.detailState.viewState.cursor = detailPosition{line: 999}
	subject.detailState.lastIdentity = "stale"

	subject.syncViewShellState(viewDetailName, view)

	expectedWrapWidth := effectiveMarkdownWidth(view.InnerWidth())
	if actual := subject.detailState.wrapWidth; actual != expectedWrapWidth {
		t.Fatalf("expected detail wrap width %d, actual %d", expectedWrapWidth, actual)
	}
	if actual := subject.detailState.lastIdentity; actual == "stale" {
		t.Fatalf("expected detail identity to sync away from %q", "stale")
	}
	if actual := subject.detailState.viewState.cursor.line; actual >= 999 {
		t.Fatalf("expected the detail cursor to be clamped during shell sync, actual line %d", actual)
	}
}
