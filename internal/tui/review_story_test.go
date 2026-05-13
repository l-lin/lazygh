package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestActionsPopup_GivenStoryReviewActionWithoutConfiguredAgent_WhenExecuting_ThenItShowsTheConfigurationErrorOnTheStatusLine(t *testing.T) {
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
	if popupView.Title != "Actions" {
		t.Fatalf("expected popup title %q, actual %q", "Actions", popupView.Title)
	}
	then_statusLineContains(t, gui, "story_review.agent_command")
	if subject.reviewSession.active {
		t.Fatal("expected review mode to stay inactive after the configuration error")
	}
}

func TestReviewStoryMode_GivenTheAgentIsStillRunning_WhenStartingStoryReview_ThenTheStatusLineShowsOnlyTheSpinner(t *testing.T) {
	loader := &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()}}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 42})}
	subject.asyncRunner = asyncRunner
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
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

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued story review load, actual %d", len(asyncRunner.runs))
	}
	then_statusLineIs(t, gui, string(loadingSpinnerFrames[0]))
	if subject.reviewSession.active {
		t.Fatal("expected story review mode to stay inactive until the async load finishes")
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
	for _, unexpected := range []string{"## Files", "internal/tui/render.go", "internal/tui/model.go"} {
		if strings.Contains(detailView.Buffer(), unexpected) {
			t.Fatalf("expected chapter detail to hide impacted files in view 0, unexpected %q in %q", unexpected, detailView.Buffer())
		}
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

func TestReviewStoryMode_GivenTeamOwnedFiles_WhenRenderingTheChapterTree_ThenItShowsTeamOwnersBesideFileNames(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
		fileTeamOwners: map[string]map[string][]string{
			"acme/widgets#42": {
				"internal/tui/render.go": {"P3C"},
				"internal/tui/model.go":  {"FYP"},
			},
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Files: []string{"internal/tui/model.go"}},
		},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingStoryReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	for _, expected := range []string{"internal/tui/render.go  " + reviewDiffTeamOwnershipIcon + " P3C", "internal/tui/model.go  " + reviewDiffTeamOwnershipIcon + " FYP"} {
		if !strings.Contains(filesView.Buffer(), expected) {
			t.Fatalf("expected chapter tree to contain %q, actual %q", expected, filesView.Buffer())
		}
	}
	if !reflect.DeepEqual(loader.fileTeamOwnerCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected team ownership calls %v, actual %v", []string{"acme/widgets#42"}, loader.fileTeamOwnerCalls)
	}
}

func TestReviewStoryMode_GivenAChapterRow_WhenPressingEnter_ThenItOpensViewZeroWithoutTogglingTheChapter(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Files: []string{"internal/tui/model.go"}},
		},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingStoryReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(filesView.Buffer(), " "+reviewModeChapterIcon+" 1 - The Renderer Wakes (1 file)") {
		t.Fatalf("expected the chapter to start expanded with a fold chevron, actual %q", filesView.Buffer())
	}

	openDetailHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyEnter)
	actualErr = openDetailHandler(gui, filesView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	if !strings.Contains(filesView.Buffer(), " "+reviewModeChapterIcon+" 1 - The Renderer Wakes (1 file)") {
		t.Fatalf("expected enter to keep the chapter expanded, actual %q", filesView.Buffer())
	}
	if !strings.Contains(filesView.Buffer(), "render.go") {
		t.Fatalf("expected enter to keep the chapter files visible, actual %q", filesView.Buffer())
	}
}

func TestReviewStoryMode_GivenTheSelectedFileIsInsideAChapter_WhenPressingZA_ThenItTogglesTheContainingChapter(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Files: []string{"internal/tui/model.go"}},
		},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingStoryReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	selectedLineIndex := given_viewLineIndexContaining(t, filesView, "render.go")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, selectedLineIndex, "render.go")

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'z')
	collapseHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'a')
	actualErr = prefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = collapseHandler(gui, filesView)
	then_noError(t, actualErr)
	if !strings.Contains(filesView.Buffer(), " "+reviewModeChapterIcon+" 1 - The Renderer Wakes (1 file)") {
		t.Fatalf("expected za on a chapter file to collapse the containing chapter, actual %q", filesView.Buffer())
	}
	if strings.Contains(filesView.Buffer(), "render.go") {
		t.Fatalf("expected za on a chapter file to hide the chapter files, actual %q", filesView.Buffer())
	}
	chapterLineIndex := given_viewLineIndexContaining(t, filesView, "The Renderer Wakes")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, chapterLineIndex, "The Renderer Wakes")

	actualErr = prefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = collapseHandler(gui, filesView)
	then_noError(t, actualErr)
	if !strings.Contains(filesView.Buffer(), "render.go") {
		t.Fatalf("expected za to reopen the containing chapter, actual %q", filesView.Buffer())
	}
	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)
	selectedLineIndex = given_viewLineIndexContaining(t, filesView, "render.go")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, selectedLineIndex, "render.go")
}

func TestReviewStoryMode_GivenTheChapterTree_WhenPressingZMAndZR_ThenItClosesAndOpensEveryChapterWhileKeepingSelectionOnTheCurrentChapter(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	storyGenerator := &fakeStoryGenerator{review: story.Review{
		Chapters: []story.Chapter{
			{ID: "chapter-1", Title: "The Renderer Wakes", Files: []string{"internal/tui/render.go"}},
			{ID: "chapter-2", Title: "The Model Answers", Files: []string{"internal/tui/model.go"}},
		},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingStoryReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'z')
	closeAllHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'M')
	openAllHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'R')

	actualErr = prefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, filesView)
	then_noError(t, actualErr)
	for _, expected := range []string{" " + reviewModeChapterIcon + " 1 - The Renderer Wakes (1 file)", " " + reviewModeChapterIcon + " 2 - The Model Answers (1 file)"} {
		if !strings.Contains(filesView.Buffer(), expected) {
			t.Fatalf("expected zM to collapse the chapter tree and keep %q visible, actual %q", expected, filesView.Buffer())
		}
	}
	for _, hidden := range []string{"render.go", "model.go"} {
		if strings.Contains(filesView.Buffer(), hidden) {
			t.Fatalf("expected zM to hide %q from the collapsed chapter tree, actual %q", hidden, filesView.Buffer())
		}
	}
	chapterLineIndex := given_viewLineIndexContaining(t, filesView, "The Renderer Wakes")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, chapterLineIndex, "The Renderer Wakes")

	actualErr = prefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, filesView)
	then_noError(t, actualErr)
	for _, expected := range []string{" " + reviewModeChapterIcon + " 1 - The Renderer Wakes (1 file)", " " + reviewModeChapterIcon + " 2 - The Model Answers (1 file)", "render.go", "model.go"} {
		if !strings.Contains(filesView.Buffer(), expected) {
			t.Fatalf("expected zR to reopen the chapter tree and show %q, actual %q", expected, filesView.Buffer())
		}
	}
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, chapterLineIndex, "The Renderer Wakes")
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
