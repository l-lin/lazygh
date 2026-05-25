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

	expectedDirectFieldNames := []string{"model", "programDeps", "programStores", "programViewRuntime", "programShellRuntime"}
	actualDirectFieldNames := make([]string, 0, programType.NumField())
	for index := range programType.NumField() {
		actualDirectFieldNames = append(actualDirectFieldNames, programType.Field(index).Name)
	}
	if !reflect.DeepEqual(actualDirectFieldNames, expectedDirectFieldNames) {
		t.Fatalf("expected Program to segment shell state into %v, actual %v", expectedDirectFieldNames, actualDirectFieldNames)
	}
	if actual := programType.NumField(); actual > len(expectedDirectFieldNames) {
		t.Fatalf("expected Program direct field count <= %d after shell decomposition, actual %d", len(expectedDirectFieldNames), actual)
	}
}
