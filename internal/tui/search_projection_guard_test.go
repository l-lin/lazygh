package tui

import "testing"

func TestSubmitSearch_GivenReviewTreeSearchTarget_WhenSubmitting_ThenItStoresTheAppliedReviewTreeQueryInTheModel(t *testing.T) {
	subject := given_model()

	subject.StartSearchForReviewTree(MyPullRequestsTab)
	subject.UpdateSearchDraft("render")
	subject.SubmitSearch()

	if actual := subject.ReviewTreeSearchQuery(); actual != "render" {
		t.Fatalf("expected applied review-tree search query %q, actual %q", "render", actual)
	}
	if actual := subject.SearchTargetKind(); actual != SearchTargetReviewTree {
		t.Fatalf("expected search target kind %v, actual %v", SearchTargetReviewTree, actual)
	}
}

func TestProjectedScreenStateApplication_GivenBrowserScreenTabs_WhenProjecting_ThenItReturnsThePureModelAndDetailUpdates(t *testing.T) {
	state := newBrowserScreenState(FocusDetailView, RequestedPullRequestsTab, []TabState{{Label: "Description"}, {Label: "Comments"}, {Label: "Commits"}, {Label: "Changes"}}, []TabState{{Label: "Mine"}, {Label: "Requested"}})
	state = state.WithViewTabs(mainPanelViewNumber, 3, state.MainPanel.Views[0].Tabs)

	actual := projectScreenStateApplication(state)

	if actual.focus != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, actual.focus)
	}
	if actual.lastSideFocus != FocusUserView {
		t.Fatalf("expected last side focus %v, actual %v", FocusUserView, actual.lastSideFocus)
	}
	if actual.activePullRequestTab != RequestedPullRequestsTab {
		t.Fatalf("expected pull request tab %v, actual %v", RequestedPullRequestsTab, actual.activePullRequestTab)
	}
	if actual.activeDetailTabIndex != 3 {
		t.Fatalf("expected detail tab index %d, actual %d", 3, actual.activeDetailTabIndex)
	}
}
