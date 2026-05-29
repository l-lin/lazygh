//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

const (
	pastedPullRequestTabReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	pastedPullRequestTabDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_PastedPullRequestTab(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(pastedPullRequestTabReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(pastedPullRequestTabDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual pasted-tab check needs tmux wait-for tokens")
	}

	model := given_model()
	model.FocusPullRequestsView()
	subject := NewProgramWithModel(model)
	subject.connectedUserLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.updatePastedPullRequestTabState(func(state pastedPullRequestTabState) pastedPullRequestTabState {
		return state.withPullRequestAdded(githubdomain.PullRequest{
			Title:      "Widgets PR",
			Number:     13,
			Repository: githubdomain.Repository{NameWithOwner: "acme/widgets"},
			URL:        "https://github.com/acme/widgets/pull/13",
			Body:       "Body 13",
			State:      "OPEN",
		})
	})
	pastedTab, ok := subject.syncPastedPullRequestTab()
	if !ok {
		t.Fatal("expected the pasted tab to exist")
	}
	subject.model.SetActivePullRequestTab(pastedTab)

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

		if actualErr := signalTmuxWaitTokenWhenViewAppears(gui, viewPullRequestsName, readyToken, stopPolling); actualErr != nil {
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
