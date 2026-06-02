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
	diffDisplayInitialReadyTokenEnv     = "LAZYGH_TMUX_INITIAL_READY_TOKEN"
	diffDisplayInitialContinueTokenEnv  = "LAZYGH_TMUX_INITIAL_CONTINUE_TOKEN"
	diffDisplayPageDownReadyTokenEnv    = "LAZYGH_TMUX_PAGE_DOWN_READY_TOKEN"
	diffDisplayPageDownContinueTokenEnv = "LAZYGH_TMUX_PAGE_DOWN_CONTINUE_TOKEN"
	diffDisplayCursorReadyTokenEnv      = "LAZYGH_TMUX_CURSOR_READY_TOKEN"
	diffDisplayCursorContinueTokenEnv   = "LAZYGH_TMUX_CURSOR_CONTINUE_TOKEN"
	diffDisplayPageUpReadyTokenEnv      = "LAZYGH_TMUX_PAGE_UP_READY_TOKEN"
	diffDisplayDoneTokenEnv             = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DiffDisplayViewport(t *testing.T) {
	initialReadyToken := strings.TrimSpace(os.Getenv(diffDisplayInitialReadyTokenEnv))
	initialContinueToken := strings.TrimSpace(os.Getenv(diffDisplayInitialContinueTokenEnv))
	pageDownReadyToken := strings.TrimSpace(os.Getenv(diffDisplayPageDownReadyTokenEnv))
	pageDownContinueToken := strings.TrimSpace(os.Getenv(diffDisplayPageDownContinueTokenEnv))
	cursorReadyToken := strings.TrimSpace(os.Getenv(diffDisplayCursorReadyTokenEnv))
	cursorContinueToken := strings.TrimSpace(os.Getenv(diffDisplayCursorContinueTokenEnv))
	pageUpReadyToken := strings.TrimSpace(os.Getenv(diffDisplayPageUpReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(diffDisplayDoneTokenEnv))
	if initialReadyToken == "" || initialContinueToken == "" || pageDownReadyToken == "" || pageDownContinueToken == "" || cursorReadyToken == "" || cursorContinueToken == "" || pageUpReadyToken == "" || doneToken == "" {
		t.Skip("manualvisual diff-display viewport check needs tmux wait-for tokens")
	}

	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache[largeCachedDiffFixturePullRequestKey()] = pullRequestDiffResult{
		data:                    buildReviewDiffData(given_largeCachedReviewSessionPullRequestDiff(200)),
		fileTeamOwnersAttempted: true,
	}
	summary := githubcli.PullRequest{
		Title:      "First PR",
		Number:     largeCachedDiffFixturePullRequestNumber,
		Repository: githubcli.Repository{NameWithOwner: largeCachedDiffFixtureRepositoryName},
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
		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return prepareManualVisualDiffDisplayViewportState(gui, subject, summary)
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(initialReadyToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(initialContinueToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}

		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return applyManualVisualDiffDisplayViewportStep(gui, subject, func(gui *gocui.Gui, detailView *gocui.View) error {
				return subject.fullPageDown(gui, detailView)
			})
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(pageDownReadyToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(pageDownContinueToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}

		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return applyManualVisualDiffDisplayViewportStep(gui, subject, func(gui *gocui.Gui, detailView *gocui.View) error {
				stepCount := maxInt(1, detailView.InnerHeight()/2)
				for range stepCount {
					if actualErr := subject.moveSelectionDown(gui, detailView); actualErr != nil {
						return actualErr
					}
				}
				return nil
			})
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(cursorReadyToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(cursorContinueToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}

		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return applyManualVisualDiffDisplayViewportStep(gui, subject, func(gui *gocui.Gui, detailView *gocui.View) error {
				return subject.fullPageUp(gui, detailView)
			})
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(pageUpReadyToken); actualErr != nil {
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

func prepareManualVisualDiffDisplayViewportState(gui *gocui.Gui, subject *Program, summary githubcli.PullRequest) error {
	subject.startReviewSession(summary, largeCachedDiffFixturePendingReviewID)
	if actualErr := subject.afterStateChange(gui); actualErr != nil {
		return actualErr
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		return actualErr
	}
	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil {
		return actualErr
	}
	document := subject.currentDetailDocument(detailView)
	subject.detailState.viewState.cursor = document.positionForRow(0, 0)
	subject.detailState.viewState.preferredColumn = 0
	subject.detailState.viewState.sync(document, detailView.InnerHeight())
	return subject.refreshDetailView(gui)
}

func applyManualVisualDiffDisplayViewportStep(gui *gocui.Gui, subject *Program, step func(*gocui.Gui, *gocui.View) error) error {
	detailView, actualErr := gui.View(viewDetailName)
	if actualErr != nil {
		return actualErr
	}
	if actualErr := step(gui, detailView); actualErr != nil {
		return actualErr
	}
	return subject.refreshDetailView(gui)
}
