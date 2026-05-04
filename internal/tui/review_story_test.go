package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/story"
)

func TestActionsPopup_GivenStoryReviewActionWithoutConfiguredAgent_WhenExecuting_ThenItKeepsThePopupOpenAndShowsHowToConfigureIt(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("story", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "story"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Title, "story_review.agent_command") {
		t.Fatalf("expected popup title to contain %q, actual %q", "story_review.agent_command", popupView.Title)
	}
	if subject.reviewSession.active {
		t.Fatal("expected review mode to stay inactive after the configuration error")
	}
}

func TestReviewStoryMode_GivenGeneratedChapters_WhenExecutingTheAction_ThenItShowsChapterNarrativeAndNestedFiles(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/story",
				ChangedFiles: 2,
				Author:       &githubcli.PullRequestAuthor{Login: "octocat"},
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Summary: "A calmer way to review the pull request.",
		Chapters: []story.Chapter{
			{
				ID:        "chapter-1",
				Title:     "The Renderer Wakes",
				Narrative: "## Chapter 1\nThis chapter explains the rendering shift.",
				Files:     []string{"internal/tui/render.go"},
			},
			{
				ID:        "chapter-2",
				Title:     "The Model Answers",
				Narrative: "## Chapter 2\nThis chapter explains the model update.",
				Files:     []string{"internal/tui/model.go"},
			},
		},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{
		AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"},
		Prompt:       "Tell the story with dry professionalism.",
	})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingStoryReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	if !subject.reviewSession.active {
		t.Fatal("expected review mode to be active")
	}
	if subject.reviewSession.pendingReviewID != "PRR_story" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_story", subject.reviewSession.pendingReviewID)
	}
	if len(storyGenerator.configs) != 1 {
		t.Fatalf("expected one story review generation call, actual %d", len(storyGenerator.configs))
	}
	if storyGenerator.configs[0].Prompt != "Tell the story with dry professionalism." {
		t.Fatalf("expected configured prompt %q, actual %q", "Tell the story with dry professionalism.", storyGenerator.configs[0].Prompt)
	}
	if !strings.Contains(storyGenerator.requests[0].DiffText, "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected the story request to include the diff text, actual %q", storyGenerator.requests[0].DiffText)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if filesView.Title != reviewModeChaptersTitle {
		t.Fatalf("expected files view title %q, actual %q", reviewModeChaptersTitle, filesView.Title)
	}
	if !strings.Contains(filesView.Buffer(), "The Renderer Wakes") || !strings.Contains(filesView.Buffer(), "render.go") || !strings.Contains(filesView.Buffer(), "The Model Answers") {
		t.Fatalf("expected chapter tree buffer to contain the generated chapters, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeChapterTitle {
		t.Fatalf("expected detail view title %q, actual %q", reviewModeChapterTitle, detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), "This chapter explains the rendering shift") {
		t.Fatalf("expected chapter narrative in detail view, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected chapter view to hide raw diff text, actual %q", detailView.Buffer())
	}

	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDiffTitle {
		t.Fatalf("expected detail view title %q after selecting a file, actual %q", reviewModeDiffTitle, detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected file diff in detail view after selecting a file, actual %q", detailView.Buffer())
	}
}

func given_startingStoryReviewMode(t *testing.T, gui *gocui.Gui, subject *Program) error {
	t.Helper()

	actualErr := subject.openActionsPopup(gui, nil)
	if actualErr != nil {
		return actualErr
	}
	subject.model.UpdateActionsPopupSearch("story", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "story"))
	actualErr = subject.refreshViews(gui)
	if actualErr != nil {
		return actualErr
	}

	return subject.executeSelectedActionsPopupAction(gui, nil)
}

type fakeStoryGenerator struct {
	configs  []story.Config
	requests []story.Request
	review   story.Review
	err      error
}

func (generator *fakeStoryGenerator) Generate(config story.Config, request story.Request) (story.Review, error) {
	generator.configs = append(generator.configs, config)
	generator.requests = append(generator.requests, request)
	if generator.err != nil {
		return story.Review{}, generator.err
	}
	return generator.review, nil
}
