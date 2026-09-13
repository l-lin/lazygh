package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgErrorReported_WhenApplying_ThenItRecordsTheMessageTimestampShowsThePopupAndReturnsATypedExpiryCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	now := time.Date(2026, time.May, 20, 12, 34, 56, 0, time.UTC)
	subject.timingState.now = func() time.Time { return now }

	actual := Update(subject, MsgErrorReported{Message: "boom"})

	if len(actual) != 1 {
		t.Fatalf("expected one transient-popup expiry command, actual %d", len(actual))
	}
	if _, ok := actual[0].(transientErrorPopupExpiryCmd); !ok {
		t.Fatalf("expected a transientErrorPopupExpiryCmd, actual %T", actual[0])
	}
	if actualMessage := subject.overlayState.transientErrorPopup.message; actualMessage != "boom" {
		t.Fatalf("expected transient popup message %q, actual %q", "boom", actualMessage)
	}
	expectedRecord := recordedErrorMessage{message: "boom", timestamp: now}
	if len(subject.overlayState.errorMessages) != 1 || subject.overlayState.errorMessages[0] != expectedRecord {
		t.Fatalf("expected recorded error %+v, actual %v", expectedRecord, subject.overlayState.errorMessages)
	}
}

func TestUpdate_GivenMsgErrorReportedWithZeroDuration_WhenApplying_ThenItShowsThePopupWithoutSchedulingExpiry(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.timingState.transientErrorPopupDuration = 0

	actual := Update(subject, MsgErrorReported{Message: "boom"})

	if len(actual) != 0 {
		t.Fatalf("expected no expiry commands, actual %d", len(actual))
	}
	if actualMessage := subject.overlayState.transientErrorPopup.message; actualMessage != "boom" {
		t.Fatalf("expected transient popup message %q, actual %q", "boom", actualMessage)
	}
}

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

	if _, actualErr := gui.View(viewTransientErrorPopupName); actualErr == nil {
		t.Fatal("expected the transient error popup to stay hidden")
	}
	then_statusLineContains(t, gui, iconStatusFailure)
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
	subject.timingState.now = func() time.Time { return currentTime }
	subject.timingState.after = func(time.Duration) <-chan time.Time { return delay }
	gui := given_headlessGuiWithSize(t, 120, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	given_transientErrorReported(subject, gui, "boom")
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

func TestAfterStateChange_GivenAnExpiredTransientErrorPopup_WhenRefreshing_ThenItKeepsThePopupVisibleUntilTheExpiryMessageArrives(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	currentTime := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	subject.timingState.now = func() time.Time { return currentTime }
	gui := given_headlessGuiWithSize(t, 120, 30)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	given_transientErrorReported(subject, gui, "boom")
	currentTime = currentTime.Add(defaultTransientErrorPopupDuration + time.Millisecond)

	then_noError(t, subject.afterStateChange(gui))

	then_transientErrorPopupContains(t, gui, "boom")
}

func TestScreenLayout_GivenATransientErrorPopup_WhenPlanningOverlays_ThenItPinsThePopupAboveTheStatusLineAtTheBottomRight(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	given_transientErrorReported(subject, nil, "boom")

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
	given_transientErrorReported(subject, nil, "First error")
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

func TestActionsPopup_GivenRecordedErrors_WhenExecutingTheRecentErrorsAction_ThenItOpensTheHistoryPopupInChronologicalOrderWithTimestampedErrors(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	currentTime := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	subject.timingState.now = func() time.Time { return currentTime }
	given_transientErrorReported(subject, nil, "First error")
	currentTime = currentTime.Add(time.Minute)
	given_transientErrorReported(subject, nil, "Second error")
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
	expectedBody := "12:00:00 First error\n12:01:00 Second error"
	if actual := subject.pullRequestBuildRunPopup.body; actual != expectedBody {
		t.Fatalf("expected chronological timestamped body %q, actual %q", expectedBody, actual)
	}
	if strings.Index(popupView.Buffer(), "Second error") < strings.Index(popupView.Buffer(), "First error") {
		t.Fatalf("expected the oldest error to render first, actual %q", popupView.Buffer())
	}
	then_viewOccupiesAtLeastPercentOfScreen(t, gui, viewPullRequestBuildInfoName, 90, 90)
}

func TestRenderRecordedErrorMessages_GivenMultipleRecords_WhenRendering_ThenItUsesAppendOrderAndTimePrefixes(t *testing.T) {
	records := []recordedErrorMessage{
		{message: "First error", timestamp: time.Date(2026, time.May, 20, 9, 1, 2, 0, time.UTC)},
		{message: "Second error", timestamp: time.Date(2026, time.May, 20, 9, 3, 4, 0, time.UTC)},
	}
	expected := "09:01:02 First error\n09:03:04 Second error"

	actual := renderRecordedErrorMessages(records)

	if actual != expected {
		t.Fatalf("expected chronological rendered errors %q, actual %q", expected, actual)
	}
}

func TestRenderRecordedErrorMessages_GivenMultilineRecord_WhenRendering_ThenItPrefixesOnlyTheFirstLineAndSeparatesEntries(t *testing.T) {
	records := []recordedErrorMessage{
		{message: "First line\nsecond line", timestamp: time.Date(2026, time.May, 20, 10, 11, 12, 0, time.UTC)},
		{message: "Next error", timestamp: time.Date(2026, time.May, 20, 10, 13, 14, 0, time.UTC)},
	}
	expected := "10:11:12 First line\nsecond line\n10:13:14 Next error"

	actual := renderRecordedErrorMessages(records)

	if actual != expected {
		t.Fatalf("expected multiline rendered errors %q, actual %q", expected, actual)
	}
}

func TestRenderRecordedErrorMessages_GivenNoRecords_WhenRendering_ThenItReturnsAnEmptyString(t *testing.T) {
	actual := renderRecordedErrorMessages(nil)

	if actual != "" {
		t.Fatalf("expected an empty rendered body, actual %q", actual)
	}
}

func TestRecordedErrorMessagesWithAppended_GivenAFullHistory_WhenAppendingARecord_ThenItRetainsTheNewestHundredInReportOrder(t *testing.T) {
	existing := make([]recordedErrorMessage, maxRecordedErrorMessages)
	for index := range existing {
		existing[index] = recordedErrorMessage{
			message:   fmt.Sprintf("record-%d", index),
			timestamp: time.Date(2026, time.May, 20, 0, 0, index, 0, time.UTC),
		}
	}
	newTimestamp := time.Date(2026, time.May, 20, 1, 0, 0, 0, time.UTC)

	actual := recordedErrorMessagesWithAppended(existing, "new record", newTimestamp)

	if len(actual) != maxRecordedErrorMessages {
		t.Fatalf("expected %d records, actual %d", maxRecordedErrorMessages, len(actual))
	}
	if actual[0].message != "record-1" {
		t.Fatalf("expected the oldest retained record %q, actual %q", "record-1", actual[0].message)
	}
	if actual[len(actual)-1] != (recordedErrorMessage{message: "new record", timestamp: newTimestamp}) {
		t.Fatalf("expected the newest appended record %+v, actual %+v", recordedErrorMessage{message: "new record", timestamp: newTimestamp}, actual[len(actual)-1])
	}
	for index, record := range actual[:len(actual)-1] {
		expectedMessage := fmt.Sprintf("record-%d", index+1)
		if record.message != expectedMessage {
			t.Fatalf("expected record %d to be %q, actual %q", index, expectedMessage, record.message)
		}
	}
}

func given_transientErrorReported(subject *Program, gui *gocui.Gui, message string) {
	subject.executeCmds(gui, Update(subject, MsgErrorReported{Message: message}))
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
	if actualErr != nil {
		then_statusLineContains(t, gui, iconStatusFailure)
		return
	}
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
