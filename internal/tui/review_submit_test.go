package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenReviewMode_WhenOpening_ThenItShowsReviewSubmitAndNavigationActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	then_currentViewNameIs(t, gui, viewActionsPopupName)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	for _, expected := range []string{"Review: Submit comment", "Review: Submit approval", "Review: Submit request changes", "Yank URL to clipboard", "Open PR in browser"} {
		if !strings.Contains(popupView.Buffer(), expected) {
			t.Fatalf("expected popup buffer to contain %q, actual %q", expected, popupView.Buffer())
		}
	}
	if strings.Contains(popupView.Buffer(), "Start review") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "Start review", popupView.Buffer())
	}
	if strings.Contains(popupView.Buffer(), "5 of 5 actions") {
		t.Fatalf("expected popup buffer to hide %q, actual %q", "5 of 5 actions", popupView.Buffer())
	}
	if popupView.Footer != "" {
		t.Fatalf("expected popup footer to stay empty without a search query, actual %q", popupView.Footer)
	}
}

func TestActionsPopup_GivenReviewModeSubmitCommentActionSelected_WhenSubmitting_ThenItSubmitsThePendingReviewExitsReviewModeAndRefreshesTheBrowser(t *testing.T) {
	loader := &fakePullRequestDetailLoader{details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": {Title: "First PR", Number: 42, Body: "Original body", State: "OPEN"}}, startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("submit comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "submit comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewModalEditorName)

	subject.modalEditor.editor.SetText("Looks good overall")
	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.submitReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected submitted review ids %v, actual %v", []string{"PRR_pending"}, loader.submitReviewIDs)
	}
	if !reflect.DeepEqual(loader.submitReviewEvents, []githubcli.PullRequestReviewEvent{githubcli.PullRequestReviewEventComment}) {
		t.Fatalf("expected submitted review events %v, actual %v", []githubcli.PullRequestReviewEvent{githubcli.PullRequestReviewEventComment}, loader.submitReviewEvents)
	}
	if !reflect.DeepEqual(loader.submitReviewBodies, []string{"Looks good overall"}) {
		t.Fatalf("expected submitted review bodies %v, actual %v", []string{"Looks good overall"}, loader.submitReviewBodies)
	}
	if subject.reviewSession.active {
		t.Fatal("expected review mode to be inactive after submit")
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42", "acme/widgets#42"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#42", "acme/widgets#42"}, loader.detailCalls)
	}
	if _, ok := subject.pullRequestDiffCache["acme/widgets#42"]; ok {
		t.Fatal("expected the cached pull request diff to be invalidated after review submit")
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	pullRequestsFooterView, actualErr := gui.View(viewPullRequestsFooterName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsFooterView.Buffer(), pullRequestReviewSuccessMessage) {
		t.Fatalf("expected pull requests footer to contain %q, actual %q", pullRequestReviewSuccessMessage, pullRequestsFooterView.Buffer())
	}

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), "Start review") {
		t.Fatalf("expected browser actions to return after review submit, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenReviewModeSubmitApprovalActionSelected_WhenSubmittingWithAnEmptySummary_ThenItSubmitsThePendingReviewAsApprove(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("submit approval", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "submit approval"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if !reflect.DeepEqual(loader.submitReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected submitted review ids %v, actual %v", []string{"PRR_pending"}, loader.submitReviewIDs)
	}
	if !reflect.DeepEqual(loader.submitReviewEvents, []githubcli.PullRequestReviewEvent{githubcli.PullRequestReviewEventApprove}) {
		t.Fatalf("expected submitted review events %v, actual %v", []githubcli.PullRequestReviewEvent{githubcli.PullRequestReviewEventApprove}, loader.submitReviewEvents)
	}
	if !reflect.DeepEqual(loader.submitReviewBodies, []string{""}) {
		t.Fatalf("expected submitted review bodies %v, actual %v", []string{""}, loader.submitReviewBodies)
	}
}

func TestActionsPopup_GivenReviewModeSubmitRequestChangesActionSelected_WhenSubmittingFails_ThenItKeepsTheDraftAndPendingReviewVisible(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", submitReviewErr: errors.New("boom")}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("submit request changes", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "submit request changes"))
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

	if !subject.reviewSession.active {
		t.Fatal("expected review mode to stay active after the submit error")
	}
	if subject.reviewSession.pendingReviewID != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", subject.reviewSession.pendingReviewID)
	}
	if !reflect.DeepEqual(loader.submitReviewIDs, []string{"PRR_pending"}) {
		t.Fatalf("expected submitted review ids %v, actual %v", []string{"PRR_pending"}, loader.submitReviewIDs)
	}
	if !reflect.DeepEqual(loader.submitReviewEvents, []githubcli.PullRequestReviewEvent{githubcli.PullRequestReviewEventRequestChanges}) {
		t.Fatalf("expected submitted review events %v, actual %v", []githubcli.PullRequestReviewEvent{githubcli.PullRequestReviewEventRequestChanges}, loader.submitReviewEvents)
	}
	if !reflect.DeepEqual(loader.submitReviewBodies, []string{"Needs tests"}) {
		t.Fatalf("expected submitted review bodies %v, actual %v", []string{"Needs tests"}, loader.submitReviewBodies)
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
