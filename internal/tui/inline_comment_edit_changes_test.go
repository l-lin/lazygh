package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenBrowserChangesTabCursorOnAnOwnedInlineComment_WhenOpening_ThenItShowsUpdateAndDeleteInlineCommentActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), inlineCommentUpdateEditorTitle) {
		t.Fatalf("expected the popup to contain the inline comment update action, actual %q", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), inlineCommentDeleteActionTitle) {
		t.Fatalf("expected the popup to contain the inline comment delete action, actual %q", popupView.Buffer())
	}
}

func TestEditInlineComment_GivenBrowserChangesTabSubmit_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingBackgroundRefreshes(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original inline body": "Rendered original inline body",
		"Updated inline body":  "Rendered updated inline body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("edit inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "edit inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)
	subject.modalEditor.editor.SetText("Updated inline body")

	actualHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewModalEditorName, gocui.KeyAltEnter)
	actualErr = actualHandler(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected two queued background refreshes, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued runs, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh call before the queued runs, actual %v", loader.diffCalls)
	}
	if !reflect.DeepEqual(loader.updateReviewCommentIDs, []string{"PRRC_1"}) {
		t.Fatalf("expected updated inline comment ids %v, actual %v", []string{"PRRC_1"}, loader.updateReviewCommentIDs)
	}
	if !reflect.DeepEqual(loader.updateReviewCommentBodies, []string{"Updated inline body"}) {
		t.Fatalf("expected updated inline comment bodies %v, actual %v", []string{"Updated inline body"}, loader.updateReviewCommentBodies)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered updated inline body") {
		t.Fatalf("expected detail buffer to contain %q, actual %q", "Rendered updated inline body", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "new line") {
		t.Fatalf("expected detail buffer to keep the diff content %q, actual %q", "new line", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (1)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 3)
	then_statusLineContains(t, gui, inlineCommentUpdatedSuccessMessage)
}

func TestDeleteInlineComment_GivenBrowserChangesTabAction_WhenSubmittingOptimistically_ThenItKeepsTheRenderedDiffVisibleWhileQueueingBackgroundRefreshes(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Original inline body": "Rendered original inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	subject.activeDetailTab = ChangesDetailTab
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	asyncRunner := &capturingAsyncRunner{}
	subject.asyncRunner = asyncRunner
	actualErr = subject.openActionsPopup(gui, nil)
	then_noError(t, actualErr)
	subject.model.UpdateActionsPopupSearch("delete inline comment", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "delete inline comment"))
	actualErr = subject.afterStateChange(gui)
	then_noError(t, actualErr)
	actualErr = subject.executeSelectedActionsPopupAction(gui, nil)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 2 {
		t.Fatalf("expected two queued background refreshes, actual %d", len(asyncRunner.runs))
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager detail refresh call before the queued runs, actual %v", loader.detailCalls)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected no eager diff refresh call before the queued runs, actual %v", loader.diffCalls)
	}
	if !reflect.DeepEqual(loader.deleteReviewCommentIDs, []string{"PRRC_1"}) {
		t.Fatalf("expected deleted inline comment ids %v, actual %v", []string{"PRRC_1"}, loader.deleteReviewCommentIDs)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered original inline body") {
		t.Fatalf("expected detail buffer to remove %q, actual %q", "Rendered original inline body", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "new line") {
		t.Fatalf("expected detail buffer to keep the diff content %q, actual %q", "new line", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail buffer to avoid the loading spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 3)
	then_statusLineContains(t, gui, inlineCommentDeletedSuccessMessage)
}

func given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests() githubcli.PullRequestDetail {
	return githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
		Body:        "Body 42",
		BaseRefName: "main",
		HeadRefName: "feature/comments",
		State:       "OPEN",
		InlineCommentThreads: []githubcli.PullRequestReviewThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     43,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				ID:              "PRRC_1",
				ViewerDidAuthor: true,
				Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				Body:            "Original inline body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
			}},
		}},
	}
}

func given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests() githubcli.PullRequestDiff {
	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -42,2 +42,2 @@",
			" context line",
			"-old line",
			"+new line",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 1, Deletions: 1}},
		Threads: []githubcli.PullRequestReviewThread{{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     43,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				ID:              "PRRC_1",
				ViewerDidAuthor: true,
				Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				Body:            "Original inline body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
			}},
		}},
	}
}
