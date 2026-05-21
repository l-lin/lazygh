package tui

import (
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestCurrentDetailDocument_GivenTheCommentsTabForTheSamePullRequest_WhenBuildingItTwice_ThenItReusesTheCachedDocument(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Comment body": "Rendered comment body",
		"Inline body":  "Rendered inline body",
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.activeDetailTab = CommentsDetailTab
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:  "First PR",
		Number: 42,
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Comment body",
			CreatedAt: "2026-04-18T10:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{
			Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:         "Inline body",
			CreatedAt:    "2026-04-18T10:30:00Z",
			Path:         "internal/tui/render.go",
			Line:         43,
			OriginalLine: 43,
			Side:         "RIGHT",
			DiffHunk:     "@@ -42,1 +42,1 @@\n-old line\n+new line",
		}},
	})}

	firstDocument := subject.currentDetailDocument(nil)
	secondDocument := subject.currentDetailDocument(nil)

	if string(firstDocument.text) != string(secondDocument.text) {
		t.Fatalf("expected cached detail document %q, actual %q", string(firstDocument.text), string(secondDocument.text))
	}
	if renderer.callCount != 2 {
		t.Fatalf("expected one markdown render per unique comment body, actual %d", renderer.callCount)
	}
}

func TestBrowserConversationSectionAtCursor_GivenCommentsTabDocumentAlreadyBuilt_WhenResolvingTheSameInlineThreadCursorTwice_ThenItReusesTheSemanticLineMapWithoutRenderingMarkdownAgain(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline body": "Rendered inline body",
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		Title:       "First PR",
		Number:      42,
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
				Body:            "Inline body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				ReactionGroups:  []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2}},
				DiffHunk:        "@@ -42,1 +42,1 @@\n-old line\n+new line",
			}},
		}},
	})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	then_noError(t, subject.nextDetailTab(gui, nil))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline body")

	summary, ok := subject.selectedPullRequestSummaryForDetail()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	result, ok := subject.pullRequestDetailForSummary(summary)
	if !ok || result.err != nil {
		t.Fatalf("expected cached detail, actual %+v", result)
	}
	cursorLine := subject.detailViewState.cursor.line
	renderer.callCount = 0

	for range 2 {
		actual, ok := subject.browserConversationSectionAtCursor(summary, result.detail, subject.detailWrapWidth, cursorLine)
		if !ok {
			t.Fatal("expected a conversation section at the inline thread cursor")
		}
		if actual.section.inlineThread == nil || actual.section.inlineThread.ID != "thread-1" {
			t.Fatalf("expected inline thread %q, actual %+v", "thread-1", actual.section.inlineThread)
		}
		if !actual.inBody {
			t.Fatalf("expected the cursor to stay in the inline thread body, actual %+v", actual)
		}
	}

	if renderer.callCount != 0 {
		t.Fatalf("expected the semantic line map lookup to avoid markdown rendering, actual %d", renderer.callCount)
	}
}

func TestCurrentActionsPopupActions_GivenCommentsTabDocumentAlreadyBuilt_WhenResolvingTwice_ThenItReusesTheSemanticLineMapWithoutRenderingMarkdownAgain(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline body": "Rendered inline body",
	}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), &fakePullRequestDetailLoader{})
	subject.markdownRenderer = renderer
	subject.pullRequestDetailCache["acme/widgets#42"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{
		ID:          "PR_kwDOA",
		Title:       "First PR",
		Number:      42,
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
				Body:            "Inline body",
				CreatedAt:       "2026-04-18T10:30:00Z",
				ReactionGroups:  []githubcli.ReactionGroup{{Content: githubcli.ReactionContentEyes, TotalCount: 2}},
				DiffHunk:        "@@ -42,1 +42,1 @@\n-old line\n+new line",
			}},
		}},
	})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	then_noError(t, subject.nextDetailTab(gui, nil))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered inline body")
	renderer.callCount = 0

	firstActions := subject.currentActionsPopupActions()
	secondActions := subject.currentActionsPopupActions()

	for _, expectedTitle := range []string{reactionPickerTitle, pullRequestInlineCommentReplyEditorTitle, inlineCommentUpdateEditorTitle, inlineCommentDeleteActionTitle, "Mark inline comment as resolved"} {
		if !given_hasActionTitle(firstActions, expectedTitle) {
			t.Fatalf("expected actions to contain %q, actual %v", expectedTitle, given_actionTitles(firstActions))
		}
		if !given_hasActionTitle(secondActions, expectedTitle) {
			t.Fatalf("expected repeated actions to contain %q, actual %v", expectedTitle, given_actionTitles(secondActions))
		}
	}
	if renderer.callCount != 0 {
		t.Fatalf("expected repeated actions popup lookups to avoid markdown rendering, actual %d", renderer.callCount)
	}
}

