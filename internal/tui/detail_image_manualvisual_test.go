//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	detailImageManualVisualReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	detailImageManualVisualDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DetailPrivateMarkdownImageFallsBackToResolvedURL(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(detailImageManualVisualReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(detailImageManualVisualDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual detail-image check needs tmux wait-for tokens")
	}

	model := given_pullRequestCommentModel()
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		renderedMarkdownHTML: map[string]string{
			`acme/widgets|![Architecture](./docs/diagram.png)`: `<p><img src="https://raw.githubusercontent.com/acme/widgets/main/docs/diagram.png" alt="Architecture"></p>`,
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42, Body: "![Architecture](./docs/diagram.png)", State: "OPEN"})}
	subject.detailImageStore = nil
	subject.markdownRenderer = glamourMarkdownRenderer{imageStore: &fakeDetailImageStore{}, imageProtocol: kittyImageProtocol{support: fakeImageProtocolSupport{supported: false}}, terminalCellSize: fixedTerminalCellSize{width: 10, height: 10}}

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

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := runDetailImageManualVisualSequence(t, gui, readyToken, stopPolling); actualErr != nil {
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

func runDetailImageManualVisualSequence(t *testing.T, gui *gocui.Gui, readyToken string, stop <-chan struct{}) error {
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
				buffer := detailView.Buffer()
				if !strings.Contains(buffer, "https://raw.githubusercontent.com/acme/widgets/main/docs/diagram.png") {
					ready <- false
					errCh <- nil
					return nil
				}
				if strings.Contains(buffer, "./docs/diagram.png") {
					ready <- false
					errCh <- errors.New("expected the detail buffer to stop showing the unresolved relative image URL")
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
