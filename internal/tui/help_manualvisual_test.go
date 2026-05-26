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
	helpManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	helpManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_HelpFullPageScroll(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(helpManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(helpManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual help check needs tmux wait-for tokens")
	}

	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
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
	if actualErr = subject.toggleHelp(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runHelpManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runHelpManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
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
				helpView, actualErr := gui.View(viewHelpName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}

				if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewHelpName, gocui.KeyPgdn)(gui, helpView); actualErr != nil {
					ready <- false
					errCh <- actualErr
					return actualErr
				}

				_, originY := helpView.Origin()
				currentView := gui.CurrentView()
				if originY <= 0 || currentView == nil || currentView.Name() != viewHelpName {
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
