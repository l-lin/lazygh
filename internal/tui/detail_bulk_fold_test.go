package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/githubcli"
)

func TestKeybindingSpecs_GivenProgram_WhenListingInlineConversationBindings_ThenDetailViewSupportsZMAndZRAsBulkFoldBindings(t *testing.T) {
	subject := NewProgramWithModel(given_model())

	actual := subject.keybindingSpecs()

	then_bindingKeyExists(t, actual, viewDetailName, 'M')
	then_bindingKeyExists(t, actual, viewDetailName, 'R')
}

func TestBrowserMode_GivenDescriptionOverviewSections_WhenPressingZMAndZR_ThenItClosesAndOpensEverySectionWhileKeepingTheCursorOnTheSameSection(t *testing.T) {
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

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Merge Checks")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "@reviewer-approved") || strings.Contains(detailView.Buffer(), "CI / test (Failed)") {
		t.Fatalf("expected zM to collapse every overview section, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "Merge Checks")

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "@reviewer-approved") || !strings.Contains(detailView.Buffer(), "CI / test (Failed)") {
		t.Fatalf("expected zR to expand every overview section, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "Merge Checks")
}

func TestReviewMode_GivenDescriptionOverviewSections_WhenPressingZMAndZR_ThenItClosesAndOpensEverySectionWhileKeepingTheCursorOnTheSameSection(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/review-overview-folds",
				State:       "OPEN",
				Reviews: []githubcli.PullRequestReview{{
					Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"},
					State:       "APPROVED",
					SubmittedAt: "2026-04-21T10:00:00Z",
				}},
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE"}},
			},
		},
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
	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Merge Checks")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "@reviewer-approved") || !strings.Contains(detailView.Buffer(), "CI / test (Failed)") {
		t.Fatalf("expected zR to expand every review overview section, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "Merge Checks")

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "@reviewer-approved") || strings.Contains(detailView.Buffer(), "CI / test (Failed)") {
		t.Fatalf("expected zM to collapse every review overview section, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "Merge Checks")
}

func TestBrowserMode_GivenCommentsTabInlineConversations_WhenPressingZMAndZR_ThenItClosesAndOpensEveryConversationWhileKeepingTheCursorOnTheSameThread(t *testing.T) {
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

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/model.go:60")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered thread body one") || strings.Contains(detailView.Buffer(), "Rendered thread body two") {
		t.Fatalf("expected zM to collapse every browser conversation, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go:60")

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered thread body one") || !strings.Contains(detailView.Buffer(), "Rendered thread body two") {
		t.Fatalf("expected zR to expand every browser conversation, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go:60")
}

func TestBrowserMode_GivenChangesTabFiles_WhenPressingZMAndZR_ThenItClosesAndOpensEveryFileWhileKeepingTheCursorOnTheSameFileHeader(t *testing.T) {
	loader := &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/model.go")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, detailView)
	then_noError(t, actualErr)
	for _, expected := range []string{" " + reviewDiffHeaderPathIcon + " internal/tui/render.go", " " + reviewDiffHeaderPathIcon + " internal/tui/model.go"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected zM to collapse every changes file and keep %q visible, actual %q", expected, detailView.Buffer())
		}
	}
	for _, hidden := range []string{"+another line", "+new model"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected zM to hide %q from collapsed files, actual %q", hidden, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go")

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, detailView)
	then_noError(t, actualErr)
	for _, expected := range []string{" " + reviewDiffHeaderPathIcon + " internal/tui/render.go", " " + reviewDiffHeaderPathIcon + " internal/tui/model.go", "+another line", "+new model"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected zR to reopen every changes file and show %q, actual %q", expected, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go")
}

func TestBrowserMode_GivenChangesTabThreads_WhenPressingZMAndZR_ThenItClosesAndOpensEveryFoldWhileKeepingTheCursorOnTheContainingFileHeader(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{
		{ID: "thread-1", Path: "internal/tui/render.go", Line: 3, DiffSide: "RIGHT", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Thread body one", CreatedAt: "2026-04-20T10:00:00Z"}}},
		{ID: "thread-2", Path: "internal/tui/model.go", Line: 10, DiffSide: "RIGHT", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Thread body two", CreatedAt: "2026-04-20T10:05:00Z"}}},
	}
	loader := &fakePullRequestDetailLoader{diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body one": "Rendered thread body one", "Thread body two": "Rendered thread body two"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/model.go:10")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered thread body one") || strings.Contains(detailView.Buffer(), "Rendered thread body two") {
		t.Fatalf("expected zM to collapse every changes fold, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go")

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered thread body one") || !strings.Contains(detailView.Buffer(), "Rendered thread body two") {
		t.Fatalf("expected zR to expand every changes fold, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/model.go")
}

func TestReviewMode_GivenMultipleInlineConversations_WhenPressingZMAndZR_ThenItClosesAndOpensEveryConversationWhileKeepingTheCursorOnTheSameThread(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{
		{ID: "thread-1", Path: "internal/tui/render.go", Line: 3, DiffSide: "RIGHT", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Thread body one", CreatedAt: "2026-04-20T10:00:00Z"}}},
		{ID: "thread-2", Path: "internal/tui/render.go", Line: 4, DiffSide: "RIGHT", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, Body: "Thread body two", CreatedAt: "2026-04-20T10:05:00Z"}}},
	}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body one": "Rendered thread body one", "Thread body two": "Rendered thread body two"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:4")

	prefixHandler, closeAllHandler, openAllHandler := given_detailBulkFoldHandlers(t, subject)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = closeAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Rendered thread body one") || strings.Contains(detailView.Buffer(), "Rendered thread body two") {
		t.Fatalf("expected zM to collapse every review conversation, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go:4")

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = openAllHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered thread body one") || !strings.Contains(detailView.Buffer(), "Rendered thread body two") {
		t.Fatalf("expected zR to expand every review conversation, actual %q", detailView.Buffer())
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go:4")
}

func given_detailBulkFoldHandlers(t *testing.T, subject *Program) (func(*gocui.Gui, *gocui.View) error, func(*gocui.Gui, *gocui.View) error, func(*gocui.Gui, *gocui.View) error) {
	t.Helper()

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	closeAllHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'M')
	openAllHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'R')
	return prefixHandler, closeAllHandler, openAllHandler
}
