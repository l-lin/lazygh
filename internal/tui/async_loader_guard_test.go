package tui

import (
	"os"
	"strings"
	"testing"
	"time"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestAsyncResultHandlers_GivenLoaderAndTimerFiles_WhenInspecting_ThenTheyDoNotMutateStateInsideUiUpdaterClosures(t *testing.T) {
	for _, path := range []string{
		"program_loading.go",
		"pull_request_detail_loader.go",
		"review_diff_loader.go",
		"workflow_commands.go",
		"notification_loading.go",
		"notification_detail_loader.go",
		"detail_image_loader.go",
		"loading_spinner.go",
		"error_popup.go",
	} {
		contents, actualErr := os.ReadFile(path)
		then_noError(t, actualErr)

		if strings.Contains(string(contents), "uiUpdater.Apply") {
			t.Fatalf("expected %q to dispatch async result messages instead of mutating state inside uiUpdater closures, actual source:\n%s", path, string(contents))
		}
	}
}

func TestUpdate_GivenMsgPullRequestDetailLoaded_WhenApplying_ThenItStoresTheLoadedDetailAndClearsInFlightState(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	summary := githubdomain.PullRequest{Title: "First PR", Number: 42, Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-05-24T12:00:00Z"}
	key := pullRequestDetailKey(summary.Repository, summary.Number)
	subject.pullRequestDetailLoadInFlight[key] = true
	subject.pullRequestDetailDocumentCache[pullRequestDetailDocumentCacheKey{pullRequestKey: key, tab: DescriptionDetailTab, width: 80}] = detailDocument{width: 80}
	expected := githubdomain.PullRequestDetail{Title: "Loaded PR", Number: 42, Body: "Loaded body"}

	Update(subject, MsgPullRequestDetailLoaded{Summary: summary, Detail: clonePullRequestDetail(expected)})

	if subject.pullRequestDetailLoadInFlight[key] {
		t.Fatalf("expected in-flight detail load for %q to be cleared", key)
	}
	actual, ok := subject.pullRequestDetailCache[key]
	if !ok {
		t.Fatalf("expected cached pull request detail for %q", key)
	}
	if actual.err != nil {
		t.Fatalf("expected no cached detail error, actual %v", actual.err)
	}
	if actual.detail.Title != expected.Title || actual.detail.Body != expected.Body {
		t.Fatalf("expected cached detail %+v, actual %+v", expected, actual.detail)
	}
	if actual.sourceUpdatedAt != pullRequestSummaryVersion(summary) {
		t.Fatalf("expected cached source version %q, actual %q", pullRequestSummaryVersion(summary), actual.sourceUpdatedAt)
	}
	if len(subject.pullRequestDetailDocumentCache) != 0 {
		t.Fatalf("expected detail documents to be invalidated after the reducer accepts the load result, actual %d entries", len(subject.pullRequestDetailDocumentCache))
	}
}

func TestUpdate_GivenMsgLoadingSpinnerTick_WhenApplying_ThenItOnlyAdvancesWhileLoadingWorkExists(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.notificationsLoading = true

	Update(subject, MsgLoadingSpinnerTick{})

	if actual := subject.startupState.loadingSpinnerFrameIndex; actual != 1 {
		t.Fatalf("expected loading spinner frame index %d after an active tick, actual %d", 1, actual)
	}

	subject.notificationsLoading = false
	Update(subject, MsgLoadingSpinnerTick{})

	if actual := subject.startupState.loadingSpinnerFrameIndex; actual != 1 {
		t.Fatalf("expected loading spinner frame index %d when no work remains, actual %d", 1, actual)
	}
}

func TestUpdate_GivenMsgTransientErrorPopupExpired_WhenApplying_ThenItClearsOnlyTheMatchingPopupGeneration(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.overlayState.transientErrorPopup = transientErrorPopupState{message: "boom", generation: 7, expiresAt: time.Now().Add(-time.Second)}

	Update(subject, MsgTransientErrorPopupExpired{Generation: 6})
	if actual := subject.overlayState.transientErrorPopup.message; actual != "boom" {
		t.Fatalf("expected popup message %q to stay visible for a stale expiry result, actual %q", "boom", actual)
	}

	Update(subject, MsgTransientErrorPopupExpired{Generation: 7})
	if subject.transientErrorPopupVisible() {
		t.Fatalf("expected the matching transient popup generation to be cleared, actual %+v", subject.overlayState.transientErrorPopup)
	}
}
