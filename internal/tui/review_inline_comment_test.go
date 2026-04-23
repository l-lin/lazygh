package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestReviewMode_GivenTheDetailCursorOnADiffLine_WhenOpeningTheInlineCommentComposer_ThenItShowsATenLinePopupAndReusesTheExternalEditorShortcut(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.externalEditor = &fakeExternalEditor{editedText: "Edited in $EDITOR"}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Title, pullRequestReviewInlineCommentComposerTitle) {
		t.Fatalf("expected composer title to contain %q, actual %q", pullRequestReviewInlineCommentComposerTitle, composerView.Title)
	}
	if strings.Contains(composerView.Title, pullRequestCommentComposerTitle) {
		t.Fatalf("expected review mode detail comment shortcut to avoid the PR comment composer, actual %q", composerView.Title)
	}
	_, y0, _, y1, actualErr := gui.ViewPosition(viewModalEditorName)
	then_noError(t, actualErr)
	if actual := y1 - y0 + 1; actual != reviewInlineCommentModalHeight {
		t.Fatalf("expected inline comment composer height %d, actual %d", reviewInlineCommentModalHeight, actual)
	}

	subject.modalEditor.editor.SetText("Draft inline comment")
	actualHandled := subject.editModalEditor(composerView, gocui.KeyCtrlG, 0, gocui.ModNone)
	if !actualHandled {
		t.Fatal("expected ctrl-g to be handled")
	}
	if subject.externalEditor.(*fakeExternalEditor).receivedText != "Draft inline comment" {
		t.Fatalf("expected external editor input %q, actual %q", "Draft inline comment", subject.externalEditor.(*fakeExternalEditor).receivedText)
	}
	if !strings.Contains(composerView.Buffer(), "Edited in $EDITOR") {
		t.Fatalf("expected composer buffer to contain %q, actual %q", "Edited in $EDITOR", composerView.Buffer())
	}
}

func TestReviewMode_GivenAnInlineCommentSubmit_WhenItSucceeds_ThenItReloadsTheDiffAndShowsTheNewThreadInViewZero(t *testing.T) {
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Please add context")

	actualHandler = given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.reviewThreadReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected review thread review ids %v, actual %v", []string{"PRR_pending"}, loader.reviewThreadReviewIDs)
	}
	expectedTargets := []githubcli.PullRequestReviewThreadTarget{{Path: "internal/tui/render.go", Line: 2, Side: "RIGHT", SubjectType: "LINE"}}
	if !reflect.DeepEqual(loader.reviewThreadTargets, expectedTargets) {
		t.Fatalf("expected review thread targets %+v, actual %+v", expectedTargets, loader.reviewThreadTargets)
	}
	if !reflect.DeepEqual(loader.reviewThreadBodies, []string{"Please add context"}) {
		t.Fatalf("expected review thread bodies %v, actual %v", []string{"Please add context"}, loader.reviewThreadBodies)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Please add context") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Please add context", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestReviewInlineCommentSuccessMessage)
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestReviewMode_GivenTheDetailCursorOnAnInvalidRow_WhenOpeningTheInlineCommentComposer_ThenItShowsAnErrorAndKeepsFocusOnViewZero(t *testing.T) {
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewModalEditorName)

	then_statusLineContains(t, gui, reviewThreadTargetUnavailableMessage)
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestReviewMode_GivenGitHubRejectsTheInlineComment_WhenSubmitting_ThenItKeepsTheDraftVisibleAndShowsTheError(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID:   "PRR_pending",
		reviewThreadErr: errors.New("line must be part of the diff"),
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Draft inline comment")

	actualHandler = given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Buffer(), "Draft inline comment") {
		t.Fatalf("expected composer buffer to contain %q, actual %q", "Draft inline comment", composerView.Buffer())
	}
	if !strings.Contains(composerView.Title, "line must be part of the diff") {
		t.Fatalf("expected composer title to contain %q, actual %q", "line must be part of the diff", composerView.Title)
	}
}

func given_reviewModeDetailCursorOnLineContaining(t *testing.T, gui *gocui.Gui, subject *Program, segment string) {
	t.Helper()

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	subject.syncDetailViewState(document, detailView.InnerHeight())
	lineIndex, _ := given_detailDocumentLineContaining(t, document, segment)
	subject.detailViewState.cursor = detailPosition{line: lineIndex, column: 0}
	subject.detailViewState.preferredColumn = 0
	subject.detailViewState.sync(document, detailView.InnerHeight())
	actualErr = subject.refreshDetailView(gui)
	then_noError(t, actualErr)
}
