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
	detailCharacterMotionSpaceReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	detailCharacterMotionSpaceDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DetailCharacterMotion_GivenALiveSpacebarTarget_WhenPressingVFSpaceY_ThenItYanksThroughTheSpace(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(detailCharacterMotionSpaceReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(detailCharacterMotionSpaceDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual detail character motion space check needs tmux wait-for tokens")
	}

	model := NewModel(SeedData{Users: []Item{{Title: "TTY visual", Detail: "foo bar baz"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	clipboardWriter := &manualVisualClipboardWriter{writes: make(chan string, 1)}
	subject.clipboardWriter = clipboardWriter
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	defer gui.Close()

	then_noError(t, subject.start(gui))
	then_noError(t, subject.focusDetailView(gui, nil))
	given_detailCursorOnSegment(t, gui, subject, "foo")
	then_currentViewNameIs(t, gui, viewDetailName)

	doneCh := make(chan error, 1)
	go func() {
		actual := <-clipboardWriter.writes
		if actual != "foo " {
			doneCh <- fmt.Errorf("expected copied text %q, actual %q", "foo ", actual)
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(doneToken); actualErr != nil {
			doneCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		doneCh <- nil
		gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
	}()

	then_noError(t, signalTmuxWaitToken(readyToken))

	actualErr = gui.MainLoop()
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	select {
	case actualErr := <-doneCh:
		then_noError(t, actualErr)
	case <-time.After(2 * time.Second):
		t.Fatal("expected the live key sequence to write the clipboard")
	}
}

type manualVisualClipboardWriter struct {
	writes chan string
}

func (writer *manualVisualClipboardWriter) WriteText(text string) error {
	writer.writes <- text
	return nil
}
