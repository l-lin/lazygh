package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
)

func TestModeDescriptor_GivenBrowserMode_WhenDescribingTheSidebar_ThenItKeepsViewsOneTwoAndThreeWithPullRequestTabs(t *testing.T) {
	subject := NewProgramWithModel(given_panelViewContractBrowserModel())

	descriptor := subject.modeDescriptor()
	actual := descriptor.SidebarSchema(subject)

	if descriptor.Mode() != ScreenModeBrowser {
		t.Fatalf("expected mode %v, actual %v", ScreenModeBrowser, descriptor.Mode())
	}
	then_sidebarSchemaViewNumbersAre(t, actual, []panelViewNumber{sidePanelViewOne, sidePanelViewTwo, sidePanelViewThree})
	viewTwo, ok := actual.ViewByNumber(int(sidePanelViewTwo))
	if !ok {
		t.Fatalf("expected sidebar view %d", sidePanelViewTwo)
	}
	if len(viewTwo.Tabs) != 2 {
		t.Fatalf("expected browser view 2 to keep two tabs, actual %+v", viewTwo.Tabs)
	}
	if actualLabel := viewTwo.Tabs[viewTwo.ActiveTab].Label; actualLabel != "My PRs" {
		t.Fatalf("expected active tab label %q, actual %q", "My PRs", actualLabel)
	}
}

func TestModeDescriptor_GivenReviewMode_WhenDescribingTheSidebar_ThenItSwapsOutNotificationsAndResolvesViewZeroFromTheFilesPane(t *testing.T) {
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

	descriptor := subject.modeDescriptor()
	actual := descriptor.SidebarSchema(subject)
	resolver := subject.mainViewResolver()

	if descriptor.Mode() != ScreenModeReview {
		t.Fatalf("expected mode %v, actual %v", ScreenModeReview, descriptor.Mode())
	}
	then_sidebarSchemaViewNumbersAre(t, actual, []panelViewNumber{sidePanelViewOne, sidePanelViewTwo})
	if _, ok := actual.ViewByNumber(int(sidePanelViewThree)); ok {
		t.Fatal("expected review mode to hide sidebar view 3")
	}
	if resolver.SourceView.Number != int(sidePanelViewTwo) {
		t.Fatalf("expected resolver source view %d, actual %d", sidePanelViewTwo, resolver.SourceView.Number)
	}
	if resolver.ContentKind != MainContentKindReviewDiff {
		t.Fatalf("expected content kind %v, actual %v", MainContentKindReviewDiff, resolver.ContentKind)
	}
}

func TestModeDescriptor_GivenStoryReviewMode_WhenDescribingTheSidebar_ThenItUsesTheStorySelectionToResolveViewZero(t *testing.T) {
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

	descriptor := subject.modeDescriptor()
	actual := descriptor.SidebarSchema(subject)
	resolver := subject.mainViewResolver()

	if descriptor.Mode() != ScreenModeStoryReview {
		t.Fatalf("expected mode %v, actual %v", ScreenModeStoryReview, descriptor.Mode())
	}
	then_sidebarSchemaViewNumbersAre(t, actual, []panelViewNumber{sidePanelViewOne, sidePanelViewTwo})
	if resolver.SourceView.Number != int(sidePanelViewTwo) {
		t.Fatalf("expected resolver source view %d, actual %d", sidePanelViewTwo, resolver.SourceView.Number)
	}
	if resolver.ContentKind != MainContentKindStoryChapter {
		t.Fatalf("expected content kind %v, actual %v", MainContentKindStoryChapter, resolver.ContentKind)
	}
}

