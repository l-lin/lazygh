//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
)

const (
	panelViewContractReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	panelViewContractDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_PanelViewContracts(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(panelViewContractReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(panelViewContractDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual panel/view smoke check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(given_panelViewContractManualVisualModel())
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

func signalTmuxWaitTokenWhenViewAppears(gui *gocui.Gui, viewName string, readyToken string, stop <-chan struct{}) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	var signalOnce sync.Once
	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			viewReady := make(chan bool, 1)
			gui.Update(func(gui *gocui.Gui) error {
				_, actualErr := gui.View(viewName)
				viewReady <- actualErr == nil
				return nil
			})
			if !<-viewReady {
				continue
			}

			var signalErr error
			signalOnce.Do(func() {
				signalErr = signalTmuxWaitToken(readyToken)
			})
			return signalErr
		}
	}
}

func signalTmuxWaitToken(token string) error {
	return exec.Command("tmux", "wait-for", "-S", token).Run()
}

func waitForTmuxToken(token string) error {
	return exec.Command("tmux", "wait-for", token).Run()
}
