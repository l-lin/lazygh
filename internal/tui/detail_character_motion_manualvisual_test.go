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
	detailCharacterMotionReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	detailCharacterMotionDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DetailCharacterMotion(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(detailCharacterMotionReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(detailCharacterMotionDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual detail character motion check needs tmux wait-for tokens")
	}

	model := NewModel(SeedData{Users: []Item{{Title: "TTY visual", Detail: "banana"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.clipboardWriter = &fakeClipboardWriter{}
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
		if actualErr := runDetailCharacterMotionManualVisualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runDetailCharacterMotionManualVisualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
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
				view, actualErr := gui.View(viewDetailName)
				if actualErr != nil {
					ready <- false
					errCh <- nil
					return nil
				}

				bindings := subject.registeredKeybindingSpecs()
				for range 3 {
					if actualErr := given_handlerForBinding(t, bindings, viewDetailName, 'j')(gui, view); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
				}
				for _, key := range []rune{'v', 'f', 'a', ';', 'y'} {
					if actualErr := given_handlerForBinding(t, bindings, viewDetailName, key)(gui, view); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
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
