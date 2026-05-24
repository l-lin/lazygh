//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	changesPopupReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	changesPopupDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_BrowserChangesInlineCommentResolutionActionsPopup(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(changesPopupReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(changesPopupDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual changes-popup smoke check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}

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
	if actualErr = subject.openDetail(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	subject.activeDetailTab = ChangesDetailTab
	if actualErr = subject.afterStateChange(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")
	if actualErr = subject.openActionsPopup(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			popupView, popupErr := gui.View(viewActionsPopupName)
			return popupErr == nil && strings.Contains(popupView.Buffer(), "Mark inline comment as resolved")
		}); actualErr != nil {
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
