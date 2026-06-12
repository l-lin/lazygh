//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	shortRepoLabelsReadyTokenEnv      = "LAZYGH_TMUX_READY_TOKEN"
	shortRepoLabelsDoneTokenEnv       = "LAZYGH_TMUX_DONE_TOKEN"
	shortRepoLabelsRepositoryStyleEnv = "LAZYGH_REPOSITORY_STYLE"
	shortRepoLabelsModeEnv            = "LAZYGH_REPO_LABELS_MODE"
)

func TestManualVisual_ShortRepoLabels(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(shortRepoLabelsReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(shortRepoLabelsDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual short-repo-label check needs tmux wait-for tokens")
	}

	subject := given_shortRepoLabelsManualVisualProgram()
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	defer gui.Close()

	subject.configureGUI(gui)
	gui.SetManagerFunc(subject.layout)
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)

		if actualErr := signalTmuxWaitTokenWhenViewAppears(gui, viewNotificationsName, readyToken, stopPolling); actualErr != nil {
			return
		}
		if actualErr := waitForTmuxToken(doneToken); actualErr != nil {
			return
		}
		gui.Update(func(*gocui.Gui) error {
			return gocui.ErrQuit
		})
	}()

	actualErr = gui.MainLoop()
	close(stopPolling)
	<-pollingStopped
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}

func given_shortRepoLabelsManualVisualProgram() *Program {
	mode := strings.TrimSpace(os.Getenv(shortRepoLabelsModeEnv))
	model := given_shortRepoLabelsManualVisualModel()
	if mode == "notifications" {
		loader := &fakePullRequestDetailLoader{
			details: map[string]githubcli.PullRequestDetail{
				"acme/widgets#42": {
					Title:       "Add notifications",
					Number:      42,
					Body:        "Pull request body",
					State:       "OPEN",
					BaseRefName: "main",
					HeadRefName: "feature/notifications",
				},
			},
			issueDetails: map[string]githubcli.IssueDetail{
				"acme/opencode#3235": {
					Title:    "Support notifications in issue detail",
					Number:   3235,
					Body:     "Issue body",
					State:    "open",
					Comments: 7,
				},
			},
			releaseDetails: map[string]githubcli.ReleaseDetail{
				"acme/doctoboot#317927281": {
					Name:       "Notifications 3.5.0",
					TagName:    "v3.5.0",
					Body:       "Release notes",
					PreRelease: true,
				},
			},
		}
		subject := given_programWithTestGitHubDeps(model, loader)
		subject.connectedUserLoadStarted = true
		subject.myPullRequestsLoadStarted = true
		subject.requestedPullRequestsLoadStarted = true
		subject.notificationsLoadStarted = true
		subject.asyncRunner = inlineAsyncRunner{}
		subject.uiUpdater = immediateUIUpdater{}
		subject.ApplyDisplayConfig(appconfig.DisplayConfig{RepositoryStyle: strings.TrimSpace(os.Getenv(shortRepoLabelsRepositoryStyleEnv))})
		return subject
	}

	subject := NewProgramWithModel(model)
	subject.ApplyDisplayConfig(appconfig.DisplayConfig{RepositoryStyle: strings.TrimSpace(os.Getenv(shortRepoLabelsRepositoryStyleEnv))})
	return subject
}

func given_shortRepoLabelsManualVisualModel() *Model {
	model := NewModel(DefaultSeedData())
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{myPullRequestRow(githubcli.PullRequest{
		Title:      "Ship notifications",
		Number:     42,
		Repository: githubcli.Repository{Name: "widgets", NameWithOwner: "acme/widgets"},
		URL:        "https://github.com/acme/widgets/pull/42",
		Body:       "Pull request detail body",
		State:      "OPEN",
	})})
	model.SetNotificationRows([]NotificationRow{
		given_pullRequestNotificationRow(),
		given_issueNotificationRow(),
		given_releaseNotificationRow(),
	})
	if strings.TrimSpace(os.Getenv(shortRepoLabelsModeEnv)) == "notifications" {
		model.FocusNotificationsView()
	} else {
		model.FocusPullRequestsView()
	}
	return model
}
