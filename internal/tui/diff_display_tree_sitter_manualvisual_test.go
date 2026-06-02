//go:build manualvisual

package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	"github.com/l-lin/lazygh/internal/githubcli"
)

const (
	diffDisplayTreeSitterSmallReadyTokenEnv    = "LAZYGH_TMUX_SMALL_READY_TOKEN"
	diffDisplayTreeSitterSmallContinueTokenEnv = "LAZYGH_TMUX_SMALL_CONTINUE_TOKEN"
	diffDisplayTreeSitterLargeReadyTokenEnv    = "LAZYGH_TMUX_LARGE_READY_TOKEN"
	diffDisplayTreeSitterDoneTokenEnv          = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_DiffDisplayTreeSitterFallback(t *testing.T) {
	smallReadyToken := strings.TrimSpace(os.Getenv(diffDisplayTreeSitterSmallReadyTokenEnv))
	smallContinueToken := strings.TrimSpace(os.Getenv(diffDisplayTreeSitterSmallContinueTokenEnv))
	largeReadyToken := strings.TrimSpace(os.Getenv(diffDisplayTreeSitterLargeReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(diffDisplayTreeSitterDoneTokenEnv))
	if smallReadyToken == "" || smallContinueToken == "" || largeReadyToken == "" || doneToken == "" {
		t.Skip("manualvisual tree-sitter fallback check needs tmux wait-for tokens")
	}

	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.pullRequestDiffCache[largeCachedDiffFixturePullRequestKey()] = pullRequestDiffResult{
		data:                    buildReviewDiffData(given_smallGoReviewSessionPullRequestDiff()),
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
			return prepareManualVisualDiffDisplayTreeSitterState(gui, subject, summary)
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(smallReadyToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := waitForTmuxToken(smallContinueToken); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}

		if actualErr := runManualVisualUpdateAfterMainLoop(gui, stopPolling, func(gui *gocui.Gui) error {
			return applyManualVisualDiffDisplayTreeSitterLargeFallback(gui, subject)
		}); actualErr != nil {
			manualVisualErrCh <- actualErr
			gui.Update(func(*gocui.Gui) error { return gocui.ErrQuit })
			return
		}
		if actualErr := signalTmuxWaitToken(largeReadyToken); actualErr != nil {
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

func prepareManualVisualDiffDisplayTreeSitterState(gui *gocui.Gui, subject *Program, summary githubcli.PullRequest) error {
	subject.startReviewSession(summary, largeCachedDiffFixturePendingReviewID)
	if actualErr := subject.afterStateChange(gui); actualErr != nil {
		return actualErr
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		return actualErr
	}
	return refreshManualVisualDiffDisplayTreeSitterDetail(gui, subject)
}

func applyManualVisualDiffDisplayTreeSitterLargeFallback(gui *gocui.Gui, subject *Program) error {
	subject.pullRequestDiffCache[largeCachedDiffFixturePullRequestKey()] = pullRequestDiffResult{
		data:                    buildReviewDiffData(given_largeGoFallbackManualvisualPullRequestDiff()),
		fileTeamOwnersAttempted: true,
	}
	subject.invalidateReviewDiffRenderCache()
	if actualErr := subject.afterStateChange(gui); actualErr != nil {
		return actualErr
	}
	if actualErr := subject.focusDetailView(gui, nil); actualErr != nil {
		return actualErr
	}
	return refreshManualVisualDiffDisplayTreeSitterDetail(gui, subject)
}

func refreshManualVisualDiffDisplayTreeSitterDetail(gui *gocui.Gui, subject *Program) error {
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

func given_smallGoReviewSessionPullRequestDiff() githubcli.PullRequestDiff {
	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -1,0 +1,1 @@",
			`+func addedLine00001() int { return 1 }`,
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{{
			Path:       largeCachedDiffFixtureFilePath,
			ChangeType: "modified",
			Additions:  1,
			Deletions:  0,
		}},
	}
}

func given_largeGoFallbackManualvisualPullRequestDiff() githubcli.PullRequestDiff {
	lineCount := reviewDiffSyntaxHighlightLargeFileLineCount + 10
	unifiedDiffLines := []string{
		"diff --git a/internal/tui/render.go b/internal/tui/render.go",
		"index 1111111..2222222 100644",
		"--- a/internal/tui/render.go",
		"+++ b/internal/tui/render.go",
		fmt.Sprintf("@@ -1,1 +1,%d @@", lineCount),
		`-func removedLine00001() int { return 1 }`,
		`+func addedLine00001() int { return 2 }`,
	}
	for lineNumber := 2; lineNumber <= lineCount; lineNumber++ {
		unifiedDiffLines = append(unifiedDiffLines, fmt.Sprintf(`+func addedLine%05d() int { return %d }`, lineNumber, lineNumber))
	}

	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join(unifiedDiffLines, "\n"),
		Files: []githubcli.PullRequestDiffFile{{
			Path:       largeCachedDiffFixtureFilePath,
			ChangeType: "modified",
			Additions:  lineCount,
			Deletions:  1,
		}},
	}
}
