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
	diffDisplayCacheInvalidationReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	diffDisplayCacheInvalidationDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DiffDisplayInlineThreadAfterCacheInvalidation(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(diffDisplayCacheInvalidationReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(diffDisplayCacheInvalidationDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual cache-invalidation check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	subject.asyncRunner = inlineAsyncRunner{}

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
	if actualErr = subject.openDetail(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	subject.detailState.activeTab = ChangesDetailTab
	if actualErr = subject.afterStateChange(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.focusDetailView(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	manualVisualErrCh := make(chan error, 1)
	go func() {
		defer close(pollingStopped)
		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			if actualErr := subject.openActionsPopup(gui, nil); actualErr != nil {
				return actualErr
			}
			subject.model.UpdateActionsPopupSearch("mark inline comment as resolved", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "mark inline comment as resolved"))
			if actualErr := subject.afterStateChange(gui); actualErr != nil {
				return actualErr
			}
			return subject.executeSelectedActionsPopupAction(gui, nil)
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			detailView, detailErr := gui.View(viewDetailName)
			if detailErr != nil {
				return false
			}
			buffer := detailView.Buffer()
			return strings.Contains(buffer, "Resolved") && !strings.Contains(buffer, "Rendered original inline body")
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(doneToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
	}()

	actualErr = gui.MainLoop()
	close(stopPolling)
	<-pollingStopped
	select {
	case manualVisualErr := <-manualVisualErrCh:
		t.Fatalf("expected no manualvisual error, actual %v", manualVisualErr)
	default:
	}
	if actualErr != nil && !errors.Is(actualErr, gocui.ErrQuit) {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
}
