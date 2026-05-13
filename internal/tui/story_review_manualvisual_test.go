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
	"github.com/l-lin/lazygh/internal/story"
)

const (
	storyReviewReadyTokenEnv = "LAZYGH_TMUX_READY_TOKEN"
	storyReviewDoneTokenEnv  = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_StoryReviewModeLayout(t *testing.T) {
	readyToken := strings.TrimSpace(os.Getenv(storyReviewReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(storyReviewDoneTokenEnv))
	if readyToken == "" || doneToken == "" {
		t.Skip("manualvisual story-review smoke check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", BaseRefName: "main", HeadRefName: "feature/story", State: "OPEN"},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Summary: "A calmer way to review the pull request.",
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Narrative: "## Chapter 1", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Narrative: "## Chapter 2", Files: []string{"internal/tui/model.go"}},
		},
	}}
	subject := given_pullRequestCommentProgram(given_panelViewContractManualVisualModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator

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
	if actualErr = given_startingStoryReviewMode(t, gui, subject); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, readyToken, stopPolling, func(gui *gocui.Gui) bool {
			if subject.modeDescriptor().Mode() != ScreenModeStoryReview {
				return false
			}
			if _, notificationsErr := gui.View(viewNotificationsName); notificationsErr == nil {
				return false
			}
			detailView, detailErr := gui.View(viewDetailName)
			if detailErr != nil || detailView.Title != reviewModeChapterTitle {
				return false
			}
			pullRequestsView, pullRequestsErr := gui.View(viewPullRequestsName)
			return pullRequestsErr == nil && pullRequestsView.Title == reviewModeChaptersTitle
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

func waitForStoryReviewCondition(gui *gocui.Gui, condition func(*gocui.Gui) bool) bool {
	matched := make(chan bool, 1)
	gui.Update(func(gui *gocui.Gui) error {
		matched <- condition(gui)
		return nil
	})
	return <-matched
}

func waitForStoryReviewReady(gui *gocui.Gui, stop <-chan struct{}, condition func(*gocui.Gui) bool) bool {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return false
		case <-ticker.C:
			if waitForStoryReviewCondition(gui, condition) {
				return true
			}
		}
	}
}
