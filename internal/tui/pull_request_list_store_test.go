package tui

import "testing"

func TestResetPullRequestListLoadState_GivenExistingStateAcrossDefaultAndAdditionalTabs_WhenResetting_ThenItClearsFlagsCountsAndTabMaps(t *testing.T) {
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

	subject.resetPullRequestListLoadState()

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
