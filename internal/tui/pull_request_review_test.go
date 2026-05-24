package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenApproveReviewActionSelected_WhenExecuting_ThenItKeepsThePopupOpenShowsTheStatusLineSpinnerAndDelaysTheGitHubCall(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(loader.details["acme/widgets#42"])}
	subject.pullRequestDiffCache = map[string]pullRequestDiffResult{"acme/widgets#42": {}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestReviewApprovalTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestReviewApprovalTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued approve mutation, actual %d", len(asyncRunner.runs))
	}
	if len(loader.approveCalls) != 0 {
		t.Fatalf("expected the approve call to wait for the queued run, actual %v", loader.approveCalls)
	}
	then_currentViewNameIs(t, gui, viewActionsPopupSearchName)
	_, actualErr = gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	then_statusLineContains(t, gui, string(loadingSpinnerFrames[0]))
	then_statusLineContains(t, gui, formatRunningCommandStatus(approvePullRequestCommand("acme/widgets", 42)))
}

func TestActionsPopup_GivenApproveReviewActionSelected_WhenExecuting_ThenItUsesTheReviewHandlerRefreshesTheDetailAndShowsFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDiffCache = map[string]pullRequestDiffResult{"acme/widgets#42": {}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestReviewApprovalTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestReviewApprovalTitle))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)

	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.approveCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected approve calls %v, actual %v", []string{"acme/widgets#42"}, loader.approveCalls)
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull request diff to be invalidated after review submission")
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_statusLineContains(t, gui, pullRequestReviewSuccessMessage)
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
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
	subject.model.UpdateActionsPopupSearch(pullRequestReviewCommentComposerTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestReviewCommentComposerTitle))
	actualErr = subject.afterStateChange(gui)
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
	subject.asyncRunner = &capturingAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch(pullRequestRequestChangesComposerTitle, matchingActionsPopupIndexes(subject.currentActionsPopupActions(), pullRequestRequestChangesComposerTitle))
	actualErr = subject.afterStateChange(gui)
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
	if strings.Contains(composerView.Title, "boom") {
		t.Fatalf("expected composer title to hide %q, actual %q", "boom", composerView.Title)
	}
	then_transientErrorPopupContains(t, gui, "boom")
}
