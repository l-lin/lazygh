package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
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
