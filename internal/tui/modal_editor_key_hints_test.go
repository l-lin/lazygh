package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestStatusLineKeyHints_GivenPullRequestTitleEditor_WhenRendering_ThenItShowsTheStandardGreySubmitHintAndNoModalFooter(t *testing.T) {
	subject := NewProgramWithModel(given_pullRequestCommentModel())
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("rename", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "rename"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if modalView.Footer != "" {
		t.Fatalf("expected the modal footer to stay empty, actual %q", modalView.Footer)
	}
	then_statusLineKeyHintsAre(t, gui, "Alt+Enter: Submit")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Alt+Enter: Submit")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Alt+Enter: Submit", given_themeColorHex(t, theme.InactiveTitleHex), "title editor key hints")
}

func TestStatusLineKeyHints_GivenPullRequestDescriptionEditor_WhenRendering_ThenItShowsTheStandardGreySubmitHintAndNoModalFooter(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Rich body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("body", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "body"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if modalView.Footer != "" {
		t.Fatalf("expected the modal footer to stay empty, actual %q", modalView.Footer)
	}
	then_statusLineKeyHintsAre(t, gui, "Alt+Enter: Submit")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Alt+Enter: Submit")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Alt+Enter: Submit", given_themeColorHex(t, theme.InactiveTitleHex), "description editor key hints")
}

func TestStatusLineKeyHints_GivenInlineCommentReplyEditor_WhenRendering_ThenItShowsTheStandardGreySubmitHintAndNoModalFooter(t *testing.T) {
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
	subject.model.UpdateActionsPopupSearch("reply to inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "reply to inline comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	modalView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(modalView.Title, pullRequestInlineCommentReplyEditorTitle) {
		t.Fatalf("expected modal title to contain %q, actual %q", pullRequestInlineCommentReplyEditorTitle, modalView.Title)
	}
	if modalView.Footer != "" {
		t.Fatalf("expected the modal footer to stay empty, actual %q", modalView.Footer)
	}
	then_statusLineKeyHintsAre(t, gui, "Alt+Enter: Submit")
	then_statusLineKeyHintsAreRightAligned(t, gui, "Alt+Enter: Submit")
	then_viewLineSegmentHasForegroundColor(t, gui, viewStatusLineKeyHintsName, 0, "Alt+Enter: Submit", given_themeColorHex(t, theme.InactiveTitleHex), "inline comment reply key hints")
}
