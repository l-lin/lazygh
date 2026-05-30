package tui

import (
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestModalEditorLoading_GivenQueuedPullRequestTitleEditSubmit_WhenSubmitting_ThenItShowsTheGhCommandSpinnerInTheStatusLine(t *testing.T) {
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestTitleEditorTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestTitleEditorTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.overlayState.modalEditor.lineEditor.SetText("Renamed PR")
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	titleView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyEnter)
	actualErr = actualHandler(gui, titleView)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if actual := len(asyncRunner.runs); actual != 1 {
		t.Fatalf("expected one queued async submit, actual %d", actual)
	}

	expectedLoading := formatRunningCommandStatus(formatStatusLineCommand("gh", "pr", "edit", "42", "-R", "acme/widgets", "--title", "Renamed PR"))
	if actual := subject.statusLinePresenter().Text(); actual != subject.loadingSpinnerStatus(expectedLoading) {
		t.Fatalf("expected status line %q, actual %q", subject.loadingSpinnerStatus(expectedLoading), actual)
	}
	then_statusLineContains(t, gui, expectedLoading)
}

func TestModalEditorLoading_GivenQueuedInlineCommentReplySubmit_WhenSubmitting_ThenItShowsTheGhCommandSpinnerInTheStatusLine(t *testing.T) {
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline thread body")
	actualErr := subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.overlayState.modalEditor.editor.SetText("Review reply body")
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if actual := len(asyncRunner.runs); actual != 1 {
		t.Fatalf("expected one queued async submit, actual %d", actual)
	}

	expectedLoading := formatRunningCommandStatus(formatStatusLineCommand("gh", "api", "graphql"))
	if actual := subject.statusLinePresenter().Text(); actual != subject.loadingSpinnerStatus(expectedLoading) {
		t.Fatalf("expected status line %q, actual %q", subject.loadingSpinnerStatus(expectedLoading), actual)
	}
	then_statusLineContains(t, gui, expectedLoading)
}

func TestModalEditorLoading_GivenQueuedReviewInlineCommentSubmit_WhenSubmitting_ThenItShowsTheGhCommandSpinnerInTheStatusLine(t *testing.T) {
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
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.overlayState.modalEditor.editor.SetText("Please add context")
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualHandler = given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if actual := len(asyncRunner.runs); actual != 1 {
		t.Fatalf("expected one queued async submit, actual %d", actual)
	}

	expectedLoading := formatRunningCommandStatus(formatStatusLineCommand("gh", "api", "graphql"))
	if actual := subject.statusLinePresenter().Text(); actual != subject.loadingSpinnerStatus(expectedLoading) {
		t.Fatalf("expected status line %q, actual %q", subject.loadingSpinnerStatus(expectedLoading), actual)
	}
	then_statusLineContains(t, gui, expectedLoading)
}
