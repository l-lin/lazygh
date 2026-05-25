package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestOpenStoryReviewByURL_GivenNoConfiguredAgent_WhenOpening_ThenItReturnsTheConfigurationError(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_model(), &fakePullRequestDetailLoader{})

	actualErr := subject.OpenStoryReviewByURL("https://github.com/acme/rocket/pulls/77")

	if actualErr == nil || actualErr.Error() != storyReviewConfigureAgentMessage {
		t.Fatalf("expected error %q, actual %v", storyReviewConfigureAgentMessage, actualErr)
	}
}

func TestOpenStoryReviewByURL_GivenAValidGitHubPullRequestsURLAfterLayout_WhenOpening_ThenItRefreshesThroughDispatch(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#77": {
				Title:        "Rocket PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/story-url",
				State:        "OPEN",
				ChangedFiles: 1,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Chapters: []story.Chapter{{
			ID:        "chapter-1",
			Title:     "Through the Diff",
			Narrative: "## Chapter 1\nA measured story.",
			Files:     []string{"internal/tui/render.go"},
		}},
	}}
	subject := given_pullRequestCommentProgram(given_model(), loader)
	subject.storyGenerator = storyGenerator
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	actualErr = subject.OpenStoryReviewByURL("https://github.com/acme/rocket/pulls/77")
	then_noError(t, actualErr)

	if !subject.navigationState.reviewSession.active {
		t.Fatal("expected story review mode to become active immediately")
	}
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeChapterTitle {
		t.Fatalf("expected detail view title %q, actual %q", reviewModeChapterTitle, detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), "A measured story") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "A measured story", detailView.Buffer())
	}
}

func TestOpenStoryReviewByURL_GivenAValidGitHubPullRequestsURLBeforeLayout_WhenRendering_ThenItStartsDirectlyInStoryReviewMode(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		details: map[string]githubcli.PullRequestDetail{
			"acme/rocket#77": {
				Title:        "Rocket PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/story-url",
				State:        "OPEN",
				ChangedFiles: 1,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/rocket#77": given_reviewSessionPullRequestDiff(),
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Chapters: []story.Chapter{{
			ID:        "chapter-1",
			Title:     "Through the Diff",
			Narrative: "## Chapter 1\nA measured story.",
			Files:     []string{"internal/tui/render.go"},
		}},
	}}
	subject := given_pullRequestCommentProgram(given_model(), loader)
	subject.storyGenerator = storyGenerator
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})

	actualErr := subject.OpenStoryReviewByURL("https://github.com/acme/rocket/pulls/77")
	then_noError(t, actualErr)

	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	if !subject.navigationState.reviewSession.active {
		t.Fatal("expected story review mode to be active")
	}
	if subject.navigationState.reviewSession.mode != reviewSessionModeStory {
		t.Fatalf("expected review session mode %v, actual %v", reviewSessionModeStory, subject.navigationState.reviewSession.mode)
	}
	if subject.navigationState.reviewSession.pendingReviewID != "PRR_story" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_story", subject.navigationState.reviewSession.pendingReviewID)
	}
	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/rocket#77"}) {
		t.Fatalf("expected start review calls %v, actual %v", []string{"acme/rocket#77"}, loader.startReviewCalls)
	}
	if len(storyGenerator.requests) != 1 {
		t.Fatalf("expected one story generation request, actual %d", len(storyGenerator.requests))
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeChapterTitle {
		t.Fatalf("expected detail view title %q, actual %q", reviewModeChapterTitle, detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), "A measured story") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "A measured story", detailView.Buffer())
	}
}
