//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	shortRepoLabelsReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	shortRepoLabelsDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ShortRepoLabels(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(shortRepoLabelsReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(shortRepoLabelsDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual short-repo-label check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_shortRepoLabelsManualVisualModel())
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

func given_shortRepoLabelsManualVisualModel() *Model {
	model := NewModel(DefaultSeedData())
	model.FocusNotificationsView()
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
	return model
}
