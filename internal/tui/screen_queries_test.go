package tui

import "testing"

func TestReviewScreenMode_GivenBrowserAndReviewModes_WhenCheckingVisibility_ThenOnlyReviewAndStoryReviewCountAsReview(t *testing.T) {
	if isReviewScreenMode(ScreenModeBrowser) {
		t.Fatal("expected browser mode to stay outside review mode")
	}
	if !isReviewScreenMode(ScreenModeReview) {
		t.Fatal("expected review mode to count as review")
	}
	if !isReviewScreenMode(ScreenModeStoryReview) {
		t.Fatal("expected story-review mode to count as review")
	}
}

func TestNotificationLoadEligible_GivenBrowserAndReviewModes_WhenCheckingEligibility_ThenOnlyBrowserNotificationsLoad(t *testing.T) {
	if !notificationLoadEligible(ScreenModeBrowser, FocusNotificationsView) {
		t.Fatal("expected browser notifications to stay loadable")
	}
	if notificationLoadEligible(ScreenModeReview, FocusNotificationsView) {
		t.Fatal("expected review mode to block notification loads")
	}
	if notificationLoadEligible(ScreenModeStoryReview, FocusNotificationsView) {
		t.Fatal("expected story-review mode to block notification loads")
	}
}

func TestPullRequestDiffLoadEligible_GivenBrowserReviewAndTabs_WhenCheckingEligibility_ThenReviewOrChangesTabCanLoadDiffs(t *testing.T) {
	if !pullRequestDiffLoadEligible(ScreenModeReview, false, DescriptionDetailTab) {
		t.Fatal("expected review mode to keep diff loading enabled")
	}
	if !pullRequestDiffLoadEligible(ScreenModeBrowser, true, ChangesDetailTab) {
		t.Fatal("expected browser changes tab to load diffs")
	}
	if pullRequestDiffLoadEligible(ScreenModeBrowser, true, DescriptionDetailTab) {
		t.Fatal("expected browser description tab to avoid diff loads")
	}
}

func TestReviewTreeSearchEligible_GivenModeAndSearchInputs_WhenCheckingRouting_ThenOnlyReviewTreeSearchRepeatsInReviewModes(t *testing.T) {
	if !reviewTreeSearchEligible(ScreenModeReview, true, FocusPullRequestsView) {
		t.Fatal("expected review tree search repeat to stay enabled in review mode")
	}
	if reviewTreeSearchEligible(ScreenModeBrowser, true, FocusPullRequestsView) {
		t.Fatal("expected browser search repeat to stay out of review-tree routing")
	}
	if reviewTreeSearchEligible(ScreenModeReview, true, FocusDetailView) {
		t.Fatal("expected detail search repeat to stay out of review-tree routing")
	}
}

func TestReviewTreeFoldEligible_GivenModeFocusAndBlockers_WhenCheckingAvailability_ThenOnlyReviewFileTreesCanBulkFold(t *testing.T) {
	if !reviewTreeFoldEligible(ScreenModeReview, FocusPullRequestsView, false) {
		t.Fatal("expected review file tree folds to stay enabled")
	}
	if reviewTreeFoldEligible(ScreenModeBrowser, FocusPullRequestsView, false) {
		t.Fatal("expected browser mode to avoid review-tree folds")
	}
	if reviewTreeFoldEligible(ScreenModeReview, FocusPullRequestsView, true) {
		t.Fatal("expected fold blockers to win")
	}
}

func TestBrowserSideFocus_GivenLastSideFocus_WhenRestoring_ThenItKeepsOnlyExplicitBrowserSideViews(t *testing.T) {
	if actual := browserSideFocus(FocusNotificationsView); actual != FocusNotificationsView {
		t.Fatalf("expected notifications focus %v, actual %v", FocusNotificationsView, actual)
	}
	if actual := browserSideFocus(FocusDetailView); actual != FocusUserView {
		t.Fatalf("expected invalid focus to fall back to %v, actual %v", FocusUserView, actual)
	}
}
