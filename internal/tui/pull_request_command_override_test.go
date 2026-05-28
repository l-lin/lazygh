package tui

import (
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
)

func TestApplyPullRequestSearches_GivenCustomReviewRequestedSearch_WhenReadingTheConfiguredLoadingItem_ThenItUsesReviewRequestMessagingInsteadOfTheLabel(t *testing.T) {
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
	expected := "Running `gh pr list --search review-requested:@me --state open --json title,number,repository,url,body,state,isDraft,updatedAt,id` to load review requests."
	if actualPullRequests[0].Detail != expected {
		t.Fatalf("expected loading detail %q, actual %q", expected, actualPullRequests[0].Detail)
	}
}

func TestStatusLineText_GivenCustomAuthoredSearchWhileLoading_WhenReadingTheStatusLine_ThenItUsesAuthoredMessagingInsteadOfTheLabel(t *testing.T) {
	model := NewModel(DefaultSeedData())
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{{
		Label:   "Mine",
		Command: []string{"pr", "list", "--search", "author:@me status:open"},
	}})
	subject.myPullRequestsLoading = true

	actual := subject.statusLinePresenter().Text()

	expected := string(loadingSpinnerFrames[0]) + " Running `gh pr list --search author:@me status:open --json title,number,repository,url,body,state,isDraft,updatedAt,id` to load authored pull requests."
	if actual != expected {
		t.Fatalf("expected status line %q, actual %q", expected, actual)
	}
}
