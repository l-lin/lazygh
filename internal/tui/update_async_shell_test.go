package tui

import (
	"errors"
	"testing"
)

func TestUpdate_GivenMsgActionsPopupAsyncGHCommandFinishedWithError_WhenApplying_ThenItReturnsATypedReportErrorCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.ghCommandLoadingMessage = "Running `gh pr ready`."

	actual := Update(subject, MsgActionsPopupAsyncGHCommandFinished{Err: errors.New("boom")})

	if len(actual) != 1 {
		t.Fatalf("expected one report-error command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reportErrorCmd); !ok {
		t.Fatalf("expected a reportErrorCmd, actual %T", actual[0])
	}
	if actualMessage := subject.ghCommandLoadingMessage; actualMessage != "" {
		t.Fatalf("expected gh command loading message %q, actual %q", "", actualMessage)
	}
}

func TestUpdate_GivenMsgApprovePullRequestRequested_WhenApplying_ThenItClearsPopupErrorStartsGHLoadingAndQueuesAnAsyncRequestCmd(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.actionsPopupWidget.errorMessage = "stale"

	actual := Update(subject, MsgApprovePullRequestRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	if actualMessage := subject.actionsPopupWidget.errorMessage; actualMessage != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actualMessage)
	}
	if actualMessage := subject.ghCommandLoadingMessage; actualMessage != formatRunningCommandStatus(approvePullRequestCommand("acme/widgets", 42)) {
		t.Fatalf("expected gh command loading message %q, actual %q", formatRunningCommandStatus(approvePullRequestCommand("acme/widgets", 42)), actualMessage)
	}
	given_actionsPopupAsyncCommand(t, actual)
}

func TestUpdate_GivenMsgPullRequestCommentDeleteRequested_WhenApplying_ThenItClearsPopupErrorAndQueuesAnAsyncRequestCmdWithoutGHLoading(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.actionsPopupWidget.errorMessage = "stale"
	subject.ghCommandLoadingMessage = "Running `gh pr ready`."

	actual := Update(subject, MsgPullRequestCommentDeleteRequested{Target: pullRequestCommentEditActionTarget{commentID: "comment-123"}})

	if len(actual) != 1 {
		t.Fatalf("expected one queued command, actual %d", len(actual))
	}
	if actualMessage := subject.actionsPopupWidget.errorMessage; actualMessage != "" {
		t.Fatalf("expected popup error message %q, actual %q", "", actualMessage)
	}
	if actualMessage := subject.ghCommandLoadingMessage; actualMessage != "Running `gh pr ready`." {
		t.Fatalf("expected gh command loading message %q, actual %q", "Running `gh pr ready`.", actualMessage)
	}
	given_actionsPopupAsyncCommand(t, actual)
}

func TestUpdate_GivenMsgThemePresetSaved_WhenApplying_ThenItReturnsAConfigureGUICommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.model.OpenActionsPopup(1)
	subject.actionsPopupWidget.themePicker = &themePickerState{}

	actual := Update(subject, MsgThemePresetSaved{NormalizedName: "night", Label: "Night"})

	if len(actual) != 1 {
		t.Fatalf("expected one configure-gui command, actual %d", len(actual))
	}
	if _, ok := actual[0].(configureGUICmd); !ok {
		t.Fatalf("expected a configureGUICmd, actual %T", actual[0])
	}
	if actualMessage := subject.feedbackMessage; actualMessage != "Theme changed to Night" {
		t.Fatalf("expected feedback %q, actual %q", "Theme changed to Night", actualMessage)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected the actions popup to close after a successful theme save")
	}
}

func TestUpdate_GivenMsgActionsPopupAsyncGHCommandFinishedWithThemeSuccess_WhenApplying_ThenItReturnsAConfigureGUICommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgActionsPopupAsyncGHCommandFinished{Success: actionsPopupAsyncThemeAppliedSuccess{NormalizedName: "night", Label: "Night"}})

	if len(actual) != 1 {
		t.Fatalf("expected one configure-gui command, actual %d", len(actual))
	}
	if _, ok := actual[0].(configureGUICmd); !ok {
		t.Fatalf("expected a configureGUICmd, actual %T", actual[0])
	}
	if actualMessage := subject.feedbackMessage; actualMessage != "Theme changed to Night" {
		t.Fatalf("expected feedback %q, actual %q", "Theme changed to Night", actualMessage)
	}
}

func TestUpdate_GivenMsgRefreshPullRequestListRequested_WhenApplying_ThenItReturnsManualRefreshRegistrationAndReloadCommands(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())

	actual := Update(subject, MsgRefreshPullRequestListRequested{})

	if len(actual) != 2 {
		t.Fatalf("expected two commands, actual %d", len(actual))
	}
	if _, ok := actual[0].(beginManualPullRequestListRefreshCmd); !ok {
		t.Fatalf("expected a beginManualPullRequestListRefreshCmd, actual %T", actual[0])
	}
	if _, ok := actual[1].(reloadPullRequestsTabCmd); !ok {
		t.Fatalf("expected a reloadPullRequestsTabCmd, actual %T", actual[1])
	}
}

func TestUpdate_GivenMsgRefreshNotificationsRequested_WhenApplying_ThenItReturnsATypedNotificationRefreshCommand(t *testing.T) {
	subject := NewProgramWithModel(NewModel(DefaultSeedData()))

	actual := Update(subject, MsgRefreshNotificationsRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one refresh-notifications command, actual %d", len(actual))
	}
	if _, ok := actual[0].(refreshNotificationsCmd); !ok {
		t.Fatalf("expected a refreshNotificationsCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgPullRequestsLoadedAfterManualRefreshFailure_WhenApplying_ThenItReturnsATypedReportErrorCommand(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.markManualPullRequestListRefresh(subject.model.ActivePullRequestTab())
	subject.beginManualRefresh(pullRequestListRefreshSuccessMessage, 1)

	actual := Update(subject, MsgPullRequestsLoaded{Tab: subject.model.ActivePullRequestTab(), Err: errors.New("boom")})

	if len(actual) != 1 {
		t.Fatalf("expected one report-error command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reportErrorCmd); !ok {
		t.Fatalf("expected a reportErrorCmd, actual %T", actual[0])
	}
}
