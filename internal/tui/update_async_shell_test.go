package tui

import (
	"errors"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
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

func TestUpdate_GivenMsgRefreshPullRequestListRequested_WhenApplying_ThenItRegistersManualRefreshStateAndReturnsAReloadCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})

	actual := Update(subject, MsgRefreshPullRequestListRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reloadPullRequestsTabCmd); !ok {
		t.Fatalf("expected a reloadPullRequestsTabCmd, actual %T", actual[0])
	}
	if !subject.manualRefreshState.pullRequestListPending[subject.model.ActivePullRequestTab()] {
		t.Fatalf("expected the active pull request tab %v to be registered for manual refresh", subject.model.ActivePullRequestTab())
	}
	if subject.manualRefreshState.feedback == nil {
		t.Fatal("expected manual refresh feedback state to be initialized")
	}
	if actualPendingOperations := subject.manualRefreshState.feedback.pendingOperations; actualPendingOperations != 1 {
		t.Fatalf("expected one pending manual refresh operation, actual %d", actualPendingOperations)
	}
	if actualMessage := subject.manualRefreshState.feedback.successMessage; actualMessage != pullRequestListRefreshSuccessMessage {
		t.Fatalf("expected manual refresh success message %q, actual %q", pullRequestListRefreshSuccessMessage, actualMessage)
	}
}

func TestUpdate_GivenMsgRefreshPullRequestRequestedInBrowserMode_WhenApplying_ThenItRegistersDetailAndListManualRefreshStateBeforeReturningAReloadCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	summary := githubcli.ToDomainPullRequestSummary(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}})

	actual := Update(subject, MsgRefreshPullRequestRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, Summary: summary})

	if len(actual) != 1 {
		t.Fatalf("expected one command, actual %d", len(actual))
	}
	if _, ok := actual[0].(reloadPullRequestsTabCmd); !ok {
		t.Fatalf("expected a reloadPullRequestsTabCmd, actual %T", actual[0])
	}
	if !subject.manualRefreshState.pullRequestDetailPending["acme/widgets#42"] {
		t.Fatal("expected the visible pull request detail to be registered for manual refresh")
	}
	if !subject.manualRefreshState.pullRequestListPending[subject.model.ActivePullRequestTab()] {
		t.Fatalf("expected the active pull request tab %v to be registered for manual refresh", subject.model.ActivePullRequestTab())
	}
	if subject.manualRefreshState.feedback == nil {
		t.Fatal("expected manual refresh feedback state to be initialized")
	}
	if actualPendingOperations := subject.manualRefreshState.feedback.pendingOperations; actualPendingOperations != 2 {
		t.Fatalf("expected two pending manual refresh operations, actual %d", actualPendingOperations)
	}
}

func TestUpdate_GivenMsgRefreshPullRequestRequestedInReviewMode_WhenApplying_ThenItRegistersDetailAndDiffManualRefreshStateWithoutQueuingABrowserReloadCommand(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", BaseRefName: "main", HeadRefName: "feature/review", State: "OPEN"},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))
	summary := subject.navigationState.reviewSession.summary

	actual := Update(subject, MsgRefreshPullRequestRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}, Summary: summary})

	if len(actual) != 0 {
		t.Fatalf("expected no immediate browser reload commands in review mode, actual %d", len(actual))
	}
	if !subject.manualRefreshState.pullRequestDetailPending["acme/widgets#42"] {
		t.Fatal("expected the visible pull request detail to be registered for manual refresh")
	}
	if !subject.manualRefreshState.pullRequestDiffPending["acme/widgets#42"] {
		t.Fatal("expected the visible pull request diff to be registered for manual refresh")
	}
	if subject.manualRefreshState.pullRequestListPending[subject.model.ActivePullRequestTab()] {
		t.Fatalf("expected the active pull request tab %v to stay unregistered in review mode", subject.model.ActivePullRequestTab())
	}
	if subject.manualRefreshState.feedback == nil {
		t.Fatal("expected manual refresh feedback state to be initialized")
	}
	if actualPendingOperations := subject.manualRefreshState.feedback.pendingOperations; actualPendingOperations != 2 {
		t.Fatalf("expected two pending manual refresh operations, actual %d", actualPendingOperations)
	}
}

func TestUpdate_GivenMsgRefreshNotificationsRequested_WhenApplying_ThenItRegistersManualRefreshStateAndReturnsATypedNotificationRefreshCommand(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})

	actual := Update(subject, MsgRefreshNotificationsRequested{})

	if len(actual) != 1 {
		t.Fatalf("expected one refresh-notifications command, actual %d", len(actual))
	}
	if _, ok := actual[0].(refreshNotificationsCmd); !ok {
		t.Fatalf("expected a refreshNotificationsCmd, actual %T", actual[0])
	}
	if !subject.manualRefreshState.notificationPending {
		t.Fatal("expected notifications to be registered for manual refresh")
	}
	if subject.manualRefreshState.feedback == nil {
		t.Fatal("expected manual refresh feedback state to be initialized")
	}
	if actualPendingOperations := subject.manualRefreshState.feedback.pendingOperations; actualPendingOperations != 1 {
		t.Fatalf("expected one pending manual refresh operation, actual %d", actualPendingOperations)
	}
	if actualMessage := subject.manualRefreshState.feedback.successMessage; actualMessage != notificationsRefreshSuccessMessage {
		t.Fatalf("expected manual refresh success message %q, actual %q", notificationsRefreshSuccessMessage, actualMessage)
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
