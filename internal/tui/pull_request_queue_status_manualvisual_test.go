//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	pullRequestQueueStatusManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	pullRequestQueueStatusManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_PullRequestQueuedStatusPill(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(pullRequestQueueStatusManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(pullRequestQueueStatusManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual queued-status check needs tmux wait-for tokens")
	}

	summary := given_pullRequestLifecycleSummary("OPEN", false)
	summary.IsMergeQueueEnabled = true
	summary.IsInMergeQueue = true
	summary.MergeQueueEntry = &githubcli.PullRequestMergeQueueEntry{State: "QUEUED"}
	model := given_pullRequestLifecycleModel(summary)
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(given_pullRequestMergeQueueDetail("OPEN", false, true))}

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

		if actualErr := signalTmuxWaitTokenWhenDetailContains(gui, detailStatusIcon+" QUEUED", readyToken, stopPolling); actualErr != nil {
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

func signalTmuxWaitTokenWhenDetailContains(gui *gocui.Gui, expected string, readyToken string, stop <-chan struct{}) error {
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
				view, actualErr := gui.View(viewDetailName)
				viewReady <- actualErr == nil && strings.Contains(view.Buffer(), expected)
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
