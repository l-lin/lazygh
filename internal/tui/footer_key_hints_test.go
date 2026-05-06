package tui

import (
	"strings"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"github.com/jesseduffield/gocui"
)

func TestPaneFooter_GivenActivePullRequestsView_WhenRendering_ThenItShowsDefaultKeyHints(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_footerTextIs(t, gui, viewPullRequestsFooterName, "? Help  / Search  a Actions")
}

func TestPaneFooter_GivenConfiguredKeyOverrides_WhenRendering_ThenItUsesTheResolvedKeys(t *testing.T) {
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

	then_footerTextIs(t, gui, viewPullRequestsFooterName, "! Help  s/<c-s> Search  p Actions")
}

func TestPaneFooter_GivenAContextWithoutActions_WhenRendering_ThenItOmitsTheActionsHint(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	then_footerTextIs(t, gui, viewUserFooterName, "? Help  / Search")
}

func TestPaneFooter_GivenScopedActionOverrides_WhenRenderingBrowserModePullRequestsAndDetail_ThenEachPaneUsesItsOwnActionScope(t *testing.T) {
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
	then_footerTextIs(t, gui, viewPullRequestsFooterName, "? Help  / Search  p Actions")

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	then_footerTextIs(t, gui, viewDetailFooterName, "? Help  / Search  d Actions")
}

func TestPaneFooter_GivenScopedActionOverrides_WhenRenderingReviewModeFilesAndDiff_ThenEachPaneUsesItsOwnActionScope(t *testing.T) {
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
	then_footerTextIs(t, gui, viewPullRequestsFooterName, "? Help  / Search  p Actions")

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	then_footerTextIs(t, gui, viewDetailFooterName, "? Help  / Search  d Actions")
}

func then_footerTextIs(t *testing.T, gui *gocui.Gui, viewName string, expected string) {
	t.Helper()

	footerView, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	if actual := strings.TrimSpace(footerView.Buffer()); actual != expected {
		t.Fatalf("expected footer %q for view %q, actual %q", expected, viewName, actual)
	}
}
