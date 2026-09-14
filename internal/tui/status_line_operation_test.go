package tui

import (
	"errors"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestStatusStore_GivenTwoStatusOperations_WhenTheOlderOneFinishesFirst_ThenTheNewestOperationRemainsVisible(t *testing.T) {
	subject := *newStatusStore()

	first, firstID := subject.withStatusLineOperationStarted(statusLineOperationDescriptor{
		loadingLabel: "Refreshing #1: First",
		failureLabel: "Refreshing #1",
	})
	second, secondID := first.withStatusLineOperationStarted(statusLineOperationDescriptor{
		loadingLabel: "Approving #2: Second",
		failureLabel: "Approving #2",
	})

	actual := second.withStatusLineOperationFinished(firstID, errors.New("first failed"))

	if actual.statusLineOperation.id != secondID {
		t.Fatalf("expected the newest operation ID %d, actual %d", secondID, actual.statusLineOperation.id)
	}
	if actual.statusLineOperation.failureMessage != "" {
		t.Fatalf("expected no stale failure message, actual %q", actual.statusLineOperation.failureMessage)
	}
}

func TestProgram_GivenTwoStatusOperations_WhenTheOlderOneFailsAfterTheNewerOneStarts_ThenItDoesNotRecordTheStaleError(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	firstOperationID := subject.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: "Refreshing #1", failureLabel: "Refreshing #1"})
	subject.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: "Approving #2", failureLabel: "Approving #2"})

	subject.finishStatusLineOperation(firstOperationID, errors.New("first failed"))

	actual := subject.overlayState.errorMessages
	expected := []recordedErrorMessage(nil)
	if len(actual) != len(expected) {
		t.Fatalf("expected no recorded stale errors, actual %v", actual)
	}
}

func TestStatusStore_GivenTheNewestStatusOperation_WhenItFails_ThenItShowsTheFailureWithoutItsTitle(t *testing.T) {
	subject := *newStatusStore()
	started, operationID := subject.withStatusLineOperationStarted(statusLineOperationDescriptor{
		loadingLabel: "Approving #123: Fix login timeout",
		failureLabel: "Approving #123",
	})

	actual := started.withStatusLineOperationFinished(operationID, errors.New("permission denied"))
	expected := iconStatusFailure + " Approving #123: permission denied"

	if actual.statusLineOperation.failureMessage != expected {
		t.Fatalf("expected failure message %q, actual %q", expected, actual.statusLineOperation.failureMessage)
	}
	if actual.statusLineOperation.loadingLabel != "" {
		t.Fatalf("expected loading label %q, actual %q", "", actual.statusLineOperation.loadingLabel)
	}
}

func TestStatusLinePullRequestOperation_GivenAPullRequestSummary_WhenFormattingTheOperation_ThenItUsesTheNumberAndCleanTitle(t *testing.T) {
	actual := statusLinePullRequestOperation("Approving", githubdomain.PullRequest{Number: 123, Title: " Fix\nlogin timeout "})

	if actual.loadingLabel != "Approving #123: Fix login timeout" {
		t.Fatalf("expected loading label %q, actual %q", "Approving #123: Fix login timeout", actual.loadingLabel)
	}
	if actual.failureLabel != "Approving #123" {
		t.Fatalf("expected failure label %q, actual %q", "Approving #123", actual.failureLabel)
	}
}

func TestStatusLineManualRefreshOperation_GivenStandardReviewModeAndPullRequestRefresh_WhenResolving_ThenItUsesTheReviewLabel(t *testing.T) {
	actual := statusLineManualRefreshOperation(pullRequestRefreshSuccessMessage, true)

	if actual.loadingLabel != "Refreshing PR" || actual.failureLabel != "Refreshing PR" {
		t.Fatalf("expected review refresh labels %q, actual loading %q and failure %q", "Refreshing PR", actual.loadingLabel, actual.failureLabel)
	}
}

func TestStatusLineManualRefreshOperation_GivenBrowserModeAndPullRequestRefresh_WhenResolving_ThenItUsesThePullRequestListLabel(t *testing.T) {
	actual := statusLineManualRefreshOperation(pullRequestRefreshSuccessMessage, false)

	if actual.loadingLabel != "Refreshing PR list" || actual.failureLabel != "Refreshing PR list" {
		t.Fatalf("expected browser refresh labels %q, actual loading %q and failure %q", "Refreshing PR list", actual.loadingLabel, actual.failureLabel)
	}
}

func TestStatusLineManualRefreshOperation_GivenNotificationsRefresh_WhenResolving_ThenItUsesTheNotificationsLabel(t *testing.T) {
	actual := statusLineManualRefreshOperation(notificationsRefreshSuccessMessage, false)

	if actual.loadingLabel != "Refreshing notifications" || actual.failureLabel != "Refreshing notifications" {
		t.Fatalf("expected notification refresh labels %q, actual loading %q and failure %q", "Refreshing notifications", actual.loadingLabel, actual.failureLabel)
	}
}

