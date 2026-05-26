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
	actionsPopupManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	actionsPopupManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ActionsPopupViewportPlacementTop(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(actionsPopupManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(actionsPopupManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual actions-popup viewport check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_pullRequestCommentModel())
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
		if actualErr := runActionsPopupViewportManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runActionsPopupViewportManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	const expectedActionTitle = "Edit PR description"

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
				if actualErr := subject.focusActionsPopupList(gui, nil); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				popupView, actualErr := gui.View(viewActionsPopupName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}

				for attempt := 0; attempt < 20; attempt++ {
					action, ok := subject.selectedActionsPopupAction()
					if ok && action.title == expectedActionTitle {
						break
					}
					if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'j')(gui, popupView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
					popupView, actualErr = gui.View(viewActionsPopupName)
					if actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
				}

				action, ok := subject.selectedActionsPopupAction()
				if !ok || action.title != expectedActionTitle {
					ready <- false
					errCh <- errors.New("expected to focus the target actions-popup row")
					return nil
				}
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 'z')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}
				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewActionsPopupName, 't')(gui, popupView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}

				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewActionsPopupName {
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
