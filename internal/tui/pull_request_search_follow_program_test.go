package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestPullRequestsView_GivenSubmittedSearch_WhenRepeatingWithNAndN_ThenItMovesBetweenMatchingPullRequestsWithWraparound(t *testing.T) {
	model := NewModel(SeedData{
		Users: []Item{{Title: "Only user", Detail: "detail"}},
		MyPullRequests: []Item{
			{Title: "alpha", Detail: "detail alpha"},
			{Title: "beta one", Detail: "detail beta one"},
			{Title: "gamma beta", Detail: "detail gamma beta"},
			{Title: "delta", Detail: "detail delta"},
		},
	})
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.openSearch(gui, nil)
	then_noError(t, actualErr)
	searchView, actualErr := gui.View(viewSearchName)
	then_noError(t, actualErr)
	when_typingSearchQuery(t, subject, searchView, "beta")

	actualErr = subject.submitSearch(gui, searchView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_pullRequestSelectionIs(t, gui, subject, "beta one")

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "alpha") || !strings.Contains(pullRequestsView.Buffer(), "delta") {
		t.Fatalf("expected the pull request list to stay visible after search, actual %q", pullRequestsView.Buffer())
	}

	actualErr = subject.nextPullRequestsSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_pullRequestSelectionIs(t, gui, subject, "gamma beta")

	actualErr = subject.nextPullRequestsSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_pullRequestSelectionIs(t, gui, subject, "beta one")

	actualErr = subject.previousPullRequestsSearchMatch(gui, nil)
	then_noError(t, actualErr)
	then_pullRequestSelectionIs(t, gui, subject, "gamma beta")
}

func then_pullRequestSelectionIs(t *testing.T, gui *gocui.Gui, subject *Program, expectedTitle string) {
	t.Helper()

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	expectedRow := given_viewLineIndexContaining(t, pullRequestsView, expectedTitle)
	if subject.model.SelectedPullRequestIndex(subject.model.ActivePullRequestTab()) != expectedRow {
		t.Fatalf("expected selected pull request row %d for %q, actual %d", expectedRow, expectedTitle, subject.model.SelectedPullRequestIndex(subject.model.ActivePullRequestTab()))
	}
}
