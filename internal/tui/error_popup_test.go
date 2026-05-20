package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
)

func TestActionsPopup_GivenApprovePullRequestFailure_WhenRendering_ThenItShowsATransientErrorPopupAtTheBottomRight(t *testing.T) {
	loader := &fakePullRequestDetailLoader{approveErr: errors.New("GitHub rejected the approval because a reviewer cannot approve their own pull request")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = &capturingAsyncRunner{}
	gui := given_headlessGuiWithSize(t, 120, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	subject.model.UpdateActionsPopupSearch(pullRequestReviewApprovalTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestReviewApprovalTitle))
	then_noError(t, subject.refreshViews(gui))

	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))

	toastView, actualErr := gui.View(viewTransientErrorPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(toastView.Buffer(), "GitHub rejected the approval") {
		t.Fatalf("expected the transient error popup to contain %q, actual %q", "GitHub rejected the approval", toastView.Buffer())
	}
	then_transientErrorPopupIsBottomRightAboveStatusLine(t, gui)
}

func TestTransientErrorPopup_GivenAVisibleError_WhenItsLifetimeExpires_ThenItDisappearsAfterTheScheduledRefresh(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	asyncRunner := &capturingAsyncRunner{}
	delay := make(chan time.Time, 1)
	currentTime := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.now = func() time.Time { return currentTime }
	subject.after = func(time.Duration) <-chan time.Time { return delay }
	gui := given_headlessGuiWithSize(t, 120, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	subject.reportError(gui, "boom")
	then_noError(t, subject.refreshViews(gui))

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one scheduled transient popup cleanup, actual %d", len(asyncRunner.runs))
	}
	if _, actualErr := gui.View(viewTransientErrorPopupName); actualErr != nil {
		then_noError(t, actualErr)
	}

	currentTime = currentTime.Add(defaultTransientErrorPopupDuration)
	delay <- currentTime
	given_runQueuedAsync(t, asyncRunner, 0)

	then_viewDoesNotExist(t, gui, viewTransientErrorPopupName)
}

func TestScreenLayout_GivenATransientErrorPopup_WhenPlanningOverlays_ThenItPinsThePopupAboveTheStatusLineAtTheBottomRight(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.reportError(nil, "boom")

	actual := subject.screenLayoutForSize(100, 30)
	frame, ok := actual.OverlayFrame(viewTransientErrorPopupName)
	if !ok {
		t.Fatalf("expected overlay frame %q", viewTransientErrorPopupName)
	}
	if !frame.Visible {
		t.Fatal("expected the transient error popup to stay visible while the error is active")
	}
	if frame.Frame.x1 != actual.StatusLine.Frame.x1-1 {
		t.Fatalf("expected transient error popup right edge %d, actual %d", actual.StatusLine.Frame.x1-1, frame.Frame.x1)
	}
	if frame.Frame.y1 != actual.StatusLine.Frame.y0-1 {
		t.Fatalf("expected transient error popup bottom edge %d, actual %d", actual.StatusLine.Frame.y0-1, frame.Frame.y1)
	}
}

func then_transientErrorPopupIsBottomRightAboveStatusLine(t *testing.T, gui *gocui.Gui) {
	t.Helper()

	statusX0, statusY0, statusX1, _, actualErr := gui.ViewPosition(viewStatusLineName)
	then_noError(t, actualErr)
	popupX0, popupY0, popupX1, popupY1, actualErr := gui.ViewPosition(viewTransientErrorPopupName)
	then_noError(t, actualErr)
	if popupX1 != statusX1-1 {
		t.Fatalf("expected transient error popup right edge %d, actual %d", statusX1-1, popupX1)
	}
	if popupY1 >= statusY0 {
		t.Fatalf("expected transient error popup bottom edge above the status line y=%d, actual y=%d", statusY0, popupY1)
	}
	if popupX0 < statusX0 {
		t.Fatalf("expected transient error popup left edge >= %d, actual %d", statusX0, popupX0)
	}
	if popupY0 < 0 {
		t.Fatalf("expected transient error popup top edge >= 0, actual %d", popupY0)
	}
}