func TestStatusLinePresenter_GivenTheNewestOperationHasFinished_WhenAnOlderLoadingFlagRemains_ThenItKeepsTheStatusLineClear(t *testing.T) {
	subject := statusLinePresenter{
		statusLineOperationStarted:           true,
		activePullRequestsLoadingText:        "Refreshing PR list",
		selectedPullRequestDetailLoadingText: "Refreshing #123: Older PR",
	}

	if actual := subject.Text(); actual != "" {
		t.Fatalf("expected the completed status line to stay clear, actual %q", actual)
	}
}

func TestUpdate_GivenAnOlderNotificationLoadCompletesAfterANewerMutationStarts_ThenItKeepsTheNewerStatusOperation(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.notificationsLoading = true
	olderOperationID := subject.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: "Refreshing notifications", failureLabel: "Refreshing notifications"})
	newerOperationID := subject.startStatusLineOperation(statusLineOperationDescriptor{loadingLabel: "Marking notification as done", failureLabel: "Marking notification as done"})

	Update(subject, MsgNotificationsLoaded{StatusLineOperationID: olderOperationID, Err: errors.New("older load failed")})

	if subject.statusStore.statusLineOperation.id != newerOperationID {
		t.Fatalf("expected newer operation ID %d to remain active, actual %d", newerOperationID, subject.statusStore.statusLineOperation.id)
	}
	if actual := subject.statusLineOperationLoadingStatus(); actual != "Marking notification as done" {
		t.Fatalf("expected newer operation label %q, actual %q", "Marking notification as done", actual)
	}
}

func TestStatusLineAvailableTextWidth_GivenKeyHintsUseAllAvailableSpace_WhenCalculating_ThenItReturnsNoStatusWidth(t *testing.T) {
	actual := statusLineAvailableTextWidth(20, 20)

	if actual != 0 {
		t.Fatalf("expected no status width, actual %d", actual)
	}
}

func TestTruncateStatusLineText_GivenTextWiderThanTheAvailableWidth_WhenTruncating_ThenItKeepsTheFriendlyPrefixAndAddsAnEllipsis(t *testing.T) {
	actual := truncateStatusLineText("Refreshing #123: Fix login timeout", 19)
	expected := "Refreshing #123: F…"

	if actual != expected {
		t.Fatalf("expected truncated status %q, actual %q", expected, actual)
	}
}

func TestStatusStore_GivenACompletedSuccessfulOperation_WhenPresentingTheStatusLine_ThenItRetainsTheSuccessLabel(t *testing.T) {
	subject := *newStatusStore()
	started, operationID := subject.withStatusLineOperationStarted(statusLineOperationDescriptor{loadingLabel: "Auto-refreshing PRs at 22:08"})
	actual := started.withStatusLineOperationFinished(operationID, nil)

	expected := iconStatusSuccess + " Auto-refreshing PRs at 22:08"
	if actual.statusLineOperation.successMessage != expected {
		t.Fatalf("expected retained success %q, actual %q", expected, actual.statusLineOperation.successMessage)
	}
	if actual.statusLineOperationStarted {
		t.Fatal("expected the completed operation to stop being active")
	}
}

func TestStatusLinePresenter_GivenInformationalFeedbackAndAnActiveOperation_WhenPresenting_ThenTheOperationSpinnerTakesPrecedence(t *testing.T) {
	subject := statusLinePresenter{
		feedbackMessage:                   "Saved",
		statusLineOperationLoadingMessage: "Auto-refreshing PRs at 22:08",
		statusLineOperationStarted:        true,
		loadingSpinner:                    string(loadingSpinnerFrames[0]),
	}

	if actual := subject.Text(); actual != string(loadingSpinnerFrames[0])+" Auto-refreshing PRs at 22:08" {
		t.Fatalf("expected the active operation to win, actual %q", actual)
	}
}

func TestProgram_GivenSuccessFeedbackAlreadyHasTheSuccessIcon_WhenSettingIt_ThenItDoesNotDuplicateTheIcon(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.setSuccessFeedback(FocusPullRequestsView, iconStatusSuccess+" Saved")

	if actual := subject.feedbackMessage; actual != iconStatusSuccess+" Saved" {
		t.Fatalf("expected one success icon, actual %q", actual)
	}
}

func TestManualRefresh_GivenMultiplePendingOperations_WhenTheFirstCompletes_ThenTheAggregateOperationStaysActive(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	subject.beginManualRefresh("Pull requests refreshed", 2)

	subject.completeManualRefreshOperation(nil)
	if !subject.statusLineOperationStarted() {
		t.Fatal("expected the aggregate operation to stay active")
	}
	if subject.statusLineOperationSuccessStatus() != "" {
		t.Fatalf("expected no retained success before the final child, actual %q", subject.statusLineOperationSuccessStatus())
	}

	subject.completeManualRefreshOperation(nil)
	if subject.statusLineOperationStarted() {
		t.Fatal("expected the aggregate operation to finish after the final child")
	}
	if actual := subject.feedbackMessage; actual != iconStatusSuccess+" Pull requests refreshed" {
		t.Fatalf("expected aggregate success feedback, actual %q", actual)
	}
}
