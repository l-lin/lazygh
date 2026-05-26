//go:build manualvisual

package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
)

const (
	detailVisualYankManualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	detailVisualYankManualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DetailBottomMotionAndLineVisualYank(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(detailVisualYankManualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(detailVisualYankManualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual detail motion check needs tmux wait-for tokens")
	}

	lines := make([]string, 0, 20)
	for lineNumber := 1; lineNumber <= 20; lineNumber++ {
		lines = append(lines, fmt.Sprintf("Line %02d", lineNumber))
	}
	model := NewModel(SeedData{Users: []Item{{Title: "TTY detail", Detail: strings.Join(lines, "\n")}}})
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
		if actualErr := runDetailVisualYankManualSequence(t, gui, subject, readyToken, stopPolling); actualErr != nil {
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

func runDetailVisualYankManualSequence(t *testing.T, gui *gocui.Gui, subject *Program, readyToken string, stop <-chan struct{}) error {
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

				for _, key := range []rune{'G', 'V', 'y'} {
					if actualErr := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, key)(gui, detailView); actualErr != nil {
						ready <- false
						errCh <- actualErr
						return actualErr
					}
				}

				currentView := gui.CurrentView()
				if currentView == nil || currentView.Name() != viewDetailName {
					ready <- false
					errCh <- nil
					return nil
				}
				statusView, actualErr := gui.View(viewStatusLineName)
				if actualErr != nil || !strings.Contains(statusView.Buffer(), detailYankSuccessMessage) {
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
