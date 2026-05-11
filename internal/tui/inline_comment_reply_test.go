package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenBrowserConversationsCursorOnInlineComment_WhenOpening_ThenItShowsReplyToInlineComment(t *testing.T) {
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

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", pullRequestInlineCommentReplyEditorTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserConversationsCursorOnPullRequestComment_WhenOpening_ThenItHidesReplyToInlineComment(t *testing.T) {
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered general feedback")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected the popup to hide %q, actual %q", pullRequestInlineCommentReplyEditorTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserChangesCursorOnInlineComment_WhenOpening_ThenItShowsReplyToInlineComment(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected the popup to contain %q, actual %q", pullRequestInlineCommentReplyEditorTitle, popupView.Buffer())
	}
}

func TestActionsPopup_GivenReviewModeCursorOutsideInlineComment_WhenOpening_ThenItHidesReplyToInlineComment(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs:         map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if strings.Contains(popupView.Buffer(), pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected the popup to hide %q, actual %q", pullRequestInlineCommentReplyEditorTitle, popupView.Buffer())
	}
}

func TestReplyToInlineComment_GivenBrowserConversationsAction_WhenSubmitting_ThenItRefreshesTheThreadAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"General feedback":   "Rendered general feedback",
		"Inline thread body": "Rendered inline thread body",
		"Browser reply body": "Rendered browser reply body",
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

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	editorView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(editorView.Title, pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected editor title to contain %q, actual %q", pullRequestInlineCommentReplyEditorTitle, editorView.Title)
	}
	if actual := subject.modalEditor.Text(); actual != "" {
		t.Fatalf("expected an empty inline comment reply draft, actual %q", actual)
	}

	subject.modalEditor.editor.SetText("Browser reply body")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.reviewThreadReplyReviewIDs, []string{""}) {
		t.Fatalf("expected reply review ids %v, actual %v", []string{""}, loader.reviewThreadReplyReviewIDs)
	}
	if !reflect.DeepEqual(loader.reviewThreadReplyThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected reply thread ids %v, actual %v", []string{"thread-1"}, loader.reviewThreadReplyThreadIDs)
	}
	if !reflect.DeepEqual(loader.reviewThreadReplyBodies, []string{"Browser reply body"}) {
		t.Fatalf("expected reply bodies %v, actual %v", []string{"Browser reply body"}, loader.reviewThreadReplyBodies)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered browser reply body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered browser reply body", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestInlineCommentReplySuccessMessage)
}

func TestReplyToInlineComment_GivenBrowserChangesAction_WhenSubmitting_ThenItRefreshesTheThreadAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body":    "Rendered inline thread body",
		"Browser changes reply": "Rendered browser changes reply",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Browser changes reply")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.reviewThreadReplyReviewIDs, []string{""}) {
		t.Fatalf("expected reply review ids %v, actual %v", []string{""}, loader.reviewThreadReplyReviewIDs)
	}
	if !reflect.DeepEqual(loader.reviewThreadReplyThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected reply thread ids %v, actual %v", []string{"thread-1"}, loader.reviewThreadReplyThreadIDs)
	}
	if !reflect.DeepEqual(loader.reviewThreadReplyBodies, []string{"Browser changes reply"}) {
		t.Fatalf("expected reply bodies %v, actual %v", []string{"Browser changes reply"}, loader.reviewThreadReplyBodies)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered browser changes reply") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered browser changes reply", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestInlineCommentReplySuccessMessage)
}

func TestReplyToInlineCommentShortcut_GivenBrowserChangesCursorOnInlineComment_WhenPressingR_ThenItOpensTheReplyComposer(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'r')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Title, pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected composer title to contain %q, actual %q", pullRequestInlineCommentReplyEditorTitle, composerView.Title)
	}
}

func TestReplyToInlineCommentShortcut_GivenBrowserChangesCursorOutsideInlineComment_WhenPressingR_ThenItShowsAStatusMessageAndDoesNothing(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline thread body": "Rendered inline thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	given_browserChangesDetailFocusForInlineComment(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "new line")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'r')
	actualErr = actualHandler(gui, detailView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)
	then_viewDoesNotExist(t, gui, viewModalEditorName)
	then_statusLineContains(t, gui, inlineCommentReplyUnavailableMessage)
}

