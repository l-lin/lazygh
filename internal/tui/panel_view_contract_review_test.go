package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestPanelViewContracts_GivenReviewMode_WhenTheSidePanelIsReplaced_ThenViewZeroRemainsTheMainPanelAndKeepsItsCursorRules(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_panelViewContractBrowserModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))
	then_viewDoesNotExist(t, gui, viewNotificationsName)
	for _, viewNumber := range []panelViewNumber{mainPanelViewZero, sidePanelViewOne, sidePanelViewTwo} {
		then_panelViewShowsVisibleNumber(t, gui, viewNumber)
	}

	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if gui.Cursor {
		t.Fatal("expected review side-panel focus to hide the main cursor")
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDiffTitle {
		t.Fatalf("expected review view 0 title %q, actual %q", reviewModeDiffTitle, detailView.Title)
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '0')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)
	if !gui.Cursor {
		t.Fatal("expected review view 0 focus to show the main cursor")
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(mainPanelViewZero), '1')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewOne)
	if gui.Cursor {
		t.Fatal("expected review side-panel focus to hide the main cursor again")
	}
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDescriptionTitle {
		t.Fatalf("expected review view 0 title %q when view 1 drives the detail pane, actual %q", reviewModeDescriptionTitle, detailView.Title)
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewOne), '2')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if gui.Cursor {
		t.Fatal("expected review files focus to keep the main cursor hidden")
	}
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDiffTitle {
		t.Fatalf("expected review view 0 title %q when view 2 drives the detail pane, actual %q", reviewModeDiffTitle, detailView.Title)
	}
}

func TestPanelViewContracts_GivenStoryReviewMode_WhenTheSidePanelIsReplaced_ThenViewZeroRemainsTheMainPanelAndKeepsItsCursorRules(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_story",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/story",
				State:       "OPEN",
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
	subject := given_pullRequestCommentProgram(given_panelViewContractBrowserModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingStoryReviewMode(t, gui, subject))
	then_viewDoesNotExist(t, gui, viewNotificationsName)
	for _, viewNumber := range []panelViewNumber{mainPanelViewZero, sidePanelViewOne, sidePanelViewTwo} {
		then_panelViewShowsVisibleNumber(t, gui, viewNumber)
	}

	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if gui.Cursor {
		t.Fatal("expected story-review side-panel focus to hide the main cursor")
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeChapterTitle {
		t.Fatalf("expected story-review view 0 title %q, actual %q", reviewModeChapterTitle, detailView.Title)
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), '0')(gui, nil))
	then_currentPanelViewIs(t, gui, mainPanelViewZero)
	if !gui.Cursor {
		t.Fatal("expected story-review view 0 focus to show the main cursor")
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(mainPanelViewZero), '2')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	if gui.Cursor {
		t.Fatal("expected story-review side-panel focus to hide the main cursor again")
	}

	then_noError(t, given_handlerForBinding(t, subject.keybindingSpecs(), panelViewName(sidePanelViewTwo), 'j')(gui, nil))
	then_currentPanelViewIs(t, gui, sidePanelViewTwo)
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDiffTitle {
		t.Fatalf("expected story-review view 0 title %q after selecting a file row, actual %q", reviewModeDiffTitle, detailView.Title)
	}
	if gui.Cursor {
		t.Fatal("expected story-review file selection to keep the main cursor hidden")
	}
}
