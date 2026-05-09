package tui

import (
	"strings"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestHelpPopup_GivenConfiguredKeyOverrides_WhenTogglingHelp_ThenItShowsTheConfiguredKeys(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_programWithKeymapOverrides(model, appconfig.KeymapOverrides{
		"global": {
			"quit": {"ctrl+x"},
		},
		"main": {
			"open_search": {"s"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Search pull requests", "s")
	then_helpEntryUsesKey(t, actualBuffer, "Quit", "<c-x>")
}

func TestHelpPopup_GivenDetailFocus_WhenTogglingHelp_ThenItShowsViewportPlacementMotionsAndHalfPageRecentering(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Cursor to top/center/bottom", "zt/zz/zb")
	then_helpEntryUsesKey(t, actualBuffer, "Half-page down + recenter", "<c-d>")
	then_helpEntryUsesKey(t, actualBuffer, "Half-page up + recenter", "<c-u>")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page down", "<c-f>/pagedown")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page up", "<c-b>/pageup")
}

func TestHelpPopup_GivenDetailFocus_WhenTogglingHelp_ThenItShowsGXForOpeningTheLinkUnderCursor(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Open link under cursor", "gx")
}

func TestHelpPopup_GivenPullRequestDetailFocus_WhenTogglingHelp_ThenItShowsZMAndZRForBulkFolds(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Body 42", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Close/open all folds", "zM/zR")
}

func TestHelpPopup_GivenUserFocus_WhenTogglingHelp_ThenItShowsViewportPlacementMotionsAndHalfPageRecentering(t *testing.T) {
	subject := NewProgramWithModel(given_model())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Selection to top/center/bottom", "zt/zz/zb")
	then_helpEntryUsesKey(t, actualBuffer, "Half-page down + recenter", "<c-d>")
	then_helpEntryUsesKey(t, actualBuffer, "Half-page up + recenter", "<c-u>")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page down", "<c-f>/pagedown")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page up", "<c-b>/pageup")
}

func TestHelpPopup_GivenReviewFilesFocus_WhenTogglingHelp_ThenItShowsReviewFileAndCommentMotions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiffWithComments(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Previous/next file", "[[/]]")
	then_helpEntryUsesKey(t, actualBuffer, "Previous/next comment", "[c/]c")
	then_helpEntryUsesKey(t, actualBuffer, "Expand/collapse fold", "za")
	then_helpEntryUsesKey(t, actualBuffer, "Close/open all folds", "zM/zR")
}

func TestHelpPopup_GivenReviewDiffFocus_WhenTogglingHelp_ThenItShowsReviewFileAndCommentMotions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiffWithComments(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Previous/next file", "[[/]]")
	then_helpEntryUsesKey(t, actualBuffer, "Previous/next comment", "[c/]c")
}

func then_helpEntryUsesKey(t *testing.T, buffer string, description string, expectedKey string) {
	t.Helper()

	for _, line := range strings.Split(buffer, "\n") {
		if !strings.HasSuffix(strings.TrimSpace(line), description) {
			continue
		}

		actualKey := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), description))
		if actualKey != expectedKey {
			t.Fatalf("expected help key %q for %q, actual %q", expectedKey, description, actualKey)
		}
		return
	}

	t.Fatalf("expected help entry for %q in %q", description, buffer)
}
