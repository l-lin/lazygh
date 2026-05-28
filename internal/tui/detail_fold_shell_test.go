package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestUpdate_GivenMsgToggleInlineConversationVisibility_WhenApplying_ThenItReturnsATypedToggleInlineConversationCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgToggleInlineConversationVisibility{})

	if len(actual) != 1 {
		t.Fatalf("expected one toggle-inline-conversation command, actual %d", len(actual))
	}
	if _, ok := actual[0].(toggleInlineConversationVisibilityCmd); !ok {
		t.Fatalf("expected a toggleInlineConversationVisibilityCmd, actual %T", actual[0])
	}
}

func TestUpdate_GivenMsgSetAllDetailFolds_WhenApplying_ThenItReturnsATypedBulkDetailFoldCommand(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := Update(subject, MsgSetAllDetailFolds{Collapsed: true})

	if len(actual) != 1 {
		t.Fatalf("expected one set-all-detail-folds command, actual %d", len(actual))
	}
	command, ok := actual[0].(setAllDetailFoldsCmd)
	if !ok {
		t.Fatalf("expected a setAllDetailFoldsCmd, actual %T", actual[0])
	}
	if !command.Collapsed {
		t.Fatal("expected the typed bulk-fold command to preserve the collapsed request")
	}
}

func TestUpdate_GivenMsgToggleInlineConversationVisibilityResolved_WhenApplying_ThenItCollapsesTheSelectedBrowserConversationAndKeepsTheCursorOnTheThreadHeader(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/comments",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{
					{
						ID:       "thread-1",
						Path:     "internal/tui/render.go",
						Line:     43,
						DiffSide: "RIGHT",
						Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Thread body one", CreatedAt: "2026-04-18T10:00:00Z", DiffHunk: "@@ -42,2 +42,2 @@\n-old\n+new"}},
					},
					{
						ID:       "thread-2",
						Path:     "internal/tui/model.go",
						Line:     60,
						DiffSide: "RIGHT",
						Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Thread body two", CreatedAt: "2026-04-18T10:05:00Z", DiffHunk: "@@ -59,2 +59,2 @@\n-old\n+new"}},
					},
				},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body one": "Rendered thread body one", "Thread body two": "Rendered thread body two"}}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	then_noError(t, subject.nextDetailTab(gui, nil))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/model.go:60")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	summary, ok := subject.currentPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	result, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected a selected pull request detail")
	}
	expectedCursor, ok := subject.browserConversationSectionAtCursor(summary, result.detail, document.width, subject.detailState.viewState.cursor.line)
	if !ok {
		t.Fatal("expected the detail cursor to resolve a browser conversation section")
	}

	actual := Update(subject, MsgToggleInlineConversationVisibilityResolved{Document: document, ViewportHeight: viewPageSize(detailView)})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.browserCollapsedSectionStates[expectedCursor.section.id]; !actual {
		t.Fatalf("expected browser conversation %q to become collapsed", expectedCursor.section.id)
	}
	updatedDocument := subject.buildCurrentDetailDocument(document.width)
	if strings.Contains(string(updatedDocument.text), "Rendered thread body two") {
		t.Fatalf("expected the collapsed thread body to disappear from the updated detail document, actual %q", string(updatedDocument.text))
	}
	if actual := subject.detailState.viewState.cursor.line; actual != expectedCursor.headerFocusLine {
		t.Fatalf("expected focused cursor line %d, actual %d", expectedCursor.headerFocusLine, actual)
	}
}

func TestUpdate_GivenMsgSetAllDetailFoldsResolved_WhenApplying_ThenItCollapsesAllBrowserOverviewSectionsAndKeepsTheCursorOnTheSelectedHeader(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/folds",
				State:       "OPEN",
				Reviews: []githubcli.PullRequestReview{{
					Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"},
					State:       "APPROVED",
					SubmittedAt: "2026-04-21T10:00:00Z",
				}},
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE"}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	then_noError(t, subject.layout(gui))
	then_noError(t, subject.openDetail(gui, nil))
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Merge Checks")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	document := subject.currentDetailDocument(detailView)
	summary, ok := subject.currentPullRequestSummary()
	if !ok {
		t.Fatal("expected a selected pull request summary")
	}
	result, ok := subject.pullRequestDetailForSummary(summary)
	if !ok {
		t.Fatal("expected a selected pull request detail")
	}
	expectedCursor, ok := subject.browserOverviewSectionAtCursor(summary, result.detail, document.width, subject.detailState.viewState.cursor.line)
	if !ok {
		t.Fatal("expected the detail cursor to resolve an overview section")
	}

	actual := Update(subject, MsgSetAllDetailFoldsResolved{Collapsed: true, Document: document, ViewportHeight: viewPageSize(detailView)})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actual := subject.browserCollapsedSectionStates[expectedCursor.section.id]; !actual {
		t.Fatalf("expected overview section %q to become collapsed", expectedCursor.section.id)
	}
	updatedDocument := subject.buildCurrentDetailDocument(document.width)
	if strings.Contains(string(updatedDocument.text), "@reviewer-approved") || strings.Contains(string(updatedDocument.text), "CI / test (Failed)") {
		t.Fatalf("expected the collapsed overview entries to disappear from the updated detail document, actual %q", string(updatedDocument.text))
	}
	if actual := subject.detailState.viewState.cursor.line; actual != expectedCursor.headerFocusLine {
		t.Fatalf("expected focused cursor line %d, actual %d", expectedCursor.headerFocusLine, actual)
	}
}

