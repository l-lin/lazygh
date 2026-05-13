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
	modeDescriptorBrowserReadyTokenEnv  = "LAZYGH_TMUX_BROWSER_READY_TOKEN"
	modeDescriptorReviewReadyTokenEnv   = "LAZYGH_TMUX_REVIEW_READY_TOKEN"
	modeDescriptorBrowserReturnTokenEnv = "LAZYGH_TMUX_BROWSER_RETURN_TOKEN"
	modeDescriptorStoryReadyTokenEnv    = "LAZYGH_TMUX_STORY_READY_TOKEN"
	modeDescriptorDoneTokenEnv          = "LAZYGH_TMUX_DONE_TOKEN"
)

func TestManualVisual_ModeDescriptors(t *testing.T) {
	browserReadyToken := strings.TrimSpace(os.Getenv(modeDescriptorBrowserReadyTokenEnv))
	reviewReadyToken := strings.TrimSpace(os.Getenv(modeDescriptorReviewReadyTokenEnv))
	browserReturnToken := strings.TrimSpace(os.Getenv(modeDescriptorBrowserReturnTokenEnv))
	storyReadyToken := strings.TrimSpace(os.Getenv(modeDescriptorStoryReadyTokenEnv))
	doneToken := strings.TrimSpace(os.Getenv(modeDescriptorDoneTokenEnv))
	if browserReadyToken == "" || reviewReadyToken == "" || browserReturnToken == "" || storyReadyToken == "" || doneToken == "" {
		t.Skip("manualvisual mode-descriptor smoke check needs tmux wait-for tokens")
	}

	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", BaseRefName: "main", HeadRefName: "feature/review", State: "OPEN"},
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
	if actualErr = subject.setKeybindings(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}

	stopPolling := make(chan struct{})
	pollingStopped := make(chan struct{})
	go func() {
		defer close(pollingStopped)

		if actualErr := signalTmuxWaitTokenWhenCondition(gui, browserReadyToken, stopPolling, func(gui *gocui.Gui) bool {
			if subject.modeDescriptor().Mode() != ScreenModeBrowser {
				return false
			}
			_, notificationsErr := gui.View(viewNotificationsName)
			return notificationsErr == nil
		}); actualErr != nil {
			return
		}
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, reviewReadyToken, stopPolling, func(gui *gocui.Gui) bool {
			if subject.modeDescriptor().Mode() != ScreenModeReview {
				return false
			}
			if _, notificationsErr := gui.View(viewNotificationsName); notificationsErr == nil {
				return false
			}
			detailView, detailErr := gui.View(viewDetailName)
			return detailErr == nil && detailView.Title == reviewModeDiffTitle
		}); actualErr != nil {
			return
		}
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, browserReturnToken, stopPolling, func(gui *gocui.Gui) bool {
			if subject.modeDescriptor().Mode() != ScreenModeBrowser {
				return false
			}
			_, notificationsErr := gui.View(viewNotificationsName)
			return notificationsErr == nil
		}); actualErr != nil {
			return
		}
		if actualErr := signalTmuxWaitTokenWhenCondition(gui, storyReadyToken, stopPolling, func(gui *gocui.Gui) bool {
			if subject.modeDescriptor().Mode() != ScreenModeStoryReview {
				return false
			}
			if _, notificationsErr := gui.View(viewNotificationsName); notificationsErr == nil {
				return false
			}
			detailView, detailErr := gui.View(viewDetailName)
			return detailErr == nil && detailView.Title == reviewModeChapterTitle
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

func signalTmuxWaitTokenWhenCondition(gui *gocui.Gui, token string, stop <-chan struct{}, condition func(*gocui.Gui) bool) error {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			matched := make(chan bool, 1)
			gui.Update(func(gui *gocui.Gui) error {
				matched <- condition(gui)
				return nil
			})
			if !<-matched {
				continue
			}
			return signalTmuxWaitToken(token)
		}
	}
}
