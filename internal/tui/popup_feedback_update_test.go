package tui

import (
	"errors"
	"testing"
)

func TestUpdate_GivenMsgActionsPopupClosedWithFeedback_WhenApplying_ThenItClosesThePopupAndShowsStatusFeedback(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(3)
	subject.actionsPopupWidget.searchEditor = newLineEditor("browser")
	subject.actionsPopupWidget.errorMessage = "stale"

	Update(subject, MsgActionsPopupClosedWithFeedback{Target: FocusDetailView, Message: "done"})

	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after the success message")
	}
	if subject.actionsPopupWidget.hasSearchEditor() {
		t.Fatal("expected the popup search editor to be cleared after the success message")
	}
	if actual := subject.actionsPopupWidget.errorMessage; actual != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actual)
	}
	if actual := subject.feedbackMessage; actual != "done" {
		t.Fatalf("expected feedback %q, actual %q", "done", actual)
	}
}

func TestUpdate_GivenMsgActionsPopupActionErrorHandledWithStatusLineError_WhenApplying_ThenItKeepsThePopupOpenAndShowsStatusFeedback(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(3)
	subject.actionsPopupWidget.errorMessage = "stale"

	Update(subject, MsgActionsPopupActionErrorHandled{Err: newActionsPopupStatusLineError(FocusDetailView, errors.New("boom"))})

	if !subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to stay open after a status-line validation error")
	}
	if actual := subject.actionsPopupWidget.errorMessage; actual != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actual)
	}
	if actual := subject.feedbackMessage; actual != "boom" {
		t.Fatalf("expected feedback %q, actual %q", "boom", actual)
	}
}

func TestUpdate_GivenMsgPendingPullRequestReviewSubmitted_WhenApplying_ThenItRestoresBrowserModeInvalidatesCachesAndShowsFeedback(t *testing.T) {
	summary := given_pullRequestMutationSummary("OPEN", false)
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.FocusPullRequestsView()
	subject.startReviewSession(summary, "PRR_pending")
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: subject.optimisticPullRequestDetailSeed(summary)}
	subject.pullRequestDiffCache["acme/widgets#42"] = pullRequestDiffResult{data: reviewDiffData{}}
	subject.reviewDiffRenderCache[reviewDiffRenderCacheKey{repositoryName: "acme/widgets", pullRequestNumber: 42, filePath: "main.go", width: 80}] = reviewDiffRenderCacheEntry{}

	Update(subject, MsgPendingPullRequestReviewSubmitted{Target: pendingPullRequestReviewTarget{repository: "acme/widgets", number: 42, pendingReviewID: "PRR_pending", sourceFocus: FocusPullRequestsView}})

	if subject.navigationState.reviewSession.active {
		t.Fatal("expected review mode to be inactive after submitting the pending review")
	}
	if _, ok := subject.pullRequestDetailCache["acme/widgets#42"]; ok {
		t.Fatal("expected the pull request detail cache to be invalidated after submitting the pending review")
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the pull request diff cache to be invalidated after submitting the pending review")
	}
	if state, ok := subject.pendingPullRequestReviewCache["acme/widgets#42"]; !ok || state.id != "" {
		t.Fatalf("expected pending review state to be cleared, actual %+v, known=%v", state, ok)
	}
	if len(subject.reviewDiffRenderCache) != 0 {
		t.Fatalf("expected the review diff render cache to be cleared, actual %d entries", len(subject.reviewDiffRenderCache))
	}
	if actual := subject.feedbackMessage; actual != pullRequestReviewSuccessMessage {
		t.Fatalf("expected feedback %q, actual %q", pullRequestReviewSuccessMessage, actual)
	}
	if actual := subject.model.Focus(); actual != FocusPullRequestsView {
		t.Fatalf("expected focus %v, actual %v", FocusPullRequestsView, actual)
	}
}
