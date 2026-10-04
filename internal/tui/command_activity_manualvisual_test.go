//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

type given_activityVisualRunner struct{}

func (given_activityVisualRunner) Run(_ string, args ...string) (githubcli.CommandResult, error) {
	if args[len(args)-1] == "failure" {
		return githubcli.CommandResult{}, errors.New("fixture failure")
	}
	return githubcli.CommandResult{}, nil
}
func (runner given_activityVisualRunner) RunWithInput(name string, _ []byte, args ...string) (githubcli.CommandResult, error) {
	return runner.Run(name, args...)
}

func TestManualVisual_CommandActivityHistory(t *testing.T) {
	readyToken := os.Getenv("LAZYGH_TMUX_READY_TOKEN")
	doneToken := os.Getenv("LAZYGH_TMUX_DONE_TOKEN")
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual activity check needs tmux wait-for tokens")
	}
	// GIVEN
	subject := NewProgramWithModel(given_model())
	executor := githubcli.NewExecutor(githubcli.NewActivityObservingRunner(given_activityVisualRunner{}, subject.ReportCommandActivity), nil, nil)
	_, _ = executor.Execute(githubcli.Command{Args: []string{"api", "success"}, DisplayArgs: []string{"api", "user"}})
	_, _ = executor.Execute(githubcli.Command{Args: []string{"api", "failure"}, DisplayArgs: []string{"api", "graphql"}})
	currentTime := time.Date(2026, 10, 4, 12, 0, 0, 0, time.Local)
	var queuedRun func()
	runObservedCommandAsync(func(run func()) { queuedRun = run }, subject.dispatchAsyncMessage, func() time.Time { return currentTime }, "refresh diff acme/widgets#42", "scheduled batch=7", func() error {
		currentTime = currentTime.Add(2 * time.Second)
		return errors.New("fixture enrichment failure")
	})
	currentTime = currentTime.Add(350 * time.Millisecond)
	queuedRun()
	subject.ReportCommandActivity(githubdomain.CommandActivity{ID: "refresh-visual", Command: "refresh detail acme/widgets#43", StartedAt: currentTime, Outcome: githubdomain.CommandRunning, Owner: "interactive", QueueWait: 20 * time.Millisecond})
	// WHEN
	Update(subject, MsgOpenCommandMode{})
	for _, character := range "history" {
		Update(subject, MsgCommandInputRequested{Intent: newLineEditorInsertRuneIntent(character)})
	}
	Update(subject, MsgSubmitCommand{})
	gui, actualErr := gocui.NewGui(gocui.NewGuiOpts{OutputMode: gocui.OutputTrue})
	if actualErr != nil {
		t.Fatal(actualErr)
	}
	defer gui.Close()
	subject.configureGUI(gui)
	gui.SetManagerFunc(subject.layout)
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatal(actualErr)
	}
	if actualErr = subject.layout(gui); actualErr != nil {
		t.Fatal(actualErr)
	}
	actualResults := make(chan error, 1)
	go func() {
		gui.Update(func(gui *gocui.Gui) error {
			// THEN
			view, actualErr := gui.View(viewPullRequestBuildInfoName)
			if actualErr == nil {
				for _, expected := range []string{"gh api user [success", "gh api graphql [failure", "refresh diff acme/widgets#42 [failure · 2s · queue=350ms · scheduled batch=7]", "refresh detail acme/widgets#43 [running · queue=20ms · interactive]"} {
					if !strings.Contains(view.Buffer(), expected) {
						actualErr = errors.New("missing history text: " + expected)
						break
					}
				}
			}
			if actualErr == nil {
				actualErr = signalTmuxWaitToken(readyToken)
			}
			actualResults <- actualErr
			return nil
		})
		if actualErr := <-actualResults; actualErr != nil {
			actualResults <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		actualResults <- waitForTmuxToken(doneToken)
		gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
	}()
	actualErr = gui.MainLoop()
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatal(actualErr)
	}
	if actualErr = <-actualResults; actualErr != nil {
		t.Fatal(actualErr)
	}
}