func TestReplyToInlineComment_GivenReviewModeAction_WhenSubmitting_ThenItAddsTheReplyToThePendingReviewThread(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs:         map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body": "Rendered inline thread body",
		"Review reply body":  "Rendered review reply body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Review reply body")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if !reflect.DeepEqual(loader.reviewThreadReplyReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected reply review ids %v, actual %v", []string{"PRR_pending"}, loader.reviewThreadReplyReviewIDs)
	}
	if !reflect.DeepEqual(loader.reviewThreadReplyThreadIDs, []string{"thread-1"}) {
		t.Fatalf("expected reply thread ids %v, actual %v", []string{"thread-1"}, loader.reviewThreadReplyThreadIDs)
	}
	if !reflect.DeepEqual(loader.reviewThreadReplyBodies, []string{"Review reply body"}) {
		t.Fatalf("expected reply bodies %v, actual %v", []string{"Review reply body"}, loader.reviewThreadReplyBodies)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.diffCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered review reply body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered review reply body", detailView.Buffer())
	}
	then_statusLineContains(t, gui, pullRequestInlineCommentReplySuccessMessage)
}

func TestReplyToInlineComment_GivenCommentsTabSubmit_WhenPostingReply_ThenItKeepsTheRenderedThreadVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"General feedback":   "Rendered general feedback",
		"Inline thread body": "Rendered inline thread body",
		"Optimistic reply":   "Rendered optimistic reply",
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

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Optimistic reply")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued run, actual %v", loader.detailCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered optimistic reply") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered optimistic reply", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (3)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 1)
}

func TestReplyToInlineComment_GivenChangesTabSubmit_WhenPostingReply_ThenItKeepsTheRenderedDiffVisibleWhileQueueingBackgroundRefreshes(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithInlineThreadForReplyTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body": "Rendered inline thread body",
		"Optimistic reply":   "Rendered optimistic reply",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Optimistic reply")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected two queued background refreshes, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued runs, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh call before the queued runs, actual %v", loader.diffCalls)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered optimistic reply") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered optimistic reply", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (3)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 3)
}

func TestReplyToInlineComment_GivenReviewModeSubmit_WhenPostingReply_ThenItKeepsTheRenderedDiffVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs:         map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionDiffWithInlineThreadForReplyTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body": "Rendered inline thread body",
		"Optimistic reply":   "Rendered optimistic reply",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)
	given_reviewModeDetailFocusForActions(t, gui, subject)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Optimistic reply")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewDetailName)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh call before the queued run, actual %v", loader.diffCalls)
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if strings.Contains(filesView.Buffer(), "Loading file tree...") {
		t.Fatalf("expected files buffer to avoid the loading state, actual %q", filesView.Buffer())
	}
	if !strings.Contains(filesView.Buffer(), "render.go  2") {
		t.Fatalf("expected files buffer to show the optimistic comment count, actual %q", filesView.Buffer())
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered optimistic reply") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered optimistic reply", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Loading pull request diff...") {
		t.Fatalf("expected detail buffer to avoid the diff loading state, actual %q", detailView.Buffer())
	}
}

func given_pullRequestDetailWithInlineThreadForReplyTests() githubcli.PullRequestDetail {
	return githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
		Body:        "Body 42",
		BaseRefName: "main",
		HeadRefName: "feature/comments",
		State:       "OPEN",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "General feedback",
			CreatedAt: "2026-04-18T10:00:00Z",
		}},
		InlineCommentThreads: []githubcli.PullRequestReviewThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     43,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				ID:              "PRRC_1",
				ViewerDidAuthor: false,
				Author:          &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
				Body:            "Inline thread body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
			}},
		}},
	}
}

func given_reviewSessionDiffWithInlineThreadForReplyTests() githubcli.PullRequestDiff {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: false,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:            "Inline thread body",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	return diff
}
