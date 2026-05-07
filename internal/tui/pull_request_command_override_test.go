package tui

import (
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

func TestApplyPullRequestSearches_GivenCustomSearches_WhenReadingTheConfiguredLoadingItem_ThenItShowsTheConfiguredCommand(t *testing.T) {
	model := NewModel(DefaultSeedData())
	subject := NewProgramWithModel(model)
	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{{
		Label:   "Team Review",
		Command: []string{"pr", "list", "--search", "review-requested:@me", "--state", "open"},
	}})

	actualPullRequests := subject.model.PullRequests(PullRequestTab(0))

	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	expected := "Running `gh pr list --search review-requested:@me --state open --json title,number,repository,url,body,state,isDraft,updatedAt` to load pull requests for Team Review."
	if actualPullRequests[0].Detail != expected {
		t.Fatalf("expected loading detail %q, actual %q", expected, actualPullRequests[0].Detail)
	}
}

func TestStatusLineText_GivenConfiguredSearchWhileLoading_WhenReadingTheStatusLine_ThenItUsesTheConfiguredCommand(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{{
		Label:   "Mine",
		Command: []string{"pr", "list", "--search", "author:@me status:open"},
	}})
	subject.myPullRequestsLoading = true

	actual := subject.statusLineText()

	expected := string(loadingSpinnerFrames[0]) + " Running `gh pr list --search author:@me status:open --json title,number,repository,url,body,state,isDraft,updatedAt` to load pull requests for Mine."
	if actual != expected {
		t.Fatalf("expected status line %q, actual %q", expected, actual)
	}
}
