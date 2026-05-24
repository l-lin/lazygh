package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestActionsPopup_GivenBrowserCommentsTabCursorOnANonOwnedInlineComment_WhenOpening_ThenItShowsUpdateAndDeleteInlineCommentActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithNonOwnedInlineThreadForEditPermissionTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original inline body": "Rendered original inline body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	then_noError(t, subject.nextDetailTab(gui, nil))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), inlineCommentUpdateEditorTitle) {
		t.Fatalf("expected the popup to contain the inline comment update action, actual %q", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), inlineCommentDeleteActionTitle) {
		t.Fatalf("expected the popup to contain the inline comment delete action, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenBrowserChangesTabCursorOnANonOwnedInlineComment_WhenOpening_ThenItShowsUpdateAndDeleteInlineCommentActions(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithNonOwnedInlineThreadForEditPermissionTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithNonOwnedInlineThreadForEditPermissionTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Original inline body": "Rendered original inline body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	subject.activeDetailTab = ChangesDetailTab
	then_noError(t, subject.afterStateChange(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")

	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), inlineCommentUpdateEditorTitle) {
		t.Fatalf("expected the popup to contain the inline comment update action, actual %q", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), inlineCommentDeleteActionTitle) {
		t.Fatalf("expected the popup to contain the inline comment delete action, actual %q", popupView.Buffer())
	}
}

func TestActionsPopup_GivenReviewModeCursorOnANonOwnedInlineComment_WhenOpening_ThenItShowsUpdateAndDeleteInlineCommentActions(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			ID:              "PRRC_1",
			ViewerDidAuthor: false,
			Author:          &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:            "Thread body",
			CreatedAt:       "2026-04-20T10:00:00Z",
			DiffHunk:        "@@ -1,2 +1,3 @@\n context\n-old line\n+new line",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))
	then_noError(t, subject.focusDetailView(gui, nil))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered thread body")

	then_noError(t, subject.openActionsPopup(gui, nil))

	popupView, actualErr := gui.View(viewActionsPopupName)
	then_noError(t, actualErr)
	if !strings.Contains(popupView.Buffer(), inlineCommentUpdateEditorTitle) {
		t.Fatalf("expected the popup to contain the inline comment update action, actual %q", popupView.Buffer())
	}
	if !strings.Contains(popupView.Buffer(), inlineCommentDeleteActionTitle) {
		t.Fatalf("expected the popup to contain the inline comment delete action, actual %q", popupView.Buffer())
	}
}

func given_pullRequestDetailWithNonOwnedInlineThreadForEditPermissionTests() githubcli.PullRequestDetail {
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
				ViewerDidAuthor: false,
				Author:          &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:            "Original inline body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
			}},
		}},
	}
}

func given_pullRequestDiffWithNonOwnedInlineThreadForEditPermissionTests() githubcli.PullRequestDiff {
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
				ViewerDidAuthor: false,
				Author:          &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:            "Original inline body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				DiffHunk:        "@@ -42,2 +42,2 @@\n context line\n-old line\n+new line",
			}},
		}},
	}
}
