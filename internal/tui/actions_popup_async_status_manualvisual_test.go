//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
)

const (
	actionsPopupAsyncStatusReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	actionsPopupAsyncStatusDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ActionsPopupAsyncStatusLinePreflight(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(actionsPopupAsyncStatusReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(actionsPopupAsyncStatusDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual popup async status check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_pullRequestCommentModel())
	subject.asyncRunner = &capturingAsyncRunner{}
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
	if actualErr = subject.layout(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.openActionsPopup(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runActionsPopupAsyncStatusManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runActionsPopupAsyncStatusManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	expectedStatus := formatRunningCommandStatus(openPullRequestInBrowserCommand("acme/widgets", 42))
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	submittedAction := false

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			ready := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				if !submittedAction {
					if actualErr := subject.dispatch(gui, MsgOpenPullRequestInBrowserRequested{Target: pullRequestActionTarget{repository: "acme/widgets", number: 42}}); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					submittedAction = true
				}

				if !subject.model.ActionsPopupVisible() {
					ready <- false
					errCh <- nil
					return nil
				}
				if _, actualErr := gui.View(viewActionsPopupName); actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}
				statusView, actualErr := gui.View(viewStatusLineName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}
				statusBuffer := statusView.Buffer()
				if !strings.Contains(statusBuffer, string(loadingSpinnerFrames[0])) || !strings.Contains(statusBuffer, expectedStatus) {
					ready <- false
					errCh <- nil
					return nil
				}

				ready <- true
				errCh <- nil
				return nil
			})
			if actualErr := <-errCh; actualErr != nil {
				return actualErr
			}
			if !<-ready {
				continue
			}
			return signalTmuxWaitToken(readyToken)
		}
	}
}
