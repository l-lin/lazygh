package tui

import "testing"

func TestCalculateMainPaneLayout_GivenDefaultResizeState_WhenCalculatingCoordinates_ThenItKeepsTheNormalThreePaneSplit(t *testing.T) {
	actual := calculateMainPaneLayout(120, 30, PaneLayoutDefault, FocusUserView)

	then_paneVisibilityIs(t, actual, true, true, true)
	then_paneFrameIs(t, actual.user, 0, 0, 41, 2)
	then_paneFrameIs(t, actual.pullRequests, 0, 3, 41, 29)
	then_paneFrameIs(t, actual.detail, 42, 0, 119, 29)
}

func TestCalculateMainPaneLayout_GivenHalfWidthResizeState_WhenCalculatingCoordinates_ThenTheSidebarUsesHalfOfTheScreen(t *testing.T) {
	actual := calculateMainPaneLayout(120, 30, PaneLayoutHalfWidth, FocusPullRequestsView)

	then_paneVisibilityIs(t, actual, true, true, true)
	then_paneFrameIs(t, actual.user, 0, 0, 59, 2)
	then_paneFrameIs(t, actual.pullRequests, 0, 3, 59, 29)
	then_paneFrameIs(t, actual.detail, 60, 0, 119, 29)
}

func TestCalculateMainPaneLayout_GivenFullscreenResizeState_WhenCalculatingCoordinates_ThenOnlyTheFocusedSidePaneUsesTheWholeContentArea(t *testing.T) {
	actual := calculateMainPaneLayout(120, 30, PaneLayoutFullscreen, FocusPullRequestsView)

	then_paneVisibilityIs(t, actual, false, true, false)
	then_paneFrameIs(t, actual.pullRequests, 0, 0, 119, 29)
}

func TestCalculateMainPaneLayout_GivenDetailFullscreenResizeState_WhenCalculatingCoordinates_ThenOnlyTheDetailPaneUsesTheWholeContentArea(t *testing.T) {
	actual := calculateMainPaneLayout(120, 30, PaneLayoutFullscreen, FocusDetailView)

	then_paneVisibilityIs(t, actual, false, false, true)
	then_paneFrameIs(t, actual.detail, 0, 0, 119, 29)
}

func then_paneVisibilityIs(t *testing.T, actual mainPaneLayout, expectedUser bool, expectedPullRequests bool, expectedDetail bool) {
	t.Helper()

	if actual.userVisible != expectedUser {
		t.Fatalf("expected user visibility %t, actual %t", expectedUser, actual.userVisible)
	}
	if actual.pullRequestsVisible != expectedPullRequests {
		t.Fatalf("expected pull requests visibility %t, actual %t", expectedPullRequests, actual.pullRequestsVisible)
	}
	if actual.detailVisible != expectedDetail {
		t.Fatalf("expected detail visibility %t, actual %t", expectedDetail, actual.detailVisible)
	}
}

func then_paneFrameIs(t *testing.T, actual paneFrame, expectedX0 int, expectedY0 int, expectedX1 int, expectedY1 int) {
	t.Helper()

	if actual.x0 != expectedX0 || actual.y0 != expectedY0 || actual.x1 != expectedX1 || actual.y1 != expectedY1 {
		t.Fatalf(
			"expected pane frame (%d,%d)-(%d,%d), actual (%d,%d)-(%d,%d)",
			expectedX0,
			expectedY0,
			expectedX1,
			expectedY1,
			actual.x0,
			actual.y0,
			actual.x1,
			actual.y1,
		)
	}
}
