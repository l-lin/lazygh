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
	commentsFilenameWordWrapReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	commentsFilenameWordWrapDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_CommentsTabFilenameWordWrap(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(commentsFilenameWordWrapReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(commentsFilenameWordWrapDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual comments filename word-wrap check needs tmux wait-for tokens")
	}

	longPath := strings.Repeat("comments/very-long-segment/", 4) + "location.go"
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				BaseRefName: "main",
				HeadRefName: "feature/comments-location-wrap",
				State:       "OPEN",
				InlineComments: []githubcli.PullRequestInlineComment{{
					Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:         "Inline body",
					CreatedAt:    "2026-05-05T10:00:00Z",
					Path:         longPath,
					Line:         43,
					OriginalLine: 43,
					Side:         "RIGHT",
					DiffHunk:     "@@ -43,1 +43,1 @@\n-old line\n+new line",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)

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
	if actualErr = subject.nextDetailTab(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := signalTmuxWaitTokenWhenViewAppears(gui, viewDetailName, readyToken, stopPolling); actualErr != nil {
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
