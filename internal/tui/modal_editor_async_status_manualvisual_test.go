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
	modalEditorAsyncStatusReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	modalEditorAsyncStatusDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ModalEditorSubmitShowsGhLoadingInStatusLine(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(modalEditorAsyncStatusReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(modalEditorAsyncStatusDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual modal-editor async status check needs tmux wait-for tokens")
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
		if actualErr := runModalEditorAsyncStatusManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runModalEditorAsyncStatusManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	expectedStatus := formatRunningCommandStatus(formatStatusLineCommand("gh", "pr", "edit", "42", "-R", "acme/widgets", "--title", "First PR!"))
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	openedModal := false
	submittedModal := false

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			ready := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				if !openedModal {
					searchView, actualErr := gui.View(viewActionsPopupSearchName)
					if actualErr != nil {
						ready <- false
						errCh <- nil
						return nil
					}
					for _, character := range pullRequestTitleEditorTitle {
						if actual := subject.editActionsPopupSearch(searchView, 0, character, gocui.ModNone); !actual {
							ready <- false
							errCh <- errors.New("expected popup search typing to be handled")
							return nil
						}
					}
					if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupSearchName, gocui.KeyEnter)(gui, searchView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					openedModal = true
					ready <- false
					errCh <- nil
					return nil
				}

				modalView, actualErr := gui.View(viewModalEditorName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}
				if !submittedModal {
					if actual := subject.editModalEditor(modalView, 0, '!', gocui.ModNone); !actual {
						ready <- false
						errCh <- errors.New("expected modal editor typing to be handled")
						return nil
					}
					if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyEnter)(gui, modalView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					submittedModal = true
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
				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewModalEditorName {
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
