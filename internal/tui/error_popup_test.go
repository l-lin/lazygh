package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenApprovePullRequestFailure_WhenRendering_ThenItShowsATransientErrorPopupAtTheBottomRight(t *testing.T) {
	loader := &fakePullRequestDetailLoader{approveErr: errors.New("GitHub rejected the approval because a reviewer cannot approve their own pull request")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	gui := given_headlessGuiWithSize(t, 120, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	subject.model.UpdateActionsPopupSearch(pullRequestReviewApprovalTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestReviewApprovalTitle))
	then_noError(t, subject.afterStateChange(gui))

	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))
	given_runQueuedAsync(t, asyncRunner, 0)

	toastView, actualErr := gui.View(viewTransientErrorPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(toastView.Buffer(), "GitHub rejected the approval") {
		t.Fatalf("expected the transient error popup to contain %q, actual %q", "GitHub rejected the approval", toastView.Buffer())
	}
	then_transientErrorPopupIsBottomRightAboveStatusLine(t, gui)
}

func TestTransientErrorPopupActionError_GivenAWrappedError_WhenResolvingItsMessage_ThenItReturnsTheNormalizedPopupMessage(t *testing.T) {
	message, ok := transientErrorPopupActionMessage(newTransientErrorPopupActionError(errors.New("run `gh pr merge 42 -R acme/widgets --squash`: exit status 1: boom")))

	if !ok {
		t.Fatal("expected the wrapped error to report a popup message")
	}
	if message != "boom" {
		t.Fatalf("expected popup message %q, actual %q", "boom", message)
	}
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
	then_noError(t, subject.afterStateChange(gui))

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

func TestActionsPopup_GivenRecordedErrors_WhenOpening_ThenItShowsTheRecentErrorsAction(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.reportError(nil, "First error")
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), recentErrorsActionTitle) {
		t.Fatalf("expected the actions popup to contain %q, actual %q", recentErrorsActionTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenRecordedErrors_WhenExecutingTheRecentErrorsAction_ThenItOpensTheHistoryPopupWithNewestErrorsFirst(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.reportError(nil, "First error")
	subject.reportError(nil, "Second error")
	gui := given_headlessGuiWithSize(t, 120, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	subject.model.UpdateActionsPopupSearch(recentErrorsActionTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), recentErrorsActionTitle))
	then_noError(t, subject.afterStateChange(gui))

	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))

	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	popupView, actualErr := gui.View(viewPullRequestBuildInfoName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, recentErrorsPopupTitle) {
		t.Fatalf("expected the popup title to contain %q, actual %q", recentErrorsPopupTitle, popupView.Title)
	}
	if strings.Index(popupView.Buffer(), "Second error") < 0 || strings.Index(popupView.Buffer(), "First error") < 0 {
		t.Fatalf("expected the popup buffer to contain both recorded errors, actual %q", popupView.Buffer())
	}
	if strings.Index(popupView.Buffer(), "Second error") > strings.Index(popupView.Buffer(), "First error") {
		t.Fatalf("expected the newest error to render first, actual %q", popupView.Buffer())
	}
	then_viewOccupiesAtLeastPercentOfScreen(t, gui, viewPullRequestBuildInfoName, 90, 90)
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

func then_transientErrorPopupContains(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	toastView, actualErr := gui.View(viewTransientErrorPopupName)
	then_noError(t, actualErr)
	toastText := strings.ReplaceAll(strings.Join(toastView.BufferLines(), ""), "\n", "")
	if !strings.Contains(toastText, expected) {
		t.Fatalf("expected the transient error popup to contain %q, actual %q", expected, toastText)
	}
}

func then_statusLineDoesNotContain(t *testing.T, gui *gocui.Gui, unexpected string) {
	t.Helper()

	statusView, actualErr := gui.View(viewStatusLineName)
	then_noError(t, actualErr)
	if strings.Contains(statusView.Buffer(), unexpected) {
		t.Fatalf("expected status line to hide %q, actual %q", unexpected, statusView.Buffer())
	}
}
