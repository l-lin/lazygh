//go:build manualvisual

package tui

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jesseduffield/gocui"
)

func TestManualVisual_ReviewPathYank_GivenViewTwo_WhenPressingY_ThenItCopiesTheSelectedPath(t *testing.T) {
	readyToken := os.Getenv("LAZYGH_TMUX_READY_TOKEN")
	doneToken := os.Getenv("LAZYGH_TMUX_DONE_TOKEN")
	stopToken := os.Getenv("LAZYGH_TMUX_STOP_TOKEN")
	if readyToken == "" || doneToken == "" || stopToken == "" {
		t.Skip("live path copy check needs tmux synchronization tokens")
	}
	mode := reviewSessionModeDiff
	expected := []string{"internal/tui/", "internal/tui/new name.go"}
	if os.Getenv("LAZYGH_PATH_YANK_STORY") == "1" {
		mode = reviewSessionModeStory
		expected = []string{"internal/tui/new name.go", "internal/tui/deleted.go"}
	}
	subject := given_pathYankProgram(mode)
	if mode == reviewSessionModeStory {
		subject.navigationState.reviewSession.selectedFileTreeRow = 1
	}
	clipboardWriter := &manualVisualClipboardWriter{writes: make(chan string, 2)}
	subject.clipboardWriter = clipboardWriter
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	then_noError(t, actualErr)
	defer gui.Close()
	then_noError(t, subject.start(gui))

	stopPolling := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		actualErr := signalTmuxWaitTokenWhenViewAppears(gui, viewPullRequestsName, readyToken, stopPolling)
		if actualErr == nil {
			for _, expected := range expected {
				actual := <-clipboardWriter.writes
				if actual != expected {
					actualErr = fmt.Errorf("expected clipboard text %q, actual %q", expected, actual)
					break
				}
			}
		}
		if actualErr == nil {
			actualErr = signalTmuxWaitToken(doneToken)
		}
		if actualErr == nil {
			actualErr = waitForTmuxToken(stopToken)
		}
		result <- actualErr
		gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
	}()

	actualErr = gui.MainLoop()
	close(stopPolling)
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatalf("expected a clean main loop exit, actual %v", actualErr)
	}
	then_noError(t, <-result)
}
