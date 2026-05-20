package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
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

func TestLayout_GivenSubmittedUserSearchWithNoMatches_WhenRendering_ThenTheUserViewKeepsTheRowsVisibleAndShowsZeroMatchesInTheFooter(t *testing.T) {
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
	if !strings.Contains(userView.Buffer(), "dummy-user-1") || !strings.Contains(userView.Buffer(), "dummy-user-2") {
		t.Fatalf("expected user rows to stay visible after a no-match search, actual %q", userView.Buffer())
	}
	then_footerTextIs(t, gui, viewUserFooterName, "/nope (0 matches)")
}

func TestLayout_GivenSubmittedUserSearchOnTheSelectedRow_WhenRendering_ThenTheUserViewKeepsTheOtherRowsVisibleAndHighlightsTheMatch(t *testing.T) {
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

	userView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if !strings.Contains(userView.Buffer(), "dummy-user-1") {
		t.Fatalf("expected the non-matching user row to stay visible, actual %q", userView.Buffer())
	}

	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewUserName, 0, "dummy-user-")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewUserName, 0, "2")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewUserName, 0, "2")
	then_viewLineSegmentIsBold(t, gui, viewUserName, 0, "dummy-user-")
	then_viewLineSegmentIsBold(t, gui, viewUserName, 0, "2")
}

func TestLayout_GivenSubmittedNotificationsSearchOnTheSelectedRow_WhenRendering_ThenTheNotificationsViewKeepsTheOtherRowsVisibleAndHighlightsTheMatch(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
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

	notificationsView, actualErr := gui.View(viewNotificationsName)
	then_noError(t, actualErr)
	if !strings.Contains(notificationsView.Buffer(), "notification-1") {
		t.Fatalf("expected the non-matching notification row to stay visible, actual %q", notificationsView.Buffer())
	}

	selectedLineIndex := given_viewLineIndexContaining(t, notificationsView, "notification-2")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewNotificationsName, selectedLineIndex, "notification-")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewNotificationsName, selectedLineIndex, "2")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewNotificationsName, selectedLineIndex, "2")
	then_viewLineSegmentIsBold(t, gui, viewNotificationsName, selectedLineIndex, "notification-")
	then_viewLineSegmentIsBold(t, gui, viewNotificationsName, selectedLineIndex, "2")
}

func TestLayout_GivenSubmittedPullRequestsSearchOnTheSelectedRow_WhenRendering_ThenThePullRequestsViewKeepsTheOtherRowsVisibleAndHighlightsTheMatch(t *testing.T) {
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

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "my-pr-1") {
		t.Fatalf("expected the non-matching row to stay visible, actual %q", pullRequestsView.Buffer())
	}

	selectedLineIndex := given_viewLineIndexContaining(t, pullRequestsView, "my-pr-2")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, selectedLineIndex, "my-pr-")
	then_viewLineSegmentHasSearchHighlightBackground(t, gui, viewPullRequestsName, selectedLineIndex, "2")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewPullRequestsName, selectedLineIndex, "2")
	then_viewLineSegmentIsBold(t, gui, viewPullRequestsName, selectedLineIndex, "my-pr-")
	then_viewLineSegmentIsBold(t, gui, viewPullRequestsName, selectedLineIndex, "2")
}