func TestCurrentDetailDocument_GivenReviewModeDescriptionAlreadyBuilt_WhenBuildingItTwice_ThenItReusesTheCachedDocumentWithoutRenderingMarkdownAgain(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42": "Rendered body 42",
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = renderer
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, given_startingReviewMode(t, gui, subject))
	then_noError(t, subject.focusUserView(gui, nil))
	if actualErr := subject.refreshDetailView(gui); actualErr != nil {
		t.Fatalf("expected no error, actual %v", actualErr)
	}
	renderer.callCount = 0

	firstDocument := subject.currentDetailDocument(nil)
	secondDocument := subject.currentDetailDocument(nil)

	if string(firstDocument.text) != string(secondDocument.text) {
		t.Fatalf("expected cached review description document %q, actual %q", string(firstDocument.text), string(secondDocument.text))
	}
	if renderer.callCount != 0 {
		t.Fatalf("expected the cached review description lookup to avoid markdown rendering, actual %d", renderer.callCount)
	}
}

func TestCurrentActionsPopupActions_GivenChangesTabDiffAlreadyRendered_WhenResolvingTwice_ThenItReusesCachedRenderedRowsWithoutRenderingMarkdownAgain(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Original inline body": "Rendered original inline body",
	}}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = renderer
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	subject.activeDetailTab = ChangesDetailTab
	then_noError(t, subject.refreshViews(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")
	renderer.callCount = 0

	firstActions := subject.currentActionsPopupActions()
	secondActions := subject.currentActionsPopupActions()

	for _, expectedTitle := range []string{reactionPickerTitle, pullRequestInlineCommentReplyEditorTitle, inlineCommentUpdateEditorTitle, inlineCommentDeleteActionTitle, "Mark inline comment as resolved"} {
		if !given_hasActionTitle(firstActions, expectedTitle) {
			t.Fatalf("expected actions to contain %q, actual %v", expectedTitle, given_actionTitles(firstActions))
		}
		if !given_hasActionTitle(secondActions, expectedTitle) {
			t.Fatalf("expected repeated actions to contain %q, actual %v", expectedTitle, given_actionTitles(secondActions))
		}
	}
	if renderer.callCount != 0 {
		t.Fatalf("expected repeated changes-tab actions popup lookups to avoid markdown rendering, actual %d", renderer.callCount)
	}
}

func TestMoveActionsPopupSelection_GivenChangesTabPopupAlreadyOpen_WhenNavigating_ThenItReusesCachedRenderedRowsWithoutRenderingMarkdownAgain(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Original inline body": "Rendered original inline body",
	}}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{"acme/widgets#42": given_pullRequestDetailWithOwnedInlineThreadForChangesEditTests()},
		diffs:   map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_pullRequestDiffWithOwnedInlineThreadForChangesEditTests()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = renderer
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	subject.activeDetailTab = ChangesDetailTab
	then_noError(t, subject.refreshViews(gui))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Rendered original inline body")
	then_noError(t, subject.openActionsPopup(gui, nil))
	renderer.callCount = 0

	then_noError(t, subject.moveActionsPopupSelectionDown(gui, nil))
	then_noError(t, subject.moveActionsPopupSelectionUp(gui, nil))

	if renderer.callCount != 0 {
		t.Fatalf("expected changes-tab popup navigation to avoid markdown rendering, actual %d", renderer.callCount)
	}
}
