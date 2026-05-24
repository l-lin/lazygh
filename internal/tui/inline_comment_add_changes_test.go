package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenBrowserChangesTabCursorOnAValidDiffLine_WhenOpening_ThenItShowsAddInlineComment(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Add inline comment") {
		t.Fatalf("expected popup to contain %q, actual %q", "Add inline comment", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserChangesTabCursorOnAnInlineComment_WhenOpening_ThenItHidesAddInlineComment(t *testing.T) {
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

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), "Add inline comment") {
		t.Fatalf("expected popup to hide %q, actual %q", "Add inline comment", popupView.Buffer())
	}
}

func TestPullRequestCommentShortcut_GivenBrowserChangesDiffLine_WhenPressingC_ThenItOpensTheInlineCommentComposer(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
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
		t.Fatalf("expected changes-tab comment shortcut to avoid the PR comment composer, actual %q", composerView.Title)
	}
	if actual := subject.modalEditor.Text(); actual != "" {
		t.Fatalf("expected an empty inline comment draft for a single diff line, actual %q", actual)
	}
}

func TestPullRequestCommentShortcut_GivenBrowserChangesLinewiseSelectionAcrossAddedLines_WhenPressingC_ThenItPrefillsASuggestionFence(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionJavaSuggestionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeLinewiseSelectionBetweenLinesContaining(t, gui, subject, "System.out.println(\"This is an example\");", "return format(version);")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	actualErr = actualHandler(gui, detailView)
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

func TestInlineComment_GivenBrowserChangesSubmitWithoutPendingReview_WhenPosting_ThenItStartsOrReusesAPendingReviewAndKeepsTheRenderedDiffVisibleWhileQueueingBackgroundRefreshes(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailForChangesInlineCommentTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Optimistic inline comment": "Rendered optimistic inline comment",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'c')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Optimistic inline comment")
	actualHandler = given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.startReviewCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected start review calls %v, actual %v", []string{"acme/widgets#42"}, loader.startReviewCalls)
	}
	if !reflect.DeepEqual(loader.reviewThreadReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected review thread review ids %v, actual %v", []string{"PRR_pending"}, loader.reviewThreadReviewIDs)
	}
	expectedTargets := []githubcli.PullRequestReviewThreadTarget{{Path: "internal/tui/render.go", Line: 2, Side: "RIGHT", SubjectType: "LINE"}}
	if !reflect.DeepEqual(loader.reviewThreadTargets, expectedTargets) {
		t.Fatalf("expected review thread targets %+v, actual %+v", expectedTargets, loader.reviewThreadTargets)
	}
	if !reflect.DeepEqual(loader.reviewThreadBodies, []string{"Optimistic inline comment"}) {
		t.Fatalf("expected review thread bodies %v, actual %v", []string{"Optimistic inline comment"}, loader.reviewThreadBodies)
	}
	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected two queued background refreshes, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued runs, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh call before the queued runs, actual %v", loader.diffCalls)
	}

	pendingState, ok := subject.pendingPullRequestReviewCache["acme/widgets#42"]
	if !ok {
		t.Fatal("expected the pending review state to be cached after creating an inline comment")
	}
	if pendingState.id != "PRR_pending" {
		t.Fatalf("expected cached pending review id %q, actual %q", "PRR_pending", pendingState.id)
	}

	detailView, actualErr = gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered optimistic inline comment") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered optimistic inline comment", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (1)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 3)
	then_statusLineContains(t, gui, pullRequestReviewInlineCommentSuccessMessage)
}

func given_browserChangesDetailFocusForInlineComment(t *testing.T, gui *gocui.Gui, subject *Program) {
	t.Helper()

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
}

func given_pullRequestDetailForChangesInlineCommentTests() githubcli.PullRequestDetail {
	return githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
		Body:        "Body 42",
		BaseRefName: "main",
		HeadRefName: "feature/comments",
		State:       "OPEN",
	}
}
