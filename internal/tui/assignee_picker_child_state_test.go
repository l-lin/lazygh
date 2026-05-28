package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestAssigneePickerState_GivenSelectedCandidate_WhenTogglingSelection_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := *newAssigneePickerState(pullRequestAssigneePickerTarget{repository: "acme/widgets", number: 42}, "viewer", "Viewer")
	subject.searchRequestID = 3

	actual := subject.withSelectionToggled(githubdomain.PullRequestAuthor{Login: "alice", Name: "Alice"})

	if !actual.selectedLogins["alice"] {
		t.Fatal("expected the toggled candidate to become selected")
	}
	if actual.knownCandidates["alice"].Name != "Alice" {
		t.Fatalf("expected the toggled candidate to be remembered as %q, actual %q", "Alice", actual.knownCandidates["alice"].Name)
	}
	if _, ok := subject.selectedLogins["alice"]; ok {
		t.Fatalf("expected the original selection map to stay unchanged, actual %v", subject.selectedLogins)
	}
	if _, ok := subject.knownCandidates["alice"]; ok {
		t.Fatalf("expected the original known-candidate map to stay unchanged, actual %v", subject.knownCandidates)
	}
	if actual.searchRequestID != 3 {
		t.Fatalf("expected the request id to stay %d, actual %d", 3, actual.searchRequestID)
	}
}

func TestAssigneePickerState_GivenWarmSearchState_WhenResettingLoadingAndCompletingSearch_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := *newAssigneePickerState(pullRequestAssigneePickerTarget{repository: "acme/widgets", number: 42}, "viewer", "Viewer")
	subject.searchRequestID = 2
	subject.searchQuery = "stale"
	subject.searchResults = []githubdomain.PullRequestAuthor{{Login: "old", Name: "Old"}}
	subject.searchLoading = true
	subject.searchCommand = "gh api graphql"

	reset, actualRequestID := subject.withSearchReset(" ali ")
	loading := reset.withSearchLoadingStarted("ali")
	actual := loading.withSearchLoaded(" ali ", []githubdomain.PullRequestAuthor{{Login: "alice", Name: "Alice"}})

	if actualRequestID != 3 {
		t.Fatalf("expected the incremented request id %d, actual %d", 3, actualRequestID)
	}
	if actual.searchRequestID != 3 {
		t.Fatalf("expected the loaded state to keep request id %d, actual %d", 3, actual.searchRequestID)
	}
	if actual.searchQuery != "ali" {
		t.Fatalf("expected the loaded query %q, actual %q", "ali", actual.searchQuery)
	}
	if actual.searchLoading {
		t.Fatal("expected the loaded state to clear the loading flag")
	}
	if actual.searchCommand != "" {
		t.Fatalf("expected the loaded command %q, actual %q", "", actual.searchCommand)
	}
	if len(actual.searchResults) != 1 || actual.searchResults[0].Login != "alice" {
		t.Fatalf("expected the loaded search results %+v, actual %+v", []githubdomain.PullRequestAuthor{{Login: "alice", Name: "Alice"}}, actual.searchResults)
	}
	if actual.knownCandidates["alice"].Name != "Alice" {
		t.Fatalf("expected the loaded candidate to be remembered as %q, actual %q", "Alice", actual.knownCandidates["alice"].Name)
	}
	if subject.searchRequestID != 2 || subject.searchQuery != "stale" || !subject.searchLoading || subject.searchCommand != "gh api graphql" {
		t.Fatalf("expected the original search state to stay unchanged, actual %+v", subject)
	}
	if len(subject.searchResults) != 1 || subject.searchResults[0].Login != "old" {
		t.Fatalf("expected the original results to stay unchanged, actual %+v", subject.searchResults)
	}
	if _, ok := subject.knownCandidates["alice"]; ok {
		t.Fatalf("expected the original known candidates to stay unchanged, actual %v", subject.knownCandidates)
	}
	if loading.searchCommand == "" {
		t.Fatal("expected the loading transition to install the gh command display")
	}
}
