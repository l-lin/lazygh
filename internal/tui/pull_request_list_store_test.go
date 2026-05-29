package tui

import "testing"

func TestPullRequestListStore_GivenDefaultAndAdditionalTabs_WhenApplyingLoadAndCountTransitions_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := newPullRequestListStore(nil)
	subject.myPullRequestsLoadStarted = true
	subject.additionalPullRequestsLoading[PullRequestTab(2)] = true

	loadStarted := subject.withPullRequestsLoadStarted(RequestedPullRequestsTab, true)
	loading := loadStarted.withPullRequestsLoading(PullRequestTab(2), false)
	counted := loading.withPullRequestsCount(PullRequestTab(2), 7, true)

	if !loadStarted.requestedPullRequestsLoadStarted {
		t.Fatal("expected the updated copy to mark requested pull requests as started")
	}
	if loading.additionalPullRequestsLoading[PullRequestTab(2)] {
		t.Fatal("expected the updated copy to clear the additional-tab loading flag")
	}
	if actual := counted.additionalPullRequestsCounts[PullRequestTab(2)]; actual.count != 7 || !actual.known {
		t.Fatalf("expected counted additional-tab state %+v, actual %+v", pullRequestCountState{count: 7, known: true}, actual)
	}

	if subject.requestedPullRequestsLoadStarted {
		t.Fatal("expected the original requested-tab load-start flag to stay false")
	}
	if !subject.additionalPullRequestsLoading[PullRequestTab(2)] {
		t.Fatal("expected the original additional-tab loading flag to stay true")
	}
	if _, ok := subject.additionalPullRequestsCounts[PullRequestTab(2)]; ok {
		t.Fatal("expected the original additional-tab count state to stay unset")
	}
}

func TestPullRequestListStore_GivenExistingStateAcrossDefaultAndAdditionalTabs_WhenResettingLoadState_ThenItClearsFlagsCountsAndTabMapsWithoutMutatingTheOriginal(t *testing.T) {
	subject := newPullRequestListStore(nil)
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

	reset := subject.withLoadStateReset()

	if actual := reset.myPullRequestsLoadStarted; actual {
		t.Fatalf("expected my pull requests load started %v, actual %v", false, actual)
	}
	if actual := reset.requestedPullRequestsLoadStarted; actual {
		t.Fatalf("expected requested pull requests load started %v, actual %v", false, actual)
	}
	if actual := reset.myPullRequestsLoading; actual {
		t.Fatalf("expected my pull requests loading %v, actual %v", false, actual)
	}
	if actual := reset.requestedPullRequestsLoading; actual {
		t.Fatalf("expected requested pull requests loading %v, actual %v", false, actual)
	}
	if actual := reset.myPullRequestsCount; actual != 0 {
		t.Fatalf("expected my pull requests count %d, actual %d", 0, actual)
	}
	if actual := reset.myPullRequestsCountKnown; actual {
		t.Fatalf("expected my pull requests count known %v, actual %v", false, actual)
	}
	if actual := reset.requestedPullRequestsCount; actual != 0 {
		t.Fatalf("expected requested pull requests count %d, actual %d", 0, actual)
	}
	if actual := reset.requestedPullRequestsCountKnown; actual {
		t.Fatalf("expected requested pull requests count known %v, actual %v", false, actual)
	}
	if actual := len(reset.additionalPullRequestsLoadStarted); actual != 0 {
		t.Fatalf("expected additional pull requests load-start map length %d, actual %d", 0, actual)
	}
	if actual := len(reset.additionalPullRequestsLoading); actual != 0 {
		t.Fatalf("expected additional pull requests loading map length %d, actual %d", 0, actual)
	}
	if actual := len(reset.additionalPullRequestsCounts); actual != 0 {
		t.Fatalf("expected additional pull requests count map length %d, actual %d", 0, actual)
	}

	if !subject.myPullRequestsLoadStarted || !subject.requestedPullRequestsLoadStarted {
		t.Fatal("expected the original load-start flags to stay true")
	}
	if !subject.additionalPullRequestsLoading[PullRequestTab(2)] {
		t.Fatal("expected the original additional-tab loading flag to stay true")
	}
	if actual := subject.additionalPullRequestsCounts[PullRequestTab(2)]; actual.count != 7 || !actual.known {
		t.Fatalf("expected the original additional-tab count state %+v, actual %+v", pullRequestCountState{count: 7, known: true}, actual)
	}
}
