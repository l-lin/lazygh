package tui

import (
	"reflect"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

func TestNextPullRequestTab_GivenThreeConfiguredTabs_WhenSwitching_ThenItCyclesAcrossAllTabs(t *testing.T) {
	subject := NewModel(SeedData{
		PullRequestTabs: []PullRequestTabSeed{
			{Label: "Mine", PullRequests: []Item{{Title: "mine-1", Detail: "detail-1"}}},
			{Label: "Requested", PullRequests: []Item{{Title: "requested-1", Detail: "detail-2"}}},
			{Label: "Escalated", PullRequests: []Item{{Title: "escalated-1", Detail: "detail-3"}}},
		},
	})
	subject.FocusPullRequestsView()

	subject.NextPullRequestTab()
	if subject.ActivePullRequestTab() != PullRequestTab(1) {
		t.Fatalf("expected active tab %d, actual %d", 1, subject.ActivePullRequestTab())
	}

	subject.NextPullRequestTab()
	if subject.ActivePullRequestTab() != PullRequestTab(2) {
		t.Fatalf("expected active tab %d, actual %d", 2, subject.ActivePullRequestTab())
	}

	subject.NextPullRequestTab()
	if subject.ActivePullRequestTab() != PullRequestTab(0) {
		t.Fatalf("expected active tab %d, actual %d", 0, subject.ActivePullRequestTab())
	}
}

func TestApplyPullRequestSearches_GivenConfiguredSearches_WhenRendering_ThenItUsesTheirLabelsInOrder(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))
	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{
		{Label: "Mine", Command: []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"}},
		{Label: "Requested", Command: []string{"search", "prs", "--review-requested", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"}},
		{Label: "Escalated", Command: []string{"search", "prs", "--search", "label:escalated state:open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"}},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	then_tabsAre(t, pullRequestsView, []string{"Mine", "Requested", "Escalated"}, 0)
}

func TestApplyPullRequestSearches_GivenAConfiguredReplacementList_WhenApplying_ThenItReplacesTheDefaultLoadingTabs(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))
	expected := []appconfig.PullRequestSearch{
		{Label: "Mine", Command: []string{"pr", "list", "--author", "@me", "--state", "open", "--json", "title,number,repository,url,body,state,isDraft,updatedAt"}},
	}

	subject.ApplyPullRequestSearches(expected)

	actualLabels := []string{subject.model.PullRequestTabLabel(PullRequestTab(0))}
	if !reflect.DeepEqual(actualLabels, []string{"Mine"}) {
		t.Fatalf("expected labels %v, actual %v", []string{"Mine"}, actualLabels)
	}
	actualPullRequests := subject.model.PullRequests(PullRequestTab(0))
	if len(actualPullRequests) != 1 {
		t.Fatalf("expected 1 pull request row, actual %d", len(actualPullRequests))
	}
	if actualPullRequests[0].Title != "Loading Mine..." {
		t.Fatalf("expected title %q, actual %q", "Loading Mine...", actualPullRequests[0].Title)
	}
}
