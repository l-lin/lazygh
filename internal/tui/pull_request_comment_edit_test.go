package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenBrowserCommentsTabCursorOnAnOwnedPullRequestComment_WhenOpening_ThenItShowsUpdateAndDeletePullRequestCommentActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedCommentForEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original PR comment": "Rendered original PR comment"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original PR comment")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), pullRequestCommentUpdateEditorTitle) {
		t.Fatalf("expected the popup to contain the PR comment update action, actual %q", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), pullRequestCommentDeleteActionTitle) {
		t.Fatalf("expected the popup to contain the PR comment delete action, actual %q", popupView.Buffer())
	}
}

func TestEditPullRequestComment_GivenBrowserCommentsTabSubmit_WhenSubmittingOptimistically_ThenItKeepsTheRenderedCommentsVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedCommentForEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original PR comment": "Rendered original PR comment",
		"Updated PR comment":  "Rendered updated PR comment",
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original PR comment")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("update pr comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "update pr comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Updated PR comment")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued run, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.updatePullRequestCommentIDs, []string{"IC_kwDOA"}) {
		t.Fatalf("expected updated PR comment ids %v, actual %v", []string{"IC_kwDOA"}, loader.updatePullRequestCommentIDs)
	}
	if !reflect.DeepEqual(loader.updatePullRequestCommentBodies, []string{"Updated PR comment"}) {
		t.Fatalf("expected updated PR comment bodies %v, actual %v", []string{"Updated PR comment"}, loader.updatePullRequestCommentBodies)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered updated PR comment") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered updated PR comment", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (1)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 1)
	then_statusLineContains(t, gui, pullRequestCommentUpdatedSuccessMessage)
}

func TestDeletePullRequestComment_GivenBrowserCommentsTabAction_WhenSubmittingOptimistically_ThenItKeepsTheCommentsTabVisibleWhileQueueingABackgroundRefresh(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedCommentForEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original PR comment": "Rendered original PR comment"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original PR comment")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("delete pr comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "delete pr comment"))
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued background refresh, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued run, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.deletePullRequestCommentIDs, []string{"IC_kwDOA"}) {
		t.Fatalf("expected deleted PR comment ids %v, actual %v", []string{"IC_kwDOA"}, loader.deletePullRequestCommentIDs)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered original PR comment") {
		t.Fatalf("expected detail buffer to remove %q, actual %q", "Rendered original PR comment", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 1)
	then_statusLineContains(t, gui, pullRequestCommentDeletedSuccessMessage)
}

func given_pullRequestDetailWithOwnedCommentForEditTests() githubcli.PullRequestDetail {
	return githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
		Body:        "Body 42",
		BaseRefName: "main",
		HeadRefName: "feature/comments",
		State:       "OPEN",
		Comments: []githubcli.PullRequestComment{{
			ID:              "IC_kwDOA",
			ViewerDidAuthor: true,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
			Body:            "Original PR comment",
			CreatedAt:       "2026-04-18T10:00:00Z",
		}},
	}
}
