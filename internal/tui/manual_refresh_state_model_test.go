package tui

import (
	"errors"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestManualRefreshStateModel_GivenPendingTargets_WhenMarkingAndConsuming_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := manualRefreshStateModel{
		pullRequestListPending:   map[PullRequestTab]bool{PullRequestTab(0): true},
		pullRequestDetailPending: map[string]bool{"acme/widgets#7": true},
		pullRequestDiffPending:   map[string]bool{"acme/widgets#8": true},
	}
	summary := githubdomain.PullRequest{Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"}, Number: 42}

	markedListState, actualListMarked := subject.withPullRequestListPending(PullRequestTab(1))
	markedDetailState, actualDetailMarked := subject.withPullRequestDetailPending(summary)
	markedDiffState, actualDiffMarked := subject.withPullRequestDiffPending(summary)
	markedNotificationState, actualNotificationMarked := subject.withNotificationPending()
	consumedListState, actualListConsumed := markedListState.withoutPullRequestListPending(PullRequestTab(1))
	consumedDetailState, actualDetailConsumed := markedDetailState.withoutPullRequestDetailPending("acme/widgets#42")
	consumedDiffState, actualDiffConsumed := markedDiffState.withoutPullRequestDiffPending("acme/widgets#42")
	consumedNotificationState, actualNotificationConsumed := markedNotificationState.withoutNotificationPending()

	if !actualListMarked || !markedListState.pullRequestListPending[PullRequestTab(1)] {
		t.Fatalf("expected the merged tab to be marked for refresh, actual %v", markedListState.pullRequestListPending)
	}
	if !actualDetailMarked || !markedDetailState.pullRequestDetailPending["acme/widgets#42"] {
		t.Fatalf("expected the detail key to be marked for refresh, actual %v", markedDetailState.pullRequestDetailPending)
	}
	if !actualDiffMarked || !markedDiffState.pullRequestDiffPending["acme/widgets#42"] {
		t.Fatalf("expected the diff key to be marked for refresh, actual %v", markedDiffState.pullRequestDiffPending)
	}
	if !actualNotificationMarked || !markedNotificationState.notificationPending {
		t.Fatal("expected notifications to be marked for refresh")
	}
	if !actualListConsumed || consumedListState.pullRequestListPending[PullRequestTab(1)] {
		t.Fatalf("expected the merged tab mark to be consumed, actual %v", consumedListState.pullRequestListPending)
	}
	if !actualDetailConsumed || consumedDetailState.pullRequestDetailPending["acme/widgets#42"] {
		t.Fatalf("expected the detail key to be consumed, actual %v", consumedDetailState.pullRequestDetailPending)
	}
	if !actualDiffConsumed || consumedDiffState.pullRequestDiffPending["acme/widgets#42"] {
		t.Fatalf("expected the diff key to be consumed, actual %v", consumedDiffState.pullRequestDiffPending)
	}
	if !actualNotificationConsumed || consumedNotificationState.notificationPending {
		t.Fatal("expected the notification mark to be consumed")
	}
	if subject.pullRequestListPending[PullRequestTab(1)] {
		t.Fatalf("expected the original list map to stay unchanged, actual %v", subject.pullRequestListPending)
	}
	if subject.pullRequestDetailPending["acme/widgets#42"] {
		t.Fatalf("expected the original detail map to stay unchanged, actual %v", subject.pullRequestDetailPending)
	}
	if subject.pullRequestDiffPending["acme/widgets#42"] {
		t.Fatalf("expected the original diff map to stay unchanged, actual %v", subject.pullRequestDiffPending)
	}
	if subject.notificationPending {
		t.Fatal("expected the original notification flag to stay unchanged")
	}
}

func TestManualRefreshStateModel_GivenSuccessFeedback_WhenBeginningAndCompleting_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := manualRefreshStateModel{}

	startedState, actualStarted := subject.withFeedbackBegun(" refreshed ", 1)
	completedState, actualCompletion, actualClearFeedback := startedState.withCompletedOperation(nil)

	if !actualStarted {
		t.Fatal("expected the feedback state to begin")
	}
	if startedState.feedback == nil {
		t.Fatal("expected the started state to carry manual-refresh feedback")
	}
	if actualCompletion.successMessage != "refreshed" {
		t.Fatalf("expected the completion success message %q, actual %q", "refreshed", actualCompletion.successMessage)
	}
	if actualCompletion.popupError != "" {
		t.Fatalf("expected the popup error %q, actual %q", "", actualCompletion.popupError)
	}
	if actualClearFeedback {
		t.Fatal("expected the success path to keep the existing status line until update handles completion")
	}
	if completedState.feedback != nil {
		t.Fatalf("expected the completed state to clear feedback, actual %+v", completedState.feedback)
	}
	if subject.feedback != nil {
		t.Fatalf("expected the original state to stay unchanged, actual %+v", subject.feedback)
	}
	if startedState.feedback.pendingOperations != 1 || startedState.feedback.successMessage != "refreshed" {
		t.Fatalf("expected the started feedback to stay unchanged before completion, actual %+v", startedState.feedback)
	}
}

func TestManualRefreshStateModel_GivenFailureFeedback_WhenCompletingOperations_ThenItReturnsUpdatedCopiesWithoutMutatingTheOriginal(t *testing.T) {
	subject := manualRefreshStateModel{}
	startedState, actualStarted := subject.withFeedbackBegun("refreshed", 2)
	failedState, actualFirstCompletion, actualFirstClearFeedback := startedState.withCompletedOperation(errors.New("boom"))
	completedState, actualSecondCompletion, actualSecondClearFeedback := failedState.withCompletedOperation(nil)

	if !actualStarted {
		t.Fatal("expected the feedback state to begin")
	}
	if !actualFirstClearFeedback {
		t.Fatal("expected the first failure to request clearing the current status line feedback")
	}
	if actualFirstCompletion.popupError != "boom" {
		t.Fatalf("expected the popup error %q, actual %q", "boom", actualFirstCompletion.popupError)
	}
	if actualFirstCompletion.successMessage != "" {
		t.Fatalf("expected the first completion success message %q, actual %q", "", actualFirstCompletion.successMessage)
	}
	if failedState.feedback == nil || failedState.feedback.pendingOperations != 1 || !failedState.feedback.failed {
		t.Fatalf("expected the failed state to keep one pending failed feedback operation, actual %+v", failedState.feedback)
	}
	if actualSecondClearFeedback {
		t.Fatal("expected later completions to avoid clearing the feedback again")
	}
	if actualSecondCompletion.popupError != "" || actualSecondCompletion.successMessage != "" {
		t.Fatalf("expected the terminal failure completion to stay silent, actual %+v", actualSecondCompletion)
	}
	if completedState.feedback != nil {
		t.Fatalf("expected the terminal failure to clear feedback, actual %+v", completedState.feedback)
	}
	if startedState.feedback == nil || startedState.feedback.pendingOperations != 2 || startedState.feedback.failed {
		t.Fatalf("expected the original started feedback to stay unchanged, actual %+v", startedState.feedback)
	}
}
