package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/story"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestStatusLineKeyHints_GivenActivePullRequestsView_WhenRendering_ThenItShowsDarkGreyCommaSeparatedLowercaseHintsRightAlignedOnTheBottomRow(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "?: help, /: search, a: action", given_themeColorHex(t, theme.InactiveTitleHex), "status line key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func TestStatusLineKeyHints_GivenConfiguredKeyOverrides_WhenRendering_ThenItUsesTheResolvedKeysInKeyColonDescriptionFormat(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_pullRequestCommentModel(), appconfig.KeymapOverrides{
		"global": {
			"toggle_help":        {"!"},
			"open_search":        {"s", "ctrl+s"},
			"open_actions_popup": {"p"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "!: help, s/Ctrl+S: search, p: action")
}

func TestStatusLineKeyHints_GivenActionsPopupVisible_WhenRendering_ThenItShowsLowercasePopupHintsIncludingCancelRightAlignedOnTheBottomRow(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: execute, Escape: cancel")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: execute, Escape: cancel", given_themeColorHex(t, theme.InactiveTitleHex), "actions popup key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: execute, Escape: cancel")
}

func TestStatusLineKeyHints_GivenActionsPopupSearchVisible_WhenRendering_ThenItShowsLowercasePopupSearchHintsIncludingCancelRightAlignedOnTheBottomRow(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusActionsPopupSearch(gui, nil)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: execute, Escape: cancel")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: execute, Escape: cancel", given_themeColorHex(t, theme.InactiveTitleHex), "actions popup search key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: execute, Escape: cancel")
}

func TestStatusLineKeyHints_GivenAssigneePickerVisible_WhenRendering_ThenItShowsDarkGreyPopupHintsRightAlignedOnTheBottomRow(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), given_pullRequestAssigneeLoader())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)

	then_statusLineKeyHintsAre(t, gui, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: toggle, Alt+Enter: submit, Escape: cancel")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: toggle, Alt+Enter: submit, Escape: cancel", given_themeColorHex(t, theme.InactiveTitleHex), "assignee picker key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Ctrl+N/↓: next, Ctrl+P/↑: previous, Enter: toggle, Alt+Enter: submit, Escape: cancel")
}

func TestStatusLineKeyHints_GivenBuildRunPopupVisible_WhenRendering_ThenItShowsPopupHintsRightAlignedOnTheBottomRow(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openPullRequestBuildRunPopup(gui, pullRequestBuildRunPopupContent{
		checkTitle: "CI / test",
		runURL:     "https://github.com/acme/widgets/actions/runs/42",
		body:       "Run #42\nStatus: completed",
	})
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "/: search, Alt+B: browser, y: yank, Alt+Y: copy, Escape: back")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "/: search, Alt+B: browser, y: yank, Alt+Y: copy, Escape: back", given_themeColorHex(t, theme.InactiveTitleHex), "build run popup key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "/: search, Alt+B: browser, y: yank, Alt+Y: copy, Escape: back")
}

func TestPaneFooter_GivenFocusedViewOneWithoutSearchSummary_WhenRendering_ThenItShowsNoPaneFooterAndTheResolvedKeyHints(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewUserFooterName)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_statusLineIs(t, gui, "")
}

func TestPaneFooter_GivenFocusedViewOneWithASearchSummary_WhenRendering_ThenItShowsTheSearchSummaryAndKeyHints(t *testing.T) {
	model := given_model()
	model.StartSearch()
	model.UpdateSearchDraft("2")
	model.SubmitSearch()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_footerTextIs(t, gui, viewUserFooterName, "/2 (1 match)")
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_statusLineIs(t, gui, "")
}

func TestStatusLineKeyHints_GivenGlobalActionOverride_WhenRenderingBrowserModePullRequestsAndDetail_ThenEachPaneUsesTheSameSharedActionKey(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/detail",
				State:       "OPEN",
			},
		},
	}
	model := given_pullRequestCommentModel()
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"global": {
			"open_actions_popup": {"p"},
		},
	})
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, p: action")

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, p: action")
}

func TestStatusLineKeyHints_GivenGlobalActionOverride_WhenRenderingReviewModeFilesAndDiff_ThenEachPaneUsesTheSameSharedActionKey(t *testing.T) {
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
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"global": {
			"open_actions_popup": {"p"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	actualErr = subject.focusPullRequestsView(gui, nil)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, p: action")

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, p: action")
}

func TestStatusLineKeyHints_GivenPullRequestCommentOverride_WhenRenderingPullRequestsView_ThenItShowsTheResolvedCommentHint(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_pullRequestCommentModel(), appconfig.KeymapOverrides{
		"global": {
			"open_actions_popup": {"alt+a"},
		},
		"pull_requests": {
			"comment_on_pull_request": {"alt+c"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, Alt+C: comment, Alt+A: action")
}

func TestStatusLineKeyHints_GivenPullRequestCommentOverride_WhenRenderingBrowserChangesDetail_ThenItShowsTheResolvedCommentHint(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"global": {
			"open_actions_popup": {"alt+a"},
		},
		"pull_requests": {
			"comment_on_pull_request": {"alt+c"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")
	actualErr := subject.afterStateChange(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, Alt+C: comment, Alt+A: action")
}

func TestStatusLineKeyHints_GivenPullRequestCommentOverride_WhenRenderingStoryReviewDiff_ThenItShowsTheResolvedCommentHint(t *testing.T) {
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
		Chapters: []story.Chapter{{
			ID:        "chapter-1",
			Title:     "The Renderer Wakes",
			Narrative: "## Chapter 1\nThis chapter explains the rendering shift.",
			Files:     []string{"internal/tui/render.go"},
		}},
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyStoryReviewConfig(story.Config{AgentCommand: []string{"pi", "-p", "@{{prompt_file}}"}})
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"global": {
			"open_actions_popup": {"alt+a"},
		},
		"pull_requests": {
			"comment_on_pull_request": {"alt+c"},
		},
	})
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
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, Alt+C: comment, Alt+A: action")
}

func TestStatusLineKeyHints_GivenBrowserChangesInlineComment_WhenRendering_ThenItShowsTheResolutionHint(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")
	then_noError(t, subject.afterStateChange(gui))

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, R: resolve, a: action")
}

func TestStatusLineKeyHints_GivenResolvedBrowserChangesInlineComment_WhenRendering_ThenItShowsTheUnresolveHint(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithResolvedInlineThreadForChangesResolutionTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithResolvedInlineThreadForChangesResolutionTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:43")
	then_noError(t, subject.afterStateChange(gui))

	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, R: unresolve, a: action")
}

func then_footerTextIs(t *testing.T, gui *gocui.Gui, viewName string, expected string) {
	t.Helper()

	footerView, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != expected {
		t.Fatalf("expected footer %q for view %q, actual %q", expected, viewName, actual)
	}
}
