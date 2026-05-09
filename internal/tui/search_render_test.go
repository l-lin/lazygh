package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestHighlightSearchMatches_GivenDetailQuery_WhenRendering_ThenMatchesUseTheSearchBackgroundColor(t *testing.T) {
	actual, actualCount := highlightSearchMatches("Hello world\nworld", "world")

	if actualCount != 2 {
		t.Fatalf("expected 2 matches, actual %d", actualCount)
	}

	expectedSequence := backgroundColorEscape(theme.SearchHighlightHex) + "world" + ansiReset
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

func TestLayout_GivenSubmittedUserSearchOnTheSelectedRow_WhenRendering_ThenTheUserViewKeepsSearchBackgroundOnTheMatchAndSelectionBackgroundElsewhere(t *testing.T) {
	model := given_model()
	model.MoveSelectionDown()
	model.StartSearch()
	model.UpdateSearchDraft("2")
	model.SubmitSearch()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewUserName, 0, "dummy-user-")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewUserName, 0, "2")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewUserName, 0, "2")
	then_viewLineSegmentIsBold(t, gui, viewUserName, 0, "dummy-user-")
	then_viewLineSegmentIsBold(t, gui, viewUserName, 0, "2")
}

func TestLayout_GivenSubmittedPullRequestsSearchOnTheSelectedRow_WhenRendering_ThenThePullRequestsViewKeepsSearchBackgroundOnTheMatchAndSelectionBackgroundElsewhere(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.MoveSelectionDown()
	model.StartSearch()
	model.UpdateSearchDraft("2")
	model.SubmitSearch()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, 0, "my-pr-")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewPullRequestsName, 0, "2")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewPullRequestsName, 0, "2")
	then_viewLineSegmentIsBold(t, gui, viewPullRequestsName, 0, "my-pr-")
	then_viewLineSegmentIsBold(t, gui, viewPullRequestsName, 0, "2")
}
