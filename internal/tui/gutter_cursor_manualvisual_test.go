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
	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	gutterCursorBrowserReadyTokenEnv    = "LAZYGH_TMUX_BROWSER_READY_TOKEN"
	gutterCursorBrowserContinueTokenEnv = "LAZYGH_TMUX_BROWSER_CONTINUE_TOKEN"
	gutterCursorReviewReadyTokenEnv     = "LAZYGH_TMUX_REVIEW_READY_TOKEN"
	gutterCursorDoneTokenEnv            = "LAZYGH_TMUX_DONE_TOKEN"
	gutterCursorExitedTokenEnv          = "LAZYGH_TMUX_EXITED_TOKEN"
	gutterCursorLogPathEnv              = "LAZYGH_MANUALVISUAL_LOG"
)

func TestManualVisual_GutterCursor(t *testing.T) {
	browserReadyToken := strings.TrimSpace(os.Getenv(gutterCursorBrowserReadyTokenEnv))
	browserContinueToken := strings.TrimSpace(os.Getenv(gutterCursorBrowserContinueTokenEnv))
	reviewReadyToken := strings.TrimSpace(os.Getenv(gutterCursorReviewReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(gutterCursorDoneTokenEnv))
	exitedToken := strings.TrimSpace(os.Getenv(gutterCursorExitedTokenEnv))
	if browserReadyToken == "" || browserContinueToken == "" || reviewReadyToken == "" || doneToken == "" || exitedToken == "" {
		t.Skip("manualvisual gutter cursor check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	summary, ok := subject.model.SelectedPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}

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
	manualVisualErrCh := make(chan error, 1)
	go func() {
		defer close(pollingStopped)
		manualVisualLogf("starting manualvisual gutter cursor sequence")
		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return prepareManualVisualBrowserChangesGutterState(t, gui, subject)
		}); actualErr != nil {
			manualVisualLogf("browser state error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		manualVisualLogf("browser state ready")
		if actualErr := signalTmuxWaitToken(browserReadyToken); actualErr != nil {
			manualVisualLogf("browser ready token error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(browserContinueToken); actualErr != nil {
			manualVisualLogf("browser continue token error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return prepareManualVisualReviewDiffGutterState(t, gui, subject, summary)
		}); actualErr != nil {
			manualVisualLogf("review state error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		manualVisualLogf("review state ready")
		if actualErr := signalTmuxWaitToken(reviewReadyToken); actualErr != nil {
			manualVisualLogf("review ready token error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(doneToken); actualErr != nil {
			manualVisualLogf("done token error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(exitedToken); actualErr != nil {
			manualVisualLogf("exited token error: %v", actualErr)
			fmt.Println("MANUALVISUAL_ERROR:", actualErr)
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		gui.Update(func(*gocui.Gui) error {
			return gocui.ErrQuit
		})
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

func runManualVisualUpdateAfterMainLoop(gui *gocui.Gui, stop <-chan struct{}, update func(*gocui.Gui) error) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			errCh := make(chan error, 1)
			gui.Update(func(gui *gocui.Gui) error {
				errCh <- update(gui)
				return nil
			})
			return <-errCh
		}
	}
}

func prepareManualVisualBrowserChangesGutterState(_ *testing.T, gui *gocui.Gui, subject *Program) error {
	manualVisualLogf("preparing browser changes state")
	if actualErr := subject.layout(gui); actualErr != nil {
		return actualErr
	}
	if actualErr := subject.openDetail(gui, nil); actualErr != nil {
		return actualErr
	}
	subject.activeDetailTab = ChangesDetailTab
	if actualErr := subject.refreshViews(gui); actualErr != nil {
		return actualErr
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		return actualErr
	}
	if summary, ok := subject.selectedPullRequestSummaryForDetail(); ok {
		if diff, diffOK := subject.pullRequestDiffForSummary(summary); diffOK {
			manualVisualLogf("browser summary ok; diff files=%d err=%v", len(diff.data.Files), diff.err)
		} else {
			manualVisualLogf("browser summary ok; diff missing")
		}
	} else {
		manualVisualLogf("browser summary missing")
	}
	manualVisualLogf("browser detail content preview: %q", subject.detailViewContent())
	return moveManualVisualDiffCursorToLine(gui, subject, "+new line")
}

func prepareManualVisualReviewDiffGutterState(_ *testing.T, gui *gocui.Gui, subject *Program, summary githubdomain.PullRequest) error {
	manualVisualLogf("preparing review diff state")
	subject.startReviewSession(summary, "PRR_pending")
	if actualErr := subject.refreshViews(gui); actualErr != nil {
		return actualErr
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		return actualErr
	}
	return moveManualVisualDiffCursorToLine(gui, subject, "+new line")
}

func moveManualVisualDiffCursorToLine(gui *gocui.Gui, subject *Program, segment string) error {
	manualVisualLogf("moving diff cursor to line containing %q", segment)
	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil {
		return actualErr
	}
	document := subject.currentDetailDocument(detailView)
	lineIndex, actualErr := detailDocumentLineContaining(document, segment)
	if actualErr != nil {
		return actualErr
	}
	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: 0}
	subject.detailViewState.preferredColumn = 0
	subject.detailViewState.sync(document, detailView.InnerHeight())
	return subject.refreshDetailView(gui)
}

func detailDocumentLineContaining(document detailDocument, segment string) (int, error) {
	for lineIndex, line := range document.lines {
		if strings.Contains(string(line), segment) {
			return lineIndex, nil
		}
	}
	visibleLines := make([]string, 0, len(document.lines))
	for lineIndex, line := range document.lines {
		prefix := ""
		if lineIndex < len(document.prefixLines) {
			prefix = string(document.prefixLines[lineIndex].runes)
		}
		visibleLines = append(visibleLines, prefix+string(line))
	}
	manualVisualLogf("missing segment %q in detail lines: %q", segment, strings.Join(visibleLines, " | "))
	return -1, errors.New("missing detail line containing " + segment)
}

func manualVisualLogf(format string, args ...any) {
	logPath := strings.TrimSpace(os.Getenv(gutterCursorLogPathEnv))
	if logPath == "" {
		return
	}
	file, actualErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if actualErr != nil {
		return
	}
	defer file.Close()
	_, _ = fmt.Fprintf(file, format+"\n", args...)
}
