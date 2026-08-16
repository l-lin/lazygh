//go:build manualvisual

package tui

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestManualVisual_StoryReviewLoadingStatusLine(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(storyReviewReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(storyReviewDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual story-review loading check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42})}
	subject.asyncRunner = asyncRunner
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "--models", "anthropic/claude-sonnet-4-6", "--no-session", "-p", "@{{prompt_file}}"}})

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
	if actualErr = subject.openActionsPopup(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	subject.model.UpdateActionsPopupSearch("story", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "story"))
	if actualErr = subject.afterStateChange(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.executeSelectedActionsPopupAction(gui, nil); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	expectedStatus := formatRunningCommandStatus("pi --models anthropic/claude-sonnet-4-6 --no-session -p @<prompt_file>")
	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			statusView, actualErr := gui.View(viewStatusLineName)
			if actualErr != nil {
				return false
			}
			statusBuffer := statusView.Buffer()
			return strings.Contains(statusBuffer, string(loadingSpinnerFrames[0])) && strings.Contains(statusBuffer, expectedStatus)
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
