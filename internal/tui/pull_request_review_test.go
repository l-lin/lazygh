package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenApproveReviewActionSelected_WhenExecuting_ThenItUsesTheReviewHandlerRefreshesTheDetailAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("lgtm", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "lgtm"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.approveCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected approve calls %v, actual %v", []string{"acme/widgets#42"}, loader.approveCalls)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	pullRequestsFooterView, actualErr := gui.View("pull-requests-footer")
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsFooterView.Buffer(), pullRequestReviewSuccessMessage) {
		t.Fatalf("expected pull requests footer to contain %q, actual %q", pullRequestReviewSuccessMessage, pullRequestsFooterView.Buffer())
	}
}

func TestActionsPopup_GivenReviewCommentActionSelected_WhenExecuting_ThenItOpensTheReviewCommentComposerAndSubmitsThroughTheReviewHandler(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("feedback", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "feedback"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Title, pullRequestReviewCommentComposerTitle) {
		t.Fatalf("expected composer title to contain %q, actual %q", pullRequestReviewCommentComposerTitle, composerView.Title)
	}

	subject.modalEditor.editor.SetText("Please add context")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.reviewCommentCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected review comment calls %v, actual %v", []string{"acme/widgets#42"}, loader.reviewCommentCalls)
	}
	if !reflect.DeepEqual(loader.reviewCommentBodies, []string{"Please add context"}) {
		t.Fatalf("expected review comment bodies %v, actual %v", []string{"Please add context"}, loader.reviewCommentBodies)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)
}

func TestActionsPopup_GivenRequestChangesActionSelected_WhenSubmittingFails_ThenItKeepsTheDraftVisibleAndUsesTheRequestChangesHandler(t *testing.T) {
	loader := &fakePullRequestDetailLoader{requestChangesErr: errors.New("boom")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("block", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "block"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Needs tests")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	if !reflect.DeepEqual(loader.requestChangesCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected request changes calls %v, actual %v", []string{"acme/widgets#42"}, loader.requestChangesCalls)
	}
	if !reflect.DeepEqual(loader.requestChangesBodies, []string{"Needs tests"}) {
		t.Fatalf("expected request changes bodies %v, actual %v", []string{"Needs tests"}, loader.requestChangesBodies)
	}
	composerView, actualErr := gui.View(viewModalEditorName)
	then_noError(t, actualErr)
	if !strings.Contains(composerView.Buffer(), "Needs tests") {
		t.Fatalf("expected composer buffer to contain %q, actual %q", "Needs tests", composerView.Buffer())
	}
	if !strings.Contains(composerView.Title, "boom") {
		t.Fatalf("expected composer title to contain %q, actual %q", "boom", composerView.Title)
	}
}
