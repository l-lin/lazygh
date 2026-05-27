//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"

	githubcli "github.com/l-lin/lazygh/internal/githubcli"
)

const (
	detailFoldManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	detailFoldManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_BrowserConversationFoldToggle(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(detailFoldManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(detailFoldManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual detail-fold check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "Comment body",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Comment body": "Rendered comment body"}}
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
		if actualErr := runDetailFoldManualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runDetailFoldManualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	detailOpened := false
	commentsTabSelected := false
	collapsed := false

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			ready := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				if !detailOpened {
					if actualErr := subject.openDetail(gui, nil); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					detailOpened = true
				}
				if !commentsTabSelected {
					if actualErr := subject.nextDetailTab(gui, nil); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					commentsTabSelected = true
				}

				detailView, actualErr := gui.View(viewDetailName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}
				if !strings.Contains(detailView.Buffer(), " Comment") {
					ready <- false
					errCh <- nil
					return nil
				}

				if !collapsed {
					prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
					toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
					if actualErr := prefixHandler(gui, detailView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					if actualErr := toggleHandler(gui, detailView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					collapsed = true
				}

				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewDetailName {
					ready <- false
					errCh <- nil
					return nil
				}
				if !strings.Contains(detailView.Buffer(), " Comment") || strings.Contains(detailView.Buffer(), "Rendered comment body") {
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
