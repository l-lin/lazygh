package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestScreenStateAdapter_GivenBrowserModeProgram_WhenProjectingTheCurrentLayout_ThenItKeepsViewsZeroThroughThreeAndThePullRequestTabs(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	subject.model.FocusPullRequestsView()

	actual := subject.screenState()

	if actual.Mode != ScreenModeBrowser {
		t.Fatalf("expected screen mode %v, actual %v", ScreenModeBrowser, actual.Mode)
	}
	then_screenFocusIs(t, actual, ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)
	then_viewTabLabelIs(t, actual, sidePanelViewTwo, "My PRs")
	for _, expectedView := range []panelViewNumber{mainPanelViewZero, sidePanelViewOne, sidePanelViewTwo, sidePanelViewThree} {
		if _, ok := actual.ViewByNumber(int(expectedView)); !ok {
			t.Fatalf("expected browser mode to expose view %d", expectedView)
		}
	}
}

func TestScreenStateAdapter_GivenReviewModeProgram_WhenProjectingTheCurrentLayout_ThenItRepresentsTheReviewLayoutWithoutNotifications(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", BaseRefName: "main", HeadRefName: "feature/review", State: "OPEN"},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_panelViewContractBrowserModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))

	actual := subject.screenState()

	if actual.Mode != ScreenModeReview {
		t.Fatalf("expected screen mode %v, actual %v", ScreenModeReview, actual.Mode)
	}
	then_screenFocusIs(t, actual, ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)
	if _, ok := actual.ViewByNumber(int(sidePanelViewThree)); ok {
		t.Fatal("expected review mode to hide browser view 3")
	}
	resolver := actual.MainViewResolver()
	if resolver.SourceView.Number != int(sidePanelViewTwo) {
		t.Fatalf("expected review view 0 to resolve from view %d, actual %d", sidePanelViewTwo, resolver.SourceView.Number)
	}

	subject.model.FocusViewNumber(int(mainPanelViewZero))
	actual = subject.screenState()
	then_screenFocusIs(t, actual, ActivePanelMain, mainPanelViewZero, FocusDetailView)
	if !actual.AllowsMainCursor() {
		t.Fatal("expected review view 0 to keep the main cursor")
	}
}

func TestScreenStateAdapter_GivenStoryReviewProgram_WhenProjectingTheCurrentLayout_ThenItRepresentsTheStoryLayoutAndResolvesViewZeroFromTheStoryTree(t *testing.T) {
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
	subject := given_pullRequestCommentProgram(given_panelViewContractBrowserModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.storyGenerator = storyGenerator
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingStoryReviewMode(t, gui, subject))

	actual := subject.screenState()

	if actual.Mode != ScreenModeStoryReview {
		t.Fatalf("expected screen mode %v, actual %v", ScreenModeStoryReview, actual.Mode)
	}
	then_screenFocusIs(t, actual, ActivePanelSide, sidePanelViewTwo, FocusPullRequestsView)
	if _, ok := actual.ViewByNumber(int(sidePanelViewThree)); ok {
		t.Fatal("expected story review mode to hide browser view 3")
	}
	resolver := actual.MainViewResolver()
	if resolver.SourceView.Number != int(sidePanelViewTwo) {
		t.Fatalf("expected story review view 0 to resolve from view %d, actual %d", sidePanelViewTwo, resolver.SourceView.Number)
	}
}

func TestScreenStateAdapter_GivenVisibleOverlays_WhenProjectingTheProgram_ThenTheTopOverlayBecomesTheKeyHintContext(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openActionsPopup(gui, nil))
	if actual := subject.screenState().KeyHintContext(); actual != KeyHintContextActionsPopup {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextActionsPopup, actual)
	}

	subject.model.UpdateActionsPopupSearch("comment on pr", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "comment on pr"))
	then_noError(t, subject.refreshViews(gui))
	then_noError(t, subject.executeSelectedActionsPopupAction(gui, nil))
	if actual := subject.screenState().KeyHintContext(); actual != KeyHintContextModalEditor {
		t.Fatalf("expected key hint context %v, actual %v", KeyHintContextModalEditor, actual)
	}
}

func given_handlerForScreenStateBinding(t *testing.T, subject *Program, viewName string, key any) func(*gocui.Gui, *gocui.View) error {
	t.Helper()
	return given_handlerForBinding(t, subject.keybindingSpecs(), viewName, key)
}
