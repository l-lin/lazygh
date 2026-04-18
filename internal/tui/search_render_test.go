package tui

import (
	"strings"
	"testing"
)

func TestHighlightSearchMatches_GivenDetailQuery_WhenRendering_ThenMatchesUseTheSearchBackgroundColor(t *testing.T) {
	actual, actualCount := highlightSearchMatches("Hello world\nworld", "world")

	if actualCount != 2 {
		t.Fatalf("expected 2 matches, actual %d", actualCount)
	}

	expectedSequence := "\x1b[48;2;249;234;179mworld\x1b[0m"
	if !strings.Contains(actual, expectedSequence) {
		t.Fatalf("expected highlighted output to contain %q, actual %q", expectedSequence, actual)
	}
}

func TestLayout_GivenSubmittedUserSearchWithNoMatches_WhenRendering_ThenTheUserViewShowsTheNoMatchesMessage(t *testing.T) {
	model := given_model()
	model.StartSearch()
	model.UpdateSearchDraft("nope")
	model.SubmitSearch()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if !strings.Contains(userView.Buffer(), `No matches for "nope".`) {
		t.Fatalf("expected user view to show the no matches message, actual %q", userView.Buffer())
	}
}
