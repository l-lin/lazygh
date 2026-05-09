package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenReviewModeDetailCursorOnAValidDiffLine_WhenOpening_ThenItShowsAddInlineComment(t *testing.T) {
	subject := given_reviewModeProgramWithDiffForActions(given_reviewSessionPullRequestDiff(), nil)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Add inline comment") {
		t.Fatalf("expected the popup to contain %q, actual %q", "Add inline comment", popupView.Buffer())
	}
}

func TestActionsPopup_GivenReviewModeDetailCursorOnInvalidTargets_WhenOpening_ThenItHidesAddInlineComment(t *testing.T) {
	testCases := []struct {
		name     string
		diff     githubcli.PullRequestDiff
		renderer MarkdownRenderer
		prepare  func(*testing.T, *gocui.Gui, *Program)
	}{
		{
			name: "file header",
			diff: given_reviewSessionPullRequestDiff(),
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go  +2  -1")
			},
		},
		{
			name: "hunk header",
			diff: given_reviewSessionPullRequestDiff(),
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "@@ -1,2 +1,3 @@")
			},
		},
		{
			name: "inline comment box",
			diff: given_reviewSessionDiffWithInlineThreadForActions(),
			renderer: &fakeMarkdownRenderer{outputs: map[string]string{
				"Thread body": "Rendered thread body",
			}},
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")
			},
		},
		{
			name: "impossible mixed-side range",
			diff: given_reviewSessionPullRequestDiff(),
			prepare: func(t *testing.T, gui *gocui.Gui, subject *Program) {
				given_reviewModeLinewiseSelectionBetweenLinesContaining(t, gui, subject, "old line", "new line")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			subject := given_reviewModeProgramWithDiffForActions(testCase.diff, testCase.renderer)
			gui := given_headlessGui(t)
			defer gui.Close()
			subject.configureGUI(gui)
			given_reviewModeDetailFocusForActions(t, gui, subject)
			testCase.prepare(t, gui, subject)

			actualErr := subject.openActionsPopup(gui, nil)
			then_noError(t, actualErr)

			popupView, actualErr := gui.View(viewActionsPopupName)
			then_noError(t, actualErr)
			if strings.Contains(popupView.Buffer(), "Add inline comment") {
				t.Fatalf("expected the popup to hide %q, actual %q", "Add inline comment", popupView.Buffer())
			}
		})
	}
}

func TestActionsPopup_GivenReviewModeDiffLine_WhenExecutingAddInlineComment_ThenItOpensTheInlineCommentComposerInsteadOfThePRCommentComposer(t *testing.T) {
	subject := given_reviewModeProgramWithDiffForActions(given_reviewSessionPullRequestDiff(), nil)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Title, pullRequestReviewInlineCommentComposerTitle) {
		t.Fatalf("expected composer title to contain %q, actual %q", pullRequestReviewInlineCommentComposerTitle, composerView.Title)
	}
	if strings.Contains(composerView.Title, pullRequestCommentComposerTitle) {
		t.Fatalf("expected popup action to avoid the PR comment composer, actual %q", composerView.Title)
	}
	if actual := subject.modalEditor.Text(); actual != "" {
		t.Fatalf("expected an empty inline comment draft for a single diff line, actual %q", actual)
	}
}

func TestActionsPopup_GivenReviewModeLinewiseSelectionAcrossAddedLines_WhenExecutingAddInlineComment_ThenItPrefillsASuggestionFence(t *testing.T) {
	subject := given_reviewModeProgramWithDiffForActions(given_reviewSessionJavaSuggestionPullRequestDiff(), nil)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeLinewiseSelectionBetweenLinesContaining(t, gui, subject, "System.out.println(\"This is an example\");", "return format(version);")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	expected := strings.Join([]string{
		"```java suggestion",
		"System.out.println(\"This is an example\");",
		"return format(version);",
		"```",
	}, "\n")
	if actual := subject.modalEditor.Text(); actual != expected {
		t.Fatalf("expected inline comment draft %q, actual %q", expected, actual)
	}
}

func TestActionsPopup_GivenSearchForAddInlineComment_WhenReviewCursorMovesToAnInlineCommentBox_ThenItShowsNoMatchingActions(t *testing.T) {
	subject := given_reviewModeProgramWithDiffForActions(given_reviewSessionDiffWithInlineThreadForActions(), &fakeMarkdownRenderer{outputs: map[string]string{
		"Thread body": "Rendered thread body",
	}})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("add inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "add inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Add inline comment") {
		t.Fatalf("expected the popup to contain %q before moving the cursor, actual %q", "Add inline comment", popupView.Buffer())
	}

	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	if strings.Contains(popupView.Buffer(), "Add inline comment") {
		t.Fatalf("expected the popup to hide %q after moving the cursor, actual %q", "Add inline comment", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), "Open PR in browser") {
		t.Fatalf("expected the popup to keep non-matching actions visible after moving the cursor, actual %q", popupView.Buffer())
	}
}

func given_reviewModeProgramWithDiffForActions(diff githubcli.PullRequestDiff, renderer MarkdownRenderer) *Program {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	if renderer != nil {
		subject.markdownRenderer = renderer
	}
	return subject
}

func given_reviewModeDetailFocusForActions(t *testing.T, gui *gocui.Gui, subject *Program) {
	t.Helper()

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
}

func given_reviewSessionDiffWithInlineThreadForActions() githubcli.PullRequestDiff {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	return diff
}
