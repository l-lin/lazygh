package tui

import (
	"reflect"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
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
		{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "Requested", Command: []string{"search", "prs", "--review-requested", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "Escalated", Command: []string{"search", "prs", "--search", "label:escalated state:open", "--sort", "updated", "--order", "desc"}},
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

func TestApplyPullRequestSearches_GivenConfiguredGUI_WhenApplying_ThenItRefreshesThePullRequestTabsThroughTheRuntimeBridge(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	then_noError(t, subject.layout(gui))

	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{
		{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "Requested", Command: []string{"search", "prs", "--review-requested", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
		{Label: "Escalated", Command: []string{"search", "prs", "--search", "label:escalated state:open", "--sort", "updated", "--order", "desc"}},
	})

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	then_tabsAre(t, pullRequestsView, []string{"Mine", "Requested", "Escalated"}, 0)
}

func TestApplyPullRequestSearches_GivenAConfiguredReplacementList_WhenApplying_ThenItReplacesTheDefaultLoadingTabs(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))
	expected := []appconfig.PullRequestSearch{
		{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open", "--sort", "updated", "--order", "desc"}},
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
	if actualPullRequests[0].Title != myPullRequestsLoadingTitle {
		t.Fatalf("expected title %q, actual %q", myPullRequestsLoadingTitle, actualPullRequests[0].Title)
	}
}

func TestApplyPullRequestSearches_GivenExistingPullRequestListLoadState_WhenApplying_ThenItResetsFlagsCountsAndAdditionalTabs(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.myPullRequestsLoading = true
	subject.requestedPullRequestsLoading = true
	subject.myPullRequestsCount = 3
	subject.myPullRequestsCountKnown = true
	subject.requestedPullRequestsCount = 4
	subject.requestedPullRequestsCountKnown = true
	subject.additionalPullRequestsLoadStarted[PullRequestTab(2)] = true
	subject.additionalPullRequestsLoading[PullRequestTab(2)] = true
	subject.additionalPullRequestsCounts[PullRequestTab(2)] = pullRequestCountState{count: 7, known: true}

	subject.ApplyPullRequestSearches([]appconfig.PullRequestSearch{{Label: "Mine", Command: []string{"search", "prs", "--author", "@me", "--state", "open"}}})

	if actual := subject.myPullRequestsLoadStarted; actual {
		t.Fatalf("expected my pull requests load started %v, actual %v", false, actual)
	}
	if actual := subject.requestedPullRequestsLoadStarted; actual {
		t.Fatalf("expected requested pull requests load started %v, actual %v", false, actual)
	}
	if actual := subject.myPullRequestsLoading; actual {
		t.Fatalf("expected my pull requests loading %v, actual %v", false, actual)
	}
	if actual := subject.requestedPullRequestsLoading; actual {
		t.Fatalf("expected requested pull requests loading %v, actual %v", false, actual)
	}
	if actual := subject.myPullRequestsCount; actual != 0 {
		t.Fatalf("expected my pull requests count %d, actual %d", 0, actual)
	}
	if actual := subject.myPullRequestsCountKnown; actual {
		t.Fatalf("expected my pull requests count known %v, actual %v", false, actual)
	}
	if actual := subject.requestedPullRequestsCount; actual != 0 {
		t.Fatalf("expected requested pull requests count %d, actual %d", 0, actual)
	}
	if actual := subject.requestedPullRequestsCountKnown; actual {
		t.Fatalf("expected requested pull requests count known %v, actual %v", false, actual)
	}
	if actual := len(subject.additionalPullRequestsLoadStarted); actual != 0 {
		t.Fatalf("expected additional pull requests load-start map length %d, actual %d", 0, actual)
	}
	if actual := len(subject.additionalPullRequestsLoading); actual != 0 {
		t.Fatalf("expected additional pull requests loading map length %d, actual %d", 0, actual)
	}
	if actual := len(subject.additionalPullRequestsCounts); actual != 0 {
		t.Fatalf("expected additional pull requests count map length %d, actual %d", 0, actual)
	}
}
