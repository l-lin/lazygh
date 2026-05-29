package tui

import (
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func TestUpdate_GivenMsgPullRequestBuildRunLoadRequested_WhenApplying_ThenItClearsFeedbackStartsLoadingAndQueuesABuildRunCommand(t *testing.T) {
	subject := given_programWithTestGitHubDeps(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.feedbackMessage = "stale"
	subject.model.OpenActionsPopup(1)
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Old", body: "old"})
	summary, ok := subject.currentPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	target := pullRequestBuildRunTarget{
		summary: summary,
		check:   githubdomain.PullRequestStatusCheck{Name: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"},
		popupContent: pullRequestBuildRunPopupContent{
			checkTitle: "CI",
			repository: "acme/widgets",
		},
	}

	actual := Update(subject, MsgPullRequestBuildRunLoadRequested{Target: target})

	if len(actual) != 1 {
		t.Fatalf("expected one build-run load command, actual %d", len(actual))
	}
	command, ok := actual[0].(pullRequestBuildRunLoadCmd)
	if !ok {
		t.Fatalf("expected a pullRequestBuildRunLoadCmd, actual %T", actual[0])
	}
	if actual := command.Repository; actual != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", actual)
	}
	if subject.pullRequestBuildRunPopup != nil {
		t.Fatal("expected load preflight to clear any stale build popup")
	}
	if subject.pullRequestBuildRunLoad == nil {
		t.Fatal("expected load preflight to start build-run loading")
	}
	if actual := subject.pullRequestBuildRunLoad.command; actual != formatPullRequestBuildRunCommand("acme/widgets", target.check) {
		t.Fatalf("expected build-run loading command %q, actual %q", formatPullRequestBuildRunCommand("acme/widgets", target.check), actual)
	}
	if actual := subject.feedbackMessage; actual != "" {
		t.Fatalf("expected feedback message %q, actual %q", "", actual)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected build-run load preflight to close the actions popup")
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunJobLogLoadRequested_WhenApplying_ThenItKeepsTheExistingPopupStartsLoadingAndQueuesAJobLogCommand(t *testing.T) {
	subject := given_programWithTestGitHubDeps(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.feedbackMessage = "stale"
	subject.model.OpenActionsPopup(1)
	previousPopup := newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Current", body: "visible"})
	subject.pullRequestBuildRunPopup = previousPopup
	summary, ok := subject.currentPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	check := githubdomain.PullRequestStatusCheck{Name: "CI", Link: "https://github.com/acme/widgets/actions/runs/42"}

	actual := Update(subject, MsgPullRequestBuildRunJobLogLoadRequested{Summary: summary, Check: check})

	if len(actual) != 1 {
		t.Fatalf("expected one build-run job-log command, actual %d", len(actual))
	}
	command, ok := actual[0].(pullRequestBuildRunJobLogLoadCmd)
	if !ok {
		t.Fatalf("expected a pullRequestBuildRunJobLogLoadCmd, actual %T", actual[0])
	}
	if actual := command.Repository; actual != "acme/widgets" {
		t.Fatalf("expected repository %q, actual %q", "acme/widgets", actual)
	}
	if subject.pullRequestBuildRunPopup != previousPopup {
		t.Fatal("expected job-log load preflight to preserve the visible popup")
	}
	if subject.pullRequestBuildRunLoad == nil {
		t.Fatal("expected job-log load preflight to start loading")
	}
	if actual := subject.pullRequestBuildRunLoad.command; actual != formatPullRequestBuildRunJobsCommand("acme/widgets", check) {
		t.Fatalf("expected build-run job-log loading command %q, actual %q", formatPullRequestBuildRunJobsCommand("acme/widgets", check), actual)
	}
	if actual := subject.feedbackMessage; actual != "" {
		t.Fatalf("expected feedback message %q, actual %q", "", actual)
	}
	if subject.model.ActionsPopupVisible() {
		t.Fatal("expected job-log load preflight to close the actions popup")
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunPopupClosedWithVisualSelection_WhenApplying_ThenItKeepsThePopupOpenAndExitsVisualMode(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Current", body: "current"})
	subject.pullRequestBuildRunPopup.viewState.enterVisualMode()

	actual := Update(subject, MsgPullRequestBuildRunPopupClosed{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the popup to stay visible while visual mode exits")
	}
	if actual := subject.pullRequestBuildRunPopup.viewState.mode; actual != detailNormalMode {
		t.Fatalf("expected popup mode %v after close preflight, actual %v", detailNormalMode, actual)
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunPopupClosedWithPendingYank_WhenApplying_ThenItKeepsThePopupOpenAndClearsThePendingPrefix(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Current", body: "current"})
	subject.pullRequestBuildRunPopup.viewState.armPendingYank()

	actual := Update(subject, MsgPullRequestBuildRunPopupClosed{})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected the popup to stay visible while the pending prefix clears")
	}
	if subject.pullRequestBuildRunPopup.viewState.hasPendingYank() {
		t.Fatal("expected close preflight to clear the pending popup prefix")
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunPopupClosed_WhenPreviousPopupExists_ThenItRestoresThePreviousPopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	previousPopup := newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Previous", body: "old"})
	subject.pullRequestBuildRunPopup = newPullRequestBuildRunPopupState(pullRequestBuildRunPopupContent{checkTitle: "Current", body: "current", previousPopup: previousPopup})

	Update(subject, MsgPullRequestBuildRunPopupClosed{})

	if subject.pullRequestBuildRunPopup != previousPopup {
		t.Fatalf("expected the previous popup to be restored, actual %+v", subject.pullRequestBuildRunPopup)
	}
}

func TestUpdate_GivenMsgPullRequestBuildRunJobLogLoaded_WhenSuccessful_ThenItClearsLoadingAndOpensThePopup(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.pullRequestBuildRunLoad = &pullRequestBuildRunLoadState{command: "gh run view --json jobs"}
	job := githubdomain.PullRequestBuildRunJob{Name: "lint", URL: "https://github.com/acme/widgets/actions/runs/42/job/99"}

	Update(subject, MsgPullRequestBuildRunJobLogLoaded{Repository: "acme/widgets", Job: job, RawLogOutput: "job body"})

	if subject.pullRequestBuildRunLoad != nil {
		t.Fatalf("expected job-log loading state to clear after the async result")
	}
	if subject.pullRequestBuildRunPopup == nil {
		t.Fatal("expected a successful job-log load to open the popup")
	}
	if actual := subject.pullRequestBuildRunPopup.title; actual != pullRequestBuildRunLogsPopupTitle(job.Name) {
		t.Fatalf("expected popup title %q, actual %q", pullRequestBuildRunLogsPopupTitle(job.Name), actual)
	}
}
