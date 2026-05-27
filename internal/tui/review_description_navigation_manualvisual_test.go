//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	reviewDescriptionNavigationReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	reviewDescriptionNavigationDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ReviewDescriptionNavigationMovesTheDetailCursor(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(reviewDescriptionNavigationReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(reviewDescriptionNavigationDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual review-description navigation check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        strings.Repeat("Body 42\n", 20),
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
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
	if actualErr = given_startingReviewMode(t, gui, subject); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.focusUserView(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.focusDetailView(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runReviewDescriptionNavigationManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runReviewDescriptionNavigationManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			ready := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				detailView, actualErr := gui.View(viewDetailName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}

				startLine := subject.detailState.viewState.cursor.line
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'j')(gui, detailView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if subject.detailState.viewState.cursor.line <= startLine {
					ready <- false
					errCh <- errors.New("expected detail cursor line to advance in review description mode")
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
