package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"github.com/jesseduffield/gocui"
)

func TestReviewMode_GivenTheFilesPane_WhenPressingDoubleBracketMotions_ThenItMovesToThePreviousOrNextFile(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
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

	nextFileHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, ']')
	previousFileHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, '[')
	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)

	actualErr = nextFileHandler(gui, filesView)
	then_noError(t, actualErr)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")

	actualErr = nextFileHandler(gui, filesView)
	then_noError(t, actualErr)
	then_selectedReviewFileIs(t, subject, "internal/tui/model.go")

	actualErr = previousFileHandler(gui, filesView)
	then_noError(t, actualErr)
	then_selectedReviewFileIs(t, subject, "internal/tui/model.go")

	actualErr = previousFileHandler(gui, filesView)
	then_noError(t, actualErr)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
}

func TestReviewMode_GivenTheDiffPane_WhenPressingDoubleBracketMotions_ThenItMovesToThePreviousOrNextFileWithoutLeavingViewZero(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
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

	nextFileHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, ']')
	previousFileHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '[')
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	actualErr = nextFileHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")

	actualErr = nextFileHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/model.go")
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "internal/tui/model.go") || !strings.Contains(detailView.Buffer(), "+new model") {
		t.Fatalf("expected detail view to show the next file diff, actual %q", detailView.Buffer())
	}

	actualErr = previousFileHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/model.go")

	actualErr = previousFileHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "internal/tui/render.go") || !strings.Contains(detailView.Buffer(), "+another line") {
		t.Fatalf("expected detail view to show the previous file diff, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenTheFilesPane_WhenPressingCommentMotions_ThenItMovesToThePreviousOrNextCommentWithoutOpeningTheComposer(t *testing.T) {
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

	nextPrefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, ']')
	previousPrefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, '[')
	commentHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'c')
	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)

	actualErr = nextPrefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = commentHandler(gui, filesView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go:3 R3 Unresolved")
	then_viewDoesNotExist(t, gui, viewModalEditorName)

	actualErr = nextPrefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = commentHandler(gui, filesView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_selectedReviewFileIs(t, subject, "internal/tui/model.go")
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go:10 L10 Unresolved")
	then_viewDoesNotExist(t, gui, viewModalEditorName)

	actualErr = previousPrefixHandler(gui, filesView)
	then_noError(t, actualErr)
	actualErr = commentHandler(gui, filesView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go:3 R3 Unresolved")
	then_viewDoesNotExist(t, gui, viewModalEditorName)
}

func TestReviewMode_GivenTheDiffPane_WhenPressingCommentMotions_ThenItMovesToThePreviousOrNextCommentWithoutLeavingViewZero(t *testing.T) {
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

	nextPrefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, ']')
	previousPrefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, '[')
	commentHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)

	actualErr = nextPrefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = commentHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go:3 R3 Unresolved")
	then_viewDoesNotExist(t, gui, viewModalEditorName)

	actualErr = nextPrefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = commentHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/model.go")
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go:10 L10 Unresolved")
	then_viewDoesNotExist(t, gui, viewModalEditorName)

	actualErr = previousPrefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = commentHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_selectedReviewFileIs(t, subject, "internal/tui/render.go")
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go:3 R3 Unresolved")
	then_viewDoesNotExist(t, gui, viewModalEditorName)
}

func then_selectedReviewFileIs(t *testing.T, subject *Program, expectedPath string) {
	t.Helper()

	selectedFile, ok := subject.selectedReviewSessionDiffFile()
	if !ok {
		t.Fatal("expected a selected review file")
	}
	if selectedFile.Path != expectedPath {
		t.Fatalf("expected selected review file %q, actual %q", expectedPath, selectedFile.Path)
	}
}

func then_reviewModeDetailCursorLineContains(t *testing.T, gui *gocui.Gui, subject *Program, expectedSegment string) {
	t.Helper()

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	subject.syncDetailViewState(document, detailView.InnerHeight())
	if subject.detailViewState.cursor.line < 0 || subject.detailViewState.cursor.line >= len(document.lines) {
		t.Fatalf("expected detail cursor line within bounds, actual %d", subject.detailViewState.cursor.line)
	}

	actualLine := string(document.lines[subject.detailViewState.cursor.line])
	if !strings.Contains(actualLine, expectedSegment) {
		t.Fatalf("expected detail cursor line to contain %q, actual %q", expectedSegment, actualLine)
	}
}

func given_reviewSessionPullRequestDiffWithComments() githubcli.PullRequestDiff {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{
		{
			ID:       "thread-render",
			Path:     "internal/tui/render.go",
			Line:     3,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:      "Render thread",
				CreatedAt: "2026-04-24T10:00:00Z",
			}},
		},
		{
			ID:           "thread-model",
			Path:         "internal/tui/model.go",
			OriginalLine: 10,
			DiffSide:     "LEFT",
			Comments: []githubcli.PullRequestComment{{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"},
				Body:      "Model thread",
				CreatedAt: "2026-04-24T11:00:00Z",
			}},
		},
	}
	return diff
}
