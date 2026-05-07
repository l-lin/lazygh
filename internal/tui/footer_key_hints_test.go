package tui

import (
	"strings"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
	"github.com/jesseduffield/gocui"
)

func TestStatusLineKeyHints_GivenActivePullRequestsView_WhenRendering_ThenItShowsDarkGreyCommaSeparatedHintsRightAlignedOnTheBottomRow(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, a: Action")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "?: Help, /: Search, a: Action", given_themeColorHex(t, theme.InactiveTitleHex), "status line key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "?: Help, /: Search, a: Action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func TestStatusLineKeyHints_GivenConfiguredKeyOverrides_WhenRendering_ThenItUsesTheResolvedKeysInKeyColonDescriptionFormat(t *testing.T) {
	subject := given_programWithKeymapOverrides(given_pullRequestCommentModel(), appconfig.KeymapOverrides{
		"main": {
			"toggle_help": {"!"},
			"open_search": {"s", "ctrl+s"},
		},
		"pull_requests": {
			"open_actions_popup": {"p"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_statusLineKeyHintsAre(t, gui, "!: Help, s/<c-s>: Search, p: Action")
}

func TestStatusLineKeyHints_GivenAssigneePickerVisible_WhenRendering_ThenItShowsDarkGreyPopupHintsRightAlignedOnTheBottomRow(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), given_pullRequestAssigneeLoader())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)

	then_statusLineKeyHintsAre(t, gui, "/: Search, Enter: Toggle, Alt+Enter: Submit")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "/: Search, Enter: Toggle, Alt+Enter: Submit", given_themeColorHex(t, theme.InactiveTitleHex), "assignee picker key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "/: Search, Enter: Toggle, Alt+Enter: Submit")
}

func TestPaneFooter_GivenFocusedViewOneWithoutSearchSummary_WhenRendering_ThenItShowsNoPaneFooterOrKeyHints(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_viewDoesNotExist(t, gui, viewUserFooterName)
	then_viewDoesNotExist(t, gui, viewStatusLineKeyHintsName)
	then_statusLineIs(t, gui, "")
}

func TestPaneFooter_GivenFocusedViewOneWithASearchSummary_WhenRendering_ThenItShowsOnlyTheSearchSummary(t *testing.T) {
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
	then_viewDoesNotExist(t, gui, viewStatusLineKeyHintsName)
	then_statusLineIs(t, gui, "")
}

func TestStatusLineKeyHints_GivenScopedActionOverrides_WhenRenderingBrowserModePullRequestsAndDetail_ThenEachPaneUsesItsOwnActionScope(t *testing.T) {
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
	subject := NewProgramWithModelAndLoader(model, loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"pull_requests": {
			"open_actions_popup": {"p"},
		},
		"detail": {
			"open_actions_popup": {"d"},
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
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, p: Action")

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, d: Action")
}

func TestStatusLineKeyHints_GivenScopedActionOverrides_WhenRenderingReviewModeFilesAndDiff_ThenEachPaneUsesItsOwnActionScope(t *testing.T) {
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
		"pull_requests": {
			"open_actions_popup": {"p"},
		},
		"detail": {
			"open_actions_popup": {"d"},
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
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, p: Action")

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, d: Action")
}

func then_footerTextIs(t *testing.T, gui *gocui.Gui, viewName string, expected string) {
	t.Helper()

	footerView, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != expected {
		t.Fatalf("expected footer %q for view %q, actual %q", expected, viewName, actual)
	}
}
