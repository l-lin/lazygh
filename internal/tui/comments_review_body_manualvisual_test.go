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
	commentsReviewBodyReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	commentsReviewBodyDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_CommentsTabSubmittedReviewBody(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(commentsReviewBodyReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(commentsReviewBodyDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual comments review-body check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				BaseRefName: "main",
				HeadRefName: "feature/comments-review-body",
				State:       "OPEN",
				Reviews: []githubcli.PullRequestReview{{
					ID:          "PRR_1",
					Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:        "looks good but I think you missed `RecommendedContentProvider`",
					State:       "COMMENTED",
					SubmittedAt: "2026-06-15T06:54:59Z",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"looks good but I think you missed `RecommendedContentProvider`": "Rendered review body with `RecommendedContentProvider`",
	}}

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
