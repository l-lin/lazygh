package tui

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestDefaultSeedData_GivenAFreshModel_WhenReadingMyPullRequests_ThenItStartsInALoadingState(t *testing.T) {
	subject := NewModel(DefaultSeedData())

	actualPullRequests := subject.PullRequests(MyPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != myPullRequestsLoadingTitle {
		t.Fatalf("expected title %q, actual %q", myPullRequestsLoadingTitle, actualPullRequests[0].Title)
	}
	if actualPullRequests[0].Detail != myPullRequestsLoadingDetail {
		t.Fatalf("expected detail %q, actual %q", myPullRequestsLoadingDetail, actualPullRequests[0].Detail)
	}
}

func TestSetPullRequests_GivenMyPullRequests_WhenSelectingThePullRequestsView_ThenDetailContentShowsMetadataAndBody(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(MyPullRequestsTab, []Item{myPullRequestItem(githubcli.PullRequest{
		Title:      "fix(P3C-6986): exclude dependencies bump PRs + bump GHA",
		Number:     422,
		Repository: githubcli.Repository{NameWithOwner: "doctolib/patient-account"},
		URL:        "https://github.com/doctolib/patient-account/pull/422",
		Body:       "No need to trigger Claude review for PRs that only bump dependencies.",
		State:      "open",
		IsDraft:    false,
		UpdatedAt:  "2026-04-17T10:39:35Z",
	})})
	subject.FocusPullRequestsView()

	actualPullRequests := subject.PullRequests(MyPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != "doctolib/patient-account#422 fix(P3C-6986): exclude dependencies bump PRs + bump GHA" {
		t.Fatalf("expected title %q, actual %q", "doctolib/patient-account#422 fix(P3C-6986): exclude dependencies bump PRs + bump GHA", actualPullRequests[0].Title)
	}

	actualDetail := subject.DetailContent()
	expectedFragments := []string{
		"Repository: doctolib/patient-account",
		"Number: #422",
		"State: open",
		"Draft: no",
		"Updated: 2026-04-17T10:39:35Z",
		"URL: https://github.com/doctolib/patient-account/pull/422",
		"No need to trigger Claude review for PRs that only bump dependencies.",
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(actualDetail, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actualDetail)
		}
	}
}

func TestSetPullRequests_GivenASelectedMyPullRequest_WhenRefreshingTheList_ThenTheSelectionIndexIsPreserved(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(MyPullRequestsTab, []Item{
		{Title: "pr-1", Detail: "detail-1"},
		{Title: "pr-2", Detail: "detail-2"},
		{Title: "pr-3", Detail: "detail-3"},
	})
	subject.FocusPullRequestsView()
	subject.MoveSelectionDown()

	subject.SetPullRequests(MyPullRequestsTab, []Item{
		{Title: "pr-1", Detail: "detail-1"},
		{Title: "pr-2", Detail: "detail-2"},
		{Title: "pr-3", Detail: "detail-3"},
	})

	actual := subject.SelectedPullRequestIndex(MyPullRequestsTab)
	if actual != 1 {
		t.Fatalf("expected selection 1, actual %d", actual)
	}
}

func TestMyPullRequestsErrorItem_GivenAnAuthenticationError_WhenBuildingTheState_ThenItShowsTheRecoveryMessage(t *testing.T) {
	actual := myPullRequestsErrorItem(fmt.Errorf("wrap: %w", githubcli.ErrUnauthenticated))

	if actual.Title != myPullRequestsUnauthenticatedTitle {
		t.Fatalf("expected title %q, actual %q", myPullRequestsUnauthenticatedTitle, actual.Title)
	}
	if actual.Detail != myPullRequestsUnauthenticatedDetail {
		t.Fatalf("expected detail %q, actual %q", myPullRequestsUnauthenticatedDetail, actual.Detail)
	}
}

func TestDefaultSeedData_GivenAFreshModel_WhenReadingRequestedPullRequests_ThenItStartsInALoadingState(t *testing.T) {
	subject := NewModel(DefaultSeedData())

	actualPullRequests := subject.PullRequests(RequestedPullRequestsTab)
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != requestedPullRequestsLoadingTitle {
		t.Fatalf("expected title %q, actual %q", requestedPullRequestsLoadingTitle, actualPullRequests[0].Title)
	}
	if actualPullRequests[0].Detail != requestedPullRequestsLoadingDetail {
		t.Fatalf("expected detail %q, actual %q", requestedPullRequestsLoadingDetail, actualPullRequests[0].Detail)
	}
}

func TestSetPullRequests_GivenRequestedPullRequests_WhenSelectingTheRequestedTab_ThenDetailContentShowsMetadataAndBody(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetPullRequests(RequestedPullRequestsTab, []Item{requestedPullRequestItem(githubcli.PullRequest{
		Title:      "feat(doctolib-postmortems): integrate post-mortem writing guide",
		Number:     845,
		Repository: githubcli.Repository{NameWithOwner: "doctolib/prompts"},
		URL:        "https://github.com/doctolib/prompts/pull/845",
		Body:       "## Summary\n\n- Adds new skill",
		State:      "open",
		IsDraft:    false,
		UpdatedAt:  "2026-04-17T20:35:05Z",
	})})
	subject.FocusPullRequestsView()
	subject.NextPullRequestTab()

	actualDetail := subject.DetailContent()
	expectedFragments := []string{
		"Repository: doctolib/prompts",
		"Number: #845",
		"State: open",
		"Draft: no",
		"Updated: 2026-04-17T20:35:05Z",
		"URL: https://github.com/doctolib/prompts/pull/845",
		"## Summary",
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(actualDetail, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actualDetail)
		}
	}
}

func TestRequestedPullRequestsErrorItem_GivenAnAuthenticationError_WhenBuildingTheState_ThenItShowsTheRecoveryMessage(t *testing.T) {
	actual := requestedPullRequestsErrorItem(fmt.Errorf("wrap: %w", githubcli.ErrUnauthenticated))

	if actual.Title != requestedPullRequestsUnauthenticatedTitle {
		t.Fatalf("expected title %q, actual %q", requestedPullRequestsUnauthenticatedTitle, actual.Title)
	}
	if actual.Detail != requestedPullRequestsUnauthenticatedDetail {
		t.Fatalf("expected detail %q, actual %q", requestedPullRequestsUnauthenticatedDetail, actual.Detail)
	}
}
