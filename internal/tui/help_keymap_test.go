package tui

import (
	"strings"
	"testing"

	appconfig "github.com/l-lin/lazygh/internal/config"
	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestHelpPopup_GivenConfiguredKeyOverrides_WhenTogglingHelp_ThenItShowsTheConfiguredKeys(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_programWithKeymapOverrides(model, appconfig.KeymapOverrides{
		"global": {
			"quit":        {"ctrl+x"},
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
	then_helpEntryUsesKey(t, actualBuffer, "Quit", "Ctrl+X")
}

func TestHelpPopup_GivenPullRequestsFocus_WhenTogglingHelp_ThenItShowsTheCustomSearchOpenPRFromURLAndCopyPRURLShortcuts(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
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
	then_helpEntryUsesKey(t, helpView.Buffer(), "Custom search", ":")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Open PR from clipboard", "Ctrl+V")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Copy PR URL", "Alt+Y")
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
	then_helpEntryUsesKey(t, actualBuffer, "Half-page down + recenter", "Ctrl+D")
	then_helpEntryUsesKey(t, actualBuffer, "Half-page up + recenter", "Ctrl+U")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page down", "Ctrl+F/PageDown")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page up", "Ctrl+B/PageUp")
}

func TestHelpPopup_GivenDetailFocus_WhenTogglingHelp_ThenItShowsYankLinkClipboardAndSearchWordBindings(t *testing.T) {
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
	then_helpEntryUsesKey(t, helpView.Buffer(), "Start yank with motion", "y")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Copy PR URL", "Alt+Y")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Open link under cursor", "gx")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Open PR from clipboard", "Ctrl+V")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Search word under cursor", "*/#")
}

func TestHelpPopup_GivenDetailFocus_WhenTogglingHelp_ThenItShowsCharacterMotionBindings(t *testing.T) {
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
	then_helpEntryUsesKey(t, helpView.Buffer(), "Find/till character", "f/F/t/T")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Repeat character motion", ";/,")
}

func TestHelpPopup_GivenCharacterMotionOverrides_WhenTogglingHelp_ThenItShowsTheConfiguredCharacterMotionBindings(t *testing.T) {
	model := given_model()
	model.OpenDetail()
	subject := given_programWithKeymapOverrides(model, appconfig.KeymapOverrides{
		"cursor": {
			"find_character_forward":           {"s"},
			"find_character_backward":          {"S"},
			"till_character_forward":           {"x"},
			"till_character_backward":          {"X"},
			"repeat_character_motion_forward":  {"r"},
			"repeat_character_motion_backward": {"R"},
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
	then_helpEntryUsesKey(t, helpView.Buffer(), "Find/till character", "s/S/x/X")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Repeat character motion", "r/R")
}

func TestHelpPopup_GivenPullRequestDetailFocus_WhenTogglingHelp_ThenItShowsTheSharedFoldKeys(t *testing.T) {
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
	then_helpEntryUsesKey(t, helpView.Buffer(), "Expand/collapse section", "Enter/za")
	then_helpEntryUsesKey(t, helpView.Buffer(), "Close/open all folds", "zM/zR")
}

func TestHelpPopup_GivenPullRequestChangesDetailFocus_WhenTogglingHelp_ThenItShowsInlineCommentShortcuts(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
	actualErr := subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Add inline comment", "c")
	then_helpEntryUsesKey(t, actualBuffer, "Reply to inline comment", "r")
}

func TestHelpPopup_GivenPullRequestCommentsDetailFocusOnInlineComment_WhenTogglingHelp_ThenItShowsInlineCommentReplyShortcut(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"General feedback":   "Rendered general feedback",
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Reply to inline comment", "r")
}

func TestHelpPopup_GivenReviewDetailFocusOnInlineComment_WhenTogglingHelp_ThenItShowsInlineCommentReplyShortcut(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	then_helpEntryUsesKey(t, helpView.Buffer(), "Reply to inline comment", "r")
}

func TestHelpPopup_GivenReviewFilesFocusAndCustomizedFoldBindings_WhenTogglingHelp_ThenItShowsTheConfiguredSingleAndTwoStepFoldKeys(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiffWithComments(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"folds": {
			"toggle_fold":     {"o"},
			"close_all_folds": {"zX"},
			"open_all_folds":  {"zO"},
		},
	})
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
	then_helpEntryUsesKey(t, actualBuffer, "Expand/collapse fold", "o")
	then_helpEntryUsesKey(t, actualBuffer, "Close/open all folds", "zX/zO")
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
	then_helpEntryUsesKey(t, actualBuffer, "Half-page down + recenter", "Ctrl+D")
	then_helpEntryUsesKey(t, actualBuffer, "Half-page up + recenter", "Ctrl+U")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page down", "Ctrl+F/PageDown")
	then_helpEntryUsesKey(t, actualBuffer, "Full-page up", "Ctrl+B/PageUp")
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
	then_helpEntryUsesKey(t, actualBuffer, "Search word under cursor", "*/#")
}

func TestHelpPopup_GivenReviewFocusAndConfiguredReviewMotionOverrides_WhenTogglingHelp_ThenItShowsTheConfiguredSequences(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiffWithComments(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.ApplyKeymapOverrides(appconfig.KeymapOverrides{
		"review": {
			"previous_file":    {"g["},
			"next_file":        {"g]"},
			"previous_comment": {"gh"},
			"next_comment":     {"gl"},
		},
	})
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
	then_helpEntryUsesKey(t, actualBuffer, "Previous/next file", "g[/g]")
	then_helpEntryUsesKey(t, actualBuffer, "Previous/next comment", "gh/gl")
}

func then_helpEntryUsesKey(t *testing.T, buffer string, description string, expectedKey string) {
	t.Helper()

	for line := range strings.SplitSeq(buffer, "\n") {
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
