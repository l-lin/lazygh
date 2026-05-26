//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

const (
	viewportPlacementReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	viewportPlacementDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_SideViewportPlacementTop(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(viewportPlacementReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(viewportPlacementDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual side viewport placement check needs tmux wait-for tokens")
	}

	subject := NewProgramWithModel(NewModel(SeedData{Users: given_manyItems("user", 40)}))
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	defer gui.Close()

	subject.configureGUI(gui)
	gui.SetManagerFunc(subject.layout)
	if actualErr = subject.layout(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	expectedOrigin := 0
	stopPolling := make(chan struct{})
	pollingStopped := make(chan error, 1)
	go func() {
		defer close(pollingStopped)
		operationDone := make(chan error, 1)
		gui.Update(func(gui *gocui.Gui) error {
			userView, viewErr := gui.View(viewUserName)
			if viewErr != nil {
				operationDone <- viewErr
				return nil
			}
			targetIndex := userView.InnerHeight() + 3
			for range targetIndex {
				if moveErr := subject.moveSelectionDown(gui, userView); moveErr != nil {
					operationDone <- moveErr
					return nil
				}
				userView, viewErr = gui.View(viewUserName)
				if viewErr != nil {
					operationDone <- viewErr
					return nil
				}
			}
			expectedOrigin = expectedViewportTopOrigin(targetIndex, userView.InnerHeight(), len(subject.model.VisibleUsers()))
			operationDone <- subject.moveSideSelectionToViewportTop(gui, userView)
			return nil
		})
		if actualErr := <-operationDone; actualErr != nil {
			pollingStopped <- actualErr
			return
		}
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			view, viewErr := gui.View(viewUserName)
			if viewErr != nil {
				return false
			}
			_, originY := view.Origin()
			return originY == expectedOrigin
		}); actualErr != nil {
			pollingStopped <- actualErr
			return
		}
		if actualErr := waitForTmuxToken(doneToken); actualErr != nil {
			pollingStopped <- actualErr
			return
		}
		gui.Update(func(*gocui.Gui) error {
			return gocui.ErrQuit
		})
		pollingStopped <- nil
	}()

	actualErr = gui.MainLoop()
	close(stopPolling)
	if pollingErr := <-pollingStopped; pollingErr != nil {
		t.Fatalf("expected no error, actual %v", pollingErr)
	}
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