func TestModeDescriptor_GivenBrowserMode_WhenStartingReviewAndExiting_ThenTheSchemaReturnsWithoutBrokenViewNumbering(t *testing.T) {
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
	then_noError(t, subject.exitReviewMode(gui, nil))

	descriptor := subject.modeDescriptor()
	actual := descriptor.SidebarSchema(subject)

	if descriptor.Mode() != ScreenModeBrowser {
		t.Fatalf("expected mode %v after exit, actual %v", ScreenModeBrowser, descriptor.Mode())
	}
	then_sidebarSchemaViewNumbersAre(t, actual, []panelViewNumber{sidePanelViewOne, sidePanelViewTwo, sidePanelViewThree})
	if _, ok := subject.screenState().ViewByNumber(int(sidePanelViewThree)); !ok {
		t.Fatal("expected browser mode to restore view 3 after exiting review mode")
	}
}

func TestActionContext_GivenReviewModeAndNotificationsView_WhenResolving_ThenItTracksPullRequestReviewAndNotificationContexts(t *testing.T) {
	model := given_model()
	model.FocusNotificationsView()
	browser := NewProgramWithModel(model)

	actualBrowser := browser.actionContext()
	if !actualBrowser.IsNotificationContext() {
		t.Fatal("expected browser notifications focus to be a notification context")
	}
	if actualBrowser.IsReviewContext() {
		t.Fatal("expected browser notifications focus to stay outside review context")
	}

	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs:         map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	review := given_pullRequestCommentProgram(given_panelViewContractBrowserModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	review.configureGUI(gui)

	then_noError(t, review.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, review))

	actualReview := review.actionContext()
	if !actualReview.IsPullRequestContext() {
		t.Fatal("expected review mode to keep pull request actions active")
	}
	if !actualReview.IsReviewContext() {
		t.Fatal("expected review mode to resolve as review context")
	}
	if actualReview.IsNotificationContext() {
		t.Fatal("expected review mode to stay outside notification context")
	}
}

func TestInputContext_GivenBrowserDescriptionBrowserChangesAndReviewDiff_WhenResolving_ThenItChoosesTheCorrectDetailInputMode(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	browser := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	browserGUI := given_headlessGui(t)
	browser.configureGUI(browserGUI)

	then_noError(t, browser.layout(browserGUI))
	then_noError(t, browser.openDetail(browserGUI, nil))
	then_noError(t, browser.afterStateChange(browserGUI))
	if actual := browser.inputContext().DetailInputMode; actual != DetailInputModePullRequestComment {
		t.Fatalf("expected input mode %v, actual %v", DetailInputModePullRequestComment, actual)
	}

	browser.activeDetailTab = ChangesDetailTab
	then_noError(t, browser.afterStateChange(browserGUI))
	if actual := browser.inputContext().DetailInputMode; actual != DetailInputModeBrowserChangesInlineComment {
		t.Fatalf("expected input mode %v, actual %v", DetailInputModeBrowserChangesInlineComment, actual)
	}

	browserGUI.Close()

	reviewLoader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs:         map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	review := given_pullRequestCommentProgram(given_pullRequestCommentModel(), reviewLoader)
	reviewGUI := given_headlessGui(t)
	defer reviewGUI.Close()
	review.configureGUI(reviewGUI)
	then_noError(t, review.layout(reviewGUI))
	then_noError(t, given_startingReviewMode(t, reviewGUI, review))
	if !review.inputContext().SearchUsesReviewTree {
		t.Fatal("expected review files focus to use the review-tree search context")
	}
	then_noError(t, review.focusDetailView(reviewGUI, nil))
	if review.inputContext().SearchUsesReviewTree {
		t.Fatal("expected review detail focus to stop using the review-tree search context")
	}
	if actual := review.inputContext().DetailInputMode; actual != DetailInputModeReviewInlineComment {
		t.Fatalf("expected input mode %v, actual %v", DetailInputModeReviewInlineComment, actual)
	}
}

func then_sidebarSchemaViewNumbersAre(t *testing.T, schema SidebarSchema, expected []panelViewNumber) {
	t.Helper()

	actual := schema.ViewNumbers()
	if len(actual) != len(expected) {
		t.Fatalf("expected view numbers %v, actual %v", expected, actual)
	}
	for index := range expected {
		if actual[index] != int(expected[index]) {
			t.Fatalf("expected view numbers %v, actual %v", expected, actual)
		}
	}
}
