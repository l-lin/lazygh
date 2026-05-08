package tui

import (
	"strings"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
	"github.com/jesseduffield/gocui"
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

	then_statusLineKeyHintsAre(t, gui, "!: help, s/<c-s>: search, p: action")
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

	then_statusLineKeyHintsAre(t, gui, "/: search, Enter: execute, Escape: cancel")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "/: search, Enter: execute, Escape: cancel", given_themeColorHex(t, theme.InactiveTitleHex), "actions popup key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "/: search, Enter: execute, Escape: cancel")
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

	then_statusLineKeyHintsAre(t, gui, "Enter/Tab: list, Escape: cancel")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Enter/Tab: list, Escape: cancel", given_themeColorHex(t, theme.InactiveTitleHex), "actions popup search key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Enter/Tab: list, Escape: cancel")
}

func TestStatusLineKeyHints_GivenAssigneePickerVisible_WhenRendering_ThenItShowsDarkGreyPopupHintsRightAlignedOnTheBottomRow(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), given_pullRequestAssigneeLoader())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	_ = given_openAssigneePicker(t, gui, subject)

	then_statusLineKeyHintsAre(t, gui, "/: search, Enter: toggle, Alt+Enter: submit, Escape: cancel")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "/: search, Enter: toggle, Alt+Enter: submit, Escape: cancel", given_themeColorHex(t, theme.InactiveTitleHex), "assignee picker key hints")
	then_statusLineKeyHintsAreRightAligned(t, gui, "/: search, Enter: toggle, Alt+Enter: submit, Escape: cancel")
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
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, p: action")

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, d: action")
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
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, p: action")

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, d: action")
}

func then_footerTextIs(t *testing.T, gui *gocui.Gui, viewName string, expected string) {
	t.Helper()

	footerView, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != expected {
		t.Fatalf("expected footer %q for view %q, actual %q", expected, viewName, actual)
	}
}
