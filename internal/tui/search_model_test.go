package tui

import (
	"reflect"
	"testing"
)

func TestStartSearch_GivenPullRequestsFocusOnRequestedTab_WhenStartingSearch_ThenSearchTargetsTheActiveTab(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.NextPullRequestTab()

	subject.StartSearch()

	if !subject.SearchActive() {
		t.Fatal("expected search to be active")
	}
	if subject.SearchTarget() != FocusPullRequestsView {
		t.Fatalf("expected search target %v, actual %v", FocusPullRequestsView, subject.SearchTarget())
	}
	if subject.SearchTargetPullRequestTab() != RequestedPullRequestsTab {
		t.Fatalf("expected search tab %v, actual %v", RequestedPullRequestsTab, subject.SearchTargetPullRequestTab())
	}
}

func TestUpdateSearchDraft_GivenUserSearch_WhenFiltering_ThenVisibleUsersAndSelectionClampToMatches(t *testing.T) {
	subject := given_model()
	subject.MoveSelectionDown()
	subject.StartSearch()

	subject.UpdateSearchDraft("1")

	expected := []string{"dummy-user-1"}
	actual := given_titles(subject.VisibleUsers())
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected visible users %v, actual %v", expected, actual)
	}
	if subject.SelectedUserIndex() != 0 {
		t.Fatalf("expected raw selected user index 0, actual %d", subject.SelectedUserIndex())
	}
	if subject.SelectedVisibleUserIndex() != 0 {
		t.Fatalf("expected visible selected user index 0, actual %d", subject.SelectedVisibleUserIndex())
	}
}

func TestSubmitSearch_GivenRequestedPullRequestsSearch_WhenSubmitting_ThenOnlyTheActiveTabKeepsTheQuery(t *testing.T) {
	subject := given_model()
	subject.NextSideView()
	subject.NextPullRequestTab()
	subject.StartSearch()
	subject.UpdateSearchDraft("2")

	subject.SubmitSearch()

	expectedRequested := []string{"requested-pr-2"}
	actualRequested := given_titles(subject.VisiblePullRequests())
	if !reflect.DeepEqual(actualRequested, expectedRequested) {
		t.Fatalf("expected requested pull requests %v, actual %v", expectedRequested, actualRequested)
	}

	subject.PreviousPullRequestTab()

	expectedMine := []string{"my-pr-1", "my-pr-2"}
	actualMine := given_titles(subject.VisiblePullRequests())
	if !reflect.DeepEqual(actualMine, expectedMine) {
		t.Fatalf("expected my pull requests %v, actual %v", expectedMine, actualMine)
	}
}

func TestCancelSearch_GivenExistingAppliedUserQuery_WhenCancelingNewDraft_ThenThePreviousQueryRemains(t *testing.T) {
	subject := given_model()
	subject.StartSearch()
	subject.UpdateSearchDraft("1")
	subject.SubmitSearch()
	subject.StartSearch()
	subject.UpdateSearchDraft("2")

	subject.CancelSearch()

	expected := []string{"dummy-user-1"}
	actual := given_titles(subject.VisibleUsers())
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected visible users %v, actual %v", expected, actual)
	}
}

func given_titles(items []Item) []string {
	titles := make([]string, 0, len(items))
	for _, item := range items {
		titles = append(titles, item.Title)
	}

	return titles
}
