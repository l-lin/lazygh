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
	overlayManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	overlayManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_HelpOverlayOpens(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(overlayManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(overlayManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual overlay check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_model())
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
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			helpView, viewErr := gui.View(viewHelpName)
			if viewErr != nil {
				return false
			}
			currentView := gui.CurrentView()
			return currentView != nil && currentView.Name() == viewHelpName && strings.Contains(helpView.Title, "Keybindings")
		}); actualErr != nil {
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

func TestManualVisual_ModalEditorOpensAndAcceptsInput(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(overlayManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(overlayManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual overlay check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_model())
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
	if actualErr = subject.openLineModalEditor(gui, "Prompt", "draft"); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runOverlayModalEditorManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runOverlayModalEditorManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
	t.Helper()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	typed := false

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			ready := make(chan bool, 1)
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				modalView, actualErr := gui.View(viewModalEditorName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}

				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewModalEditorName || !strings.Contains(modalView.Title, "Prompt") {
					ready <- false
					errCh <- nil
					return nil
				}
				if !typed {
					if strings.TrimSpace(modalView.Buffer()) != "draft" {
						ready <- false
						errCh <- nil
						return nil
					}
					if actual := subject.editModalEditor(modalView, 0, '!', gocui.ModNone); !actual {
						ready <- false
						errCh <- errors.New("expected modal editor typing to be handled")
						return nil
					}
					typed = true
					ready <- false
					errCh <- nil
					return nil
				}
				if strings.TrimSpace(modalView.Buffer()) != "draft!" {
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
