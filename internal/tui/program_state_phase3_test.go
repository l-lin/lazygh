package tui

import (
	"reflect"
	"testing"
)

func TestRefactorGuard_GivenProgramType_WhenInspecting_ThenDurableUIStateLivesInChildModels(t *testing.T) {
	programType := reflect.TypeOf(Program{})

	for _, fieldName := range []string{
		"loadingSpinnerFrameIndex",
		"detailWrapWidth",
		"activeDetailTab",
		"lastDetailIdentity",
		"detailViewState",
		"helpVisible",
		"transientErrorPopup",
		"errorMessages",
		"modalEditor",
		"reviewSession",
		"openedPullRequestSummary",
		"openedPullRequestTab",
		"pendingSelectionKeySequence",
		"pendingListViewportPlacements",
		"registeredKeybindingFingerprint",
		"keymapOverrides",
		"pullRequestSearches",
		"storyReviewConfig",
		"now",
		"after",
		"yankHighlightDuration",
		"transientErrorPopupDuration",
		"manualPullRequestListRefreshPending",
		"manualPullRequestDetailRefreshPending",
		"manualPullRequestDiffRefreshPending",
		"manualNotificationRefreshPending",
		"manualRefreshFeedback",
		"appStarted",
	} {
		if _, exists := programType.FieldByName(fieldName); exists {
			t.Fatalf("expected Program field %q to move under a child state model", fieldName)
		}
	}

	for _, fieldName := range []string{"startupState", "detailState", "overlayState", "navigationState", "runtimeConfig", "timingState", "manualRefreshState"} {
		if _, exists := programType.FieldByName(fieldName); !exists {
			t.Fatalf("expected Program to expose child state field %q", fieldName)
		}
	}

	if actual := programType.NumField(); actual > 45 {
		t.Fatalf("expected Program field count <= 45 after child-state extraction, actual %d", actual)
	}
}
