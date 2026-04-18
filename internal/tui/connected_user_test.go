package tui

import (
	"fmt"
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestDefaultSeedData_GivenAFreshModel_WhenReadingTheConnectedUser_ThenItStartsInALoadingState(t *testing.T) {
	subject := NewModel(DefaultSeedData())

	actualUsers := subject.Users()
	if len(actualUsers) != 1 {
		t.Fatalf("expected 1 user row, actual %d", len(actualUsers))
	}
	if actualUsers[0].Title != connectedUserLoadingTitle {
		t.Fatalf("expected title %q, actual %q", connectedUserLoadingTitle, actualUsers[0].Title)
	}
	if actualUsers[0].Detail != connectedUserLoadingDetail {
		t.Fatalf("expected detail %q, actual %q", connectedUserLoadingDetail, actualUsers[0].Detail)
	}
}

func TestSetUsers_GivenAConnectedUser_WhenSelectingTheUserView_ThenDetailContentShowsTheRealUserSummary(t *testing.T) {
	subject := NewModel(DefaultSeedData())
	subject.SetUsers([]Item{connectedUserItem(githubcli.ConnectedUser{
		Login:       "octocat",
		Name:        "Mona Lisa Octocat",
		Bio:         "Mascot on call",
		Company:     "GitHub",
		Location:    "The Internet",
		PublicRepos: 8,
		Followers:   42,
		URL:         "https://github.com/octocat",
	})})

	actualUsers := subject.Users()
	if len(actualUsers) != 1 {
		t.Fatalf("expected 1 user row, actual %d", len(actualUsers))
	}
	if actualUsers[0].Title != "@octocat" {
		t.Fatalf("expected title %q, actual %q", "@octocat", actualUsers[0].Title)
	}

	actualDetail := subject.DetailContent()
	expectedFragments := []string{
		"Login: @octocat",
		"Name: Mona Lisa Octocat",
		"Bio: Mascot on call",
		"Company: GitHub",
		"Location: The Internet",
		"Public repos: 8",
		"Followers: 42",
		"Profile: https://github.com/octocat",
	}
	for _, expected := range expectedFragments {
		if !strings.Contains(actualDetail, expected) {
			t.Fatalf("expected detail to contain %q, actual %q", expected, actualDetail)
		}
	}
}

func TestConnectedUserItem_GivenAnEmptyLogin_WhenBuildingTheUserState_ThenItFallsBackToTheEmptyState(t *testing.T) {
	actual := connectedUserItem(githubcli.ConnectedUser{})

	if actual.Title != connectedUserEmptyTitle {
		t.Fatalf("expected title %q, actual %q", connectedUserEmptyTitle, actual.Title)
	}
	if actual.Detail != connectedUserEmptyDetail {
		t.Fatalf("expected detail %q, actual %q", connectedUserEmptyDetail, actual.Detail)
	}
}

func TestConnectedUserErrorItem_GivenAnAuthenticationError_WhenBuildingTheUserState_ThenItShowsTheRecoveryMessage(t *testing.T) {
	actual := connectedUserErrorItem(fmt.Errorf("wrap: %w", githubcli.ErrUnauthenticated))

	if actual.Title != connectedUserUnauthenticatedTitle {
		t.Fatalf("expected title %q, actual %q", connectedUserUnauthenticatedTitle, actual.Title)
	}
	if actual.Detail != connectedUserUnauthenticatedDetail {
		t.Fatalf("expected detail %q, actual %q", connectedUserUnauthenticatedDetail, actual.Detail)
	}
}