func TestUpdate_GivenMsgDetailViewSyncPlanResolvedWithFocusedLine_WhenApplying_ThenItAppliesTheFocusedDetailState(t *testing.T) {
	model := NewModel(SeedData{Users: []Item{{Title: "Only user", Detail: "Alpha\nBeta\nGamma"}}})
	model.OpenDetail()
	subject := NewProgramWithModel(model)
	document := newDetailDocument("Alpha\nBeta\nGamma", 40)
	subject.detailState = subject.detailState.synced(subject.currentDetailIdentity(), document, 2, subject.model.DetailSearchQuery())
	subject.detailState.viewState.cursor = detailPosition{line: 0, column: 0}

	actual := Update(subject, MsgDetailViewSyncPlanResolved{Plan: detailViewSyncPlan{document: document, focusLine: 2, focusLineKnown: true}, ViewportHeight: 2})

	if len(actual) != 0 {
		t.Fatalf("expected no follow-up commands, actual %d", len(actual))
	}
	if actualLine := subject.detailState.viewState.cursor.line; actualLine != 2 {
		t.Fatalf("expected focused cursor line %d, actual %d", 2, actualLine)
	}
	expectedOrigin := visibleViewportOrigin(2, 0, 2, document.rowCount())
	if actualOrigin := subject.detailState.viewState.originRow; actualOrigin != expectedOrigin {
		t.Fatalf("expected synced origin row %d, actual %d", expectedOrigin, actualOrigin)
	}
}

func TestToggleInlineConversationVisibilityCommand_GivenAResolvedDetailDocument_WhenExecuting_ThenItDispatchesTheTypedResolvedMessage(t *testing.T) {
	expectedDocument := newDetailDocument("Alpha", 40)
	actualDispatched := []Msg(nil)

	executeToggleInlineConversationVisibilityCommand(detailFoldCommandRuntime{
		dispatch: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		currentDetailDocument: func(view *gocui.View) detailDocument {
			return expectedDocument
		},
	}, nil, toggleInlineConversationVisibilityCmd{})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgToggleInlineConversationVisibilityResolved)
	if !ok {
		t.Fatalf("expected a MsgToggleInlineConversationVisibilityResolved, actual %T", actualDispatched[0])
	}
	if actual := message.Document.id; actual != expectedDocument.id {
		t.Fatalf("expected resolved document %d, actual %d", expectedDocument.id, actual)
	}
	if actual := message.ViewportHeight; actual != 1 {
		t.Fatalf("expected fallback viewport height %d, actual %d", 1, actual)
	}
}

func TestSetAllDetailFoldsCommand_GivenAResolvedDetailDocument_WhenExecuting_ThenItDispatchesTheTypedResolvedMessage(t *testing.T) {
	expectedDocument := newDetailDocument("Alpha", 40)
	actualDispatched := []Msg(nil)

	executeSetAllDetailFoldsCommand(detailFoldCommandRuntime{
		dispatch: func(gui *gocui.Gui, msg Msg) error {
			actualDispatched = append(actualDispatched, msg)
			return nil
		},
		currentDetailDocument: func(view *gocui.View) detailDocument {
			return expectedDocument
		},
	}, nil, setAllDetailFoldsCmd{Collapsed: true})

	if len(actualDispatched) != 1 {
		t.Fatalf("expected one dispatched message, actual %d", len(actualDispatched))
	}
	message, ok := actualDispatched[0].(MsgSetAllDetailFoldsResolved)
	if !ok {
		t.Fatalf("expected a MsgSetAllDetailFoldsResolved, actual %T", actualDispatched[0])
	}
	if !message.Collapsed {
		t.Fatal("expected collapsed request to stay true")
	}
	if actual := message.Document.id; actual != expectedDocument.id {
		t.Fatalf("expected resolved document %d, actual %d", expectedDocument.id, actual)
	}
	if actual := message.ViewportHeight; actual != 1 {
		t.Fatalf("expected fallback viewport height %d, actual %d", 1, actual)
	}
}
