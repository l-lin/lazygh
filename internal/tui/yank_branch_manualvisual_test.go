//go:build manualvisual

package tui

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestManualVisual_PullRequestYank_GivenLoadedPR_WhenPressingAltYAndAltShiftY_ThenItCopiesBranchAndURL(t *testing.T) {
	readyToken := os.Getenv("LAZYGH_TMUX_READY_TOKEN")
	doneToken := os.Getenv("LAZYGH_TMUX_DONE_TOKEN")
	stopToken := os.Getenv("LAZYGH_TMUX_STOP_TOKEN")
	if readyToken == "" || doneToken == "" || stopToken == "" {
		t.Skip("live yank check needs tmux synchronization tokens")
	}

	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Branch yank TTY check", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/42"}),
	})
	subject := NewProgramWithModel(model)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "Branch yank TTY check", Number: 42, BaseRefName: "main", HeadRefName: "feature/yank-branch"})}
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
			for _, expected := range []string{"feature/yank-branch", "https://github.com/acme/widgets/pull/42"} {
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
