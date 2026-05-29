package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestNavigationStateModel_GivenOpenedPullRequestPinningTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	originalSummary := githubdomain.PullRequest{Number: 7, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, Title: "Old"}
	subject := navigationStateModel{openedPullRequestSummary: &originalSummary, openedPullRequestTab: RequestedPullRequestsTab}
	pinnedSummary := githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, Title: "Fresh"}

	pinned := subject.withOpenedPullRequestSummaryPinned(MyPullRequestsTab, pinnedSummary)
	pinnedSummary.Title = "Mutated afterwards"
	cleared := pinned.withOpenedPullRequestSummaryCleared()

	if pinned.openedPullRequestSummary == nil {
		t.Fatal("expected the pinned opened pull request summary to be present")
	}
	if actual := pinned.openedPullRequestSummary.Number; actual != 42 {
		t.Fatalf("expected pinned summary number %d, actual %d", 42, actual)
	}
	if actual := pinned.openedPullRequestSummary.Title; actual != "Fresh" {
		t.Fatalf("expected pinned summary title %q, actual %q", "Fresh", actual)
	}
	if actual := pinned.openedPullRequestTab; actual != MyPullRequestsTab {
		t.Fatalf("expected pinned opened pull request tab %v, actual %v", MyPullRequestsTab, actual)
	}
	if cleared.openedPullRequestSummary != nil {
		t.Fatalf("expected the cleared opened pull request summary to be nil, actual %+v", cleared.openedPullRequestSummary)
	}
	if actual := subject.openedPullRequestSummary.Title; actual != "Old" {
		t.Fatalf("expected the original summary title %q, actual %q", "Old", actual)
	}
	if actual := subject.openedPullRequestTab; actual != RequestedPullRequestsTab {
		t.Fatalf("expected the original opened pull request tab %v, actual %v", RequestedPullRequestsTab, actual)
	}
}

func TestNavigationStateModel_GivenPendingSelectionPrefixTransitions_WhenUpdating_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	originalTarget := keySequenceTargetFor(viewPullRequestsName, keymapScopeSelection, "move_selection_to_top")
	replacementTarget := keySequenceTargetFor(viewPullRequestsName, keymapScopeSelection, "recenter_selection")
	subject := navigationStateModel{pendingSelectionKeySequence: keySequenceState{pendingTarget: originalTarget}}

	armed := subject.withPendingSelectionKeySequenceArmed(replacementTarget)
	cleared := subject.withPendingSelectionKeySequenceCleared()

	if actual := armed.pendingSelectionKeySequenceTarget(); actual != replacementTarget {
		t.Fatalf("expected armed pending selection target %+v, actual %+v", replacementTarget, actual)
	}
	if actual := cleared.pendingSelectionKeySequenceTarget(); actual != (keySequenceTarget{}) {
		t.Fatalf("expected cleared pending selection target %+v, actual %+v", keySequenceTarget{}, actual)
	}
	if actual := subject.pendingSelectionKeySequenceTarget(); actual != originalTarget {
		t.Fatalf("expected the original pending selection target %+v, actual %+v", originalTarget, actual)
	}
}

func TestProgram_GivenLoadedRowsWithAMatchingOpenedPullRequest_WhenApplyingRows_ThenItPinsTheUpdatedSummary(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.navigationState = subject.navigationState.withOpenedPullRequestSummaryPinned(MyPullRequestsTab, githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, Title: "Stale"})
	freshSummary := githubdomain.PullRequest{Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, Title: "Fresh"}

	subject.applyLoadedPullRequestRows(MyPullRequestsTab, []githubdomain.PullRequest{freshSummary})

	actual, ok := subject.openedPullRequestSummaryForTab(MyPullRequestsTab)
	if !ok {
		t.Fatal("expected the opened pull request summary to stay pinned for the active tab")
	}
	if actual.Title != "Fresh" {
		t.Fatalf("expected the pinned summary title %q, actual %q", "Fresh", actual.Title)
	}
}
