package tui

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	githubdomain "github.com/l-lin/lazygh/internal/github"
	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestLayout_GivenSelectedPullRequestSummary_WhenRendering_ThenItLoadsRichDetailAndShowsDescriptionAndConversationsInSeparateTabs(t *testing.T) {
	firstSummary := githubcli.PullRequest{Title: "First PR", Number: 101, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-1"}
	secondSummary := githubcli.PullRequest{Title: "Second PR", Number: 102, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback-2"}
	firstDetail := githubcli.PullRequestDetail{
		Title:            "First PR",
		Number:           101,
		Body:             "Body 101",
		BaseRefName:      "main",
		HeadRefName:      "feature-101",
		State:            "OPEN",
		Mergeable:        "MERGEABLE",
		MergeStateStatus: "CLEAN",
		ReviewRequests: []githubcli.PullRequestReviewRequest{{
			RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"},
		}},
		Reviews: []githubcli.PullRequestReview{{
			Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"},
			State:       "APPROVED",
			SubmittedAt: "2026-04-21T10:00:00Z",
		}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
			Name:         "lint",
			WorkflowName: "CI",
			Status:       "COMPLETED",
			Conclusion:   "SUCCESS",
		}},
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Comment 101",
			CreatedAt: "2026-04-18T10:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{
			Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
			Body:         "Inline 101",
			CreatedAt:    "2026-04-18T10:30:00Z",
			Path:         "internal/tui/render.go",
			Line:         43,
			OriginalLine: 43,
			Side:         "RIGHT",
			DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
		}},
	}
	secondDetail := githubcli.PullRequestDetail{
		Title:       "Second PR",
		Number:      102,
		Body:        "Body 102",
		BaseRefName: "main",
		HeadRefName: "feature-102",
		State:       "OPEN",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"},
			Body:      "Comment 102",
			CreatedAt: "2026-04-18T11:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{
			Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline-two"},
			Body:         "Inline 102",
			CreatedAt:    "2026-04-18T11:30:00Z",
			Path:         "internal/tui/pull_request_detail_program_test.go",
			Line:         19,
			OriginalLine: 19,
			Side:         "RIGHT",
			DiffHunk:     "@@ -18,2 +18,2 @@\n actualErr := subject.layout(gui)\n-if !strings.Contains(detailView.Buffer(), \"Rendered comment 102\") {\n+if !strings.Contains(detailView.Buffer(), \"Rendered inline 102\") {",
		}},
	}
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(firstSummary),
		myPullRequestRow(secondSummary),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#101": firstDetail,
			"acme/widgets#102": secondDetail,
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 101":    "Rendered body 101",
		"Body 102":    "Rendered body 102",
		"Comment 101": "Rendered comment 101",
		"Comment 102": "Rendered comment 102",
		"Inline 101":  "Rendered inline 101",
		"Inline 102":  "Rendered inline 102",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	expectedSeparator := strings.Repeat("─", detailView.InnerWidth())
	for _, expected := range []string{"acme/widgets#101 First PR", " " + pullRequestOverviewPendingIcon + " Reviewers (1/2)", "@reviewer-requested", " " + pullRequestOverviewPendingIcon + " Merge Checks", " " + pullRequestOverviewSuccessIcon + " Builds", "Rendered body 101"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the overview tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	for _, hidden := range []string{"CI / lint (Successful)", "Changes can be cleanly merged."} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected the description tab to keep non-reviewer overview block bodies folded, actual %q", detailView.Buffer())
		}
	}
	headerLineIndex := given_viewLineIndexContaining(t, detailView, detailStatusIcon+" OPEN")
	reviewersLineIndex := given_viewLineIndexContaining(t, detailView, pullRequestOverviewPendingIcon+" Reviewers (1/2)")
	reviewerEntryLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-requested")
	mergeChecksLineIndex := given_viewLineIndexContaining(t, detailView, pullRequestOverviewPendingIcon+" Merge Checks")
	buildsLineIndex := given_viewLineIndexContaining(t, detailView, pullRequestOverviewSuccessIcon+" Builds")
	separatorLineIndex := given_viewLineIndexContaining(t, detailView, expectedSeparator)
	bodyLineIndex := given_viewLineIndexContaining(t, detailView, "Rendered body 101")
	if !(headerLineIndex < reviewersLineIndex && reviewersLineIndex < reviewerEntryLineIndex && reviewerEntryLineIndex < mergeChecksLineIndex && mergeChecksLineIndex < buildsLineIndex && buildsLineIndex < separatorLineIndex && separatorLineIndex < bodyLineIndex) {
		t.Fatalf("expected header, reviewer overview body, remaining overview headers, separator, and body to stay ordered, actual %q", detailView.Buffer())
	}
	if buildsLineIndex != mergeChecksLineIndex+1 {
		t.Fatalf("expected the still-folded merge check and build headers to stay consecutive without blank spacer lines, actual %q", detailView.Buffer())
	}
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, reviewersLineIndex, pullRequestOverviewPendingIcon+" Reviewers (1/2)", given_themeColorHex(t, theme.PendingHex), "reviewers overview header")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, mergeChecksLineIndex, pullRequestOverviewPendingIcon+" Merge Checks", given_themeColorHex(t, theme.PendingHex), "merge checks overview header")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, buildsLineIndex, pullRequestOverviewSuccessIcon+" Builds", given_themeColorHex(t, theme.SuccessHex), "builds overview header")
	if strings.Contains(detailView.Buffer(), "Rendered comment 101") {
		t.Fatalf("expected overview tab to hide comments, actual %q", detailView.Buffer())
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (2)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 0)
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#101"}, loader.detailCalls)
	}

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	for _, hidden := range []string{"acme/widgets#101 First PR", detailStatusIcon + " OPEN", "Reviewers", "Merge Checks", "Builds", "Rendered body 101"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected the conversations tab to hide overview metadata %q, actual %q", hidden, detailView.Buffer())
		}
	}
	for _, expected := range []string{" Comment", "Rendered comment 101", " Comment on line R43", detailInlineCommentLocationIcon + " internal/tui/render.go:43  +1  -1", "Rendered inline 101"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the conversations tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (2)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 1)

	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101"}) {
		t.Fatalf("expected cached detail calls %v, actual %v", []string{"acme/widgets#101"}, loader.detailCalls)
	}

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "Rendered comment 102") {
		t.Fatalf("expected comments tab to contain %q after selection, actual %q", "Rendered comment 102", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered inline 102") {
		t.Fatalf("expected comments tab to contain %q after selection, actual %q", "Rendered inline 102", detailView.Buffer())
	}
	if !reflect.DeepEqual(loader.detailCalls, []string{"acme/widgets#101", "acme/widgets#102"}) {
		t.Fatalf("expected detail calls %v, actual %v", []string{"acme/widgets#101", "acme/widgets#102"}, loader.detailCalls)
	}
}

func TestLayout_GivenPullRequestCommits_WhenRendering_ThenBrowserModeShowsFourDetailTabsAndRendersTheCommitTab(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/commits",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "Comment body",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
				Commits: []githubcli.PullRequestCommit{
					{
						OID:             "1111111aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						MessageHeadline: "older commit",
						MessageBody:     "Older body",
						CommittedDate:   "2026-05-19T10:00:00Z",
						Authors:         []githubcli.PullRequestCommitAuthor{{Name: "Older Dev"}},
					},
					{
						OID:             "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						MessageHeadline: "newer commit",
						MessageBody:     "Newer body",
						CommittedDate:   "2026-05-20T10:00:00Z",
						Authors:         []githubcli.PullRequestCommitAuthor{{Name: "Newer Dev"}},
					},
				},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Body 42":    "Rendered body 42",
		"Older body": "Rendered older body",
		"Newer body": "Rendered newer body",
	}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (1)", CommitsDetailTab.Label() + " (2)", ChangesDetailTab.Label()}, 0)

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	if subject.activeDetailTab != CommentsDetailTab {
		t.Fatalf("expected active detail tab %v, actual %v", CommentsDetailTab, subject.activeDetailTab)
	}
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	if subject.activeDetailTab != CommitsDetailTab {
		t.Fatalf("expected active detail tab %v, actual %v", CommitsDetailTab, subject.activeDetailTab)
	}
	for _, expected := range []string{"● 2222222 newer commit", "│ Authors: Newer Dev", "│ Rendered newer body", "● 1111111 older commit", "│ Authors: Older Dev", "│ Rendered older body"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the commits tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	if strings.Contains(detailView.Buffer(), "╭") || strings.Contains(detailView.Buffer(), "╰") {
		t.Fatalf("expected the commits tab to avoid rounded boxes, actual %q", detailView.Buffer())
	}
	newerHeaderLineIndex := given_viewLineIndexContaining(t, detailView, "● 2222222 newer commit")
	olderHeaderLineIndex := given_viewLineIndexContaining(t, detailView, "● 1111111 older commit")
	if newerHeaderLineIndex >= olderHeaderLineIndex {
		t.Fatalf("expected the newer commit to render before the older commit, actual %q", detailView.Buffer())
	}
	if detailView.BufferLines()[olderHeaderLineIndex-1] != "│" {
		t.Fatalf("expected the commits timeline to keep the vertical rail between commits, actual %q", detailView.BufferLines()[olderHeaderLineIndex-1])
	}

	actualErr = subject.previousDetailTab(gui, nil)
	then_noError(t, actualErr)
	if subject.activeDetailTab != CommentsDetailTab {
		t.Fatalf("expected active detail tab %v after going backward, actual %v", CommentsDetailTab, subject.activeDetailTab)
	}
}

func TestLayout_GivenPullRequestChanges_WhenRendering_ThenTheBrowserChangesTabLoadsTheDiffAndShowsTheRenderedFilesAndInlineComments(t *testing.T) {
	diff := githubcli.PullRequestDiff{
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
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
				Body:      "Inline thread body",
				CreatedAt: "2026-04-18T10:30:00Z",
			}},
		}},
	}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGuiWithSize(t, 120, 50)
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label() + " (0)", CommitsDetailTab.Label() + " (0)", ChangesDetailTab.Label()}, 3)
	for _, expected := range []string{"internal/tui/render.go", "@@ -42,2 +42,2 @@", "42 : 42 │  context line", "43 :    │ -old line", "   : 43 │ +new line", "Rendered inline thread body"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the changes tab to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
}

func TestBrowserMode_GivenAChangesTabFileHeader_WhenPressingEnterAndZA_ThenItTogglesTheFileVisibility(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") {
		t.Fatalf("expected the file header to start expanded in changes, actual %q", detailView.Buffer())
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") {
		t.Fatalf("expected enter to collapse the file header, actual %q", detailView.Buffer())
	}
	for _, hidden := range []string{"@@ -1,2 +1,3 @@", "+another line"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected enter to hide %q from the collapsed file, actual %q", hidden, detailView.Buffer())
		}
	}
	if !strings.Contains(detailView.Buffer(), "internal/tui/model.go") || !strings.Contains(detailView.Buffer(), "+new model") {
		t.Fatalf("expected collapsing one file to keep the other file visible, actual %q", detailView.Buffer())
	}

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	collapseHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = collapseHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") || !strings.Contains(detailView.Buffer(), "+another line") {
		t.Fatalf("expected za to expand the file again, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected za to restore the collapsed file hunk, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenTheCursorOnAChangesDiffLine_WhenPressingEnterAndZA_ThenItTogglesTheContainingFileVisibility(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": given_reviewSessionPullRequestDiff()},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "+new line")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") {
		t.Fatalf("expected enter on a diff line to collapse the containing file, actual %q", detailView.Buffer())
	}
	for _, hidden := range []string{"@@ -1,2 +1,3 @@", "+another line"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected enter on a diff line to hide %q, actual %q", hidden, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go")

	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "+another line") {
		t.Fatalf("expected enter to re-expand the containing file, actual %q", detailView.Buffer())
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "+new line")

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	collapseHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = collapseHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+reviewDiffHeaderPathIcon+" internal/tui/render.go") {
		t.Fatalf("expected za on a diff line to collapse the containing file, actual %q", detailView.Buffer())
	}
	for _, hidden := range []string{"@@ -1,2 +1,3 @@", "+another line"} {
		if strings.Contains(detailView.Buffer(), hidden) {
			t.Fatalf("expected za on a diff line to hide %q, actual %q", hidden, detailView.Buffer())
		}
	}
	then_reviewModeDetailCursorLineContains(t, gui, subject, "internal/tui/render.go")
}

func TestBrowserMode_GivenAResolvedChangesTabThread_WhenPressingEnterAndZA_ThenItTogglesTheThreadVisibility(t *testing.T) {
	diff := githubcli.PullRequestDiff{
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
			ID:         "thread-1",
			IsResolved: true,
			Path:       "internal/tui/render.go",
			Line:       43,
			DiffSide:   "RIGHT",
			Comments: []githubcli.PullRequestComment{{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
				Body:      "Inline thread body",
				CreatedAt: "2026-04-18T10:30:00Z",
			}},
		}},
	}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Inline thread body": "Rendered inline thread body",
	}}
	gui := given_headlessGuiWithSize(t, 120, 50)
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headerLineIndex := given_viewLineIndexContaining(t, detailView, " internal/tui/render.go:43")
	if headerLineIndex < 0 {
		t.Fatalf("expected the resolved thread to start folded in changes, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered inline thread body") {
		t.Fatalf("expected the folded changes thread to hide its body, actual %q", detailView.Buffer())
	}
	if headerLineIndex > 0 && strings.HasPrefix(detailView.BufferLines()[headerLineIndex-1], "────") {
		t.Fatalf("expected the folded changes thread to drop the leading horizontal separator, actual %q", detailView.Buffer())
	}
	if headerLineIndex+1 < len(detailView.BufferLines()) && strings.HasPrefix(detailView.BufferLines()[headerLineIndex+1], "────") {
		t.Fatalf("expected the folded changes thread to drop the trailing horizontal separator, actual %q", detailView.Buffer())
	}
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "internal/tui/render.go:43")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " internal/tui/render.go:43") || !strings.Contains(detailView.Buffer(), "Rendered inline thread body") {
		t.Fatalf("expected enter to expand the changes thread, actual %q", detailView.Buffer())
	}
	headerLineIndex = given_viewLineIndexContaining(t, detailView, " internal/tui/render.go:43")
	if headerLineIndex > 0 && strings.HasPrefix(detailView.BufferLines()[headerLineIndex-1], "────") {
		t.Fatalf("expected the expanded changes thread to drop the leading horizontal separator, actual %q", detailView.Buffer())
	}
	if headerLineIndex+1 < len(detailView.BufferLines()) && strings.HasPrefix(detailView.BufferLines()[headerLineIndex+1], "────") {
		t.Fatalf("expected the expanded changes thread to drop the trailing horizontal separator, actual %q", detailView.Buffer())
	}

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	collapseHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = collapseHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " internal/tui/render.go:43") {
		t.Fatalf("expected za to collapse the changes thread, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered inline thread body") {
		t.Fatalf("expected za to hide the changes thread body again, actual %q", detailView.Buffer())
	}
}

func TestLayout_GivenFailedOverviewSections_WhenRendering_ThenOnlyFailedBlocksStartExpanded(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:            "First PR",
				Number:           42,
				Body:             "Body 42",
				BaseRefName:      "main",
				HeadRefName:      "feature/failed-overview",
				State:            "OPEN",
				Mergeable:        "MERGEABLE",
				MergeStateStatus: "CLEAN",
				Reviews: []githubcli.PullRequestReview{{
					Author:      &githubcli.PullRequestCommentAuthor{Login: "reviewer-approved"},
					State:       "APPROVED",
					SubmittedAt: "2026-04-21T10:00:00Z",
				}},
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{{
					Name:         "test",
					WorkflowName: "CI",
					Status:       "COMPLETED",
					Conclusion:   "FAILURE",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 42": "Rendered body 42"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewSuccessIcon+" Reviewers (1/1)") {
		t.Fatalf("expected non-failed reviewers to stay folded, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "@reviewer-approved") {
		t.Fatalf("expected the folded reviewers block to hide its body, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewFailureIcon+" Merge Checks") || !strings.Contains(detailView.Buffer(), "1 reviewer has approved.") || !strings.Contains(detailView.Buffer(), "1 failing") {
		t.Fatalf("expected failed merge checks to start expanded, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewFailureIcon+" Builds") || !strings.Contains(detailView.Buffer(), "CI / test (Failed)") {
		t.Fatalf("expected failed builds to start expanded, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenLinkedAndPendingBuilds_WhenRendering_ThenOnlyTheLinkedBuildIsUnderlined(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/underline-builds",
				State:       "OPEN",
				StatusCheckRollup: []githubcli.PullRequestStatusCheck{
					{Name: "test", WorkflowName: "CI", Status: "COMPLETED", Conclusion: "FAILURE", Link: "https://github.com/acme/widgets/actions/runs/42"},
					{Name: "deploy", Status: "IN_PROGRESS"},
				},
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	linkedLineIndex := given_viewLineIndexContaining(t, detailView, "CI / test (Failed)")
	pendingLineIndex := given_viewLineIndexContaining(t, detailView, "deploy (Pending)")
	then_viewLineSegmentIsUnderlined(t, gui, viewDetailName, linkedLineIndex, "CI / test (Failed)")
	then_viewLineSegmentIsNotUnderlined(t, gui, viewDetailName, pendingLineIndex, "deploy (Pending)")
}

func TestBrowserMode_GivenTheCursorOnAnOverviewSection_WhenPressingEnter_ThenItTogglesTheSectionVisibility(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				BaseRefName:    "main",
				HeadRefName:    "feature/overview",
				State:          "OPEN",
				ReviewRequests: []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
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
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, " "+pullRequestOverviewPendingIcon+" Reviewers (0/1)")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewPendingIcon+" Reviewers (0/1)") {
		t.Fatalf("expected enter to collapse the overview section, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "@reviewer-requested") {
		t.Fatalf("expected the collapsed overview section to hide its body, actual %q", detailView.Buffer())
	}

	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " "+pullRequestOverviewPendingIcon+" Reviewers (0/1)") || !strings.Contains(detailView.Buffer(), "@reviewer-requested") {
		t.Fatalf("expected enter to expand the overview section again, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenTheCursorOnAConversation_WhenPressingZA_ThenItTogglesTheConversationVisibility(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "Comment body",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Comment body": "Rendered comment body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, " Comment")

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment") {
		t.Fatalf("expected za to collapse the conversation, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered comment body") {
		t.Fatalf("expected the collapsed conversation to hide its body, actual %q", detailView.Buffer())
	}

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment") || !strings.Contains(detailView.Buffer(), "Rendered comment body") {
		t.Fatalf("expected za to expand the conversation again, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenALongPullRequestCommentCodeBlockLineWithErrorTokens_WhenRenderingComments_ThenItDoesNotInsertBlankLinesBetweenWrappedCodeSegments(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
					Body:      "(Fix with Cursor)\n\n```\nApplicationContext failure threshold (1) exceeded: skipping repeated attempt to load context for [WebMergedContextConfiguration@31946dd4 testClass = foo.bar.ApplicationIntegrationTest, locations = [], classes = [foo.bar.Application, foo.bar.ApplicationIntegrationTest.CustomRestTemplateBuilderConfig], contextInitializerClasses = [], activeProfiles = [\"test\"], propertySourceDescriptors = [], propertySourceProperties = [\"org.springframework.boot.test.context.SpringBootTestContextBootstrapper=true\", \"server.port=0\"], contextCustomizers = [org.springframework.boot.test.context.filter.ExcludeFilterContextCustomizer@565983f3]]\n```",
					CreatedAt: "2026-04-18T10:00:00Z",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 80, 40)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	firstCodeLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "ApplicationContext failure threshold")
	lastCodeLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "ExcludeFilterContextCustomizer@565983f3]]")
	actualBlankLineCountInsideCodeBlock := 0
	for lineIndex := firstCodeLineIndex + 1; lineIndex < lastCodeLineIndex; lineIndex++ {
		if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[lineIndex])); actualInnerText == "" {
			actualBlankLineCountInsideCodeBlock++
		}
	}
	if actualBlankLineCountInsideCodeBlock != 0 {
		t.Fatalf("expected the comments tab code block to wrap without interior blank lines, actual %d in %q", actualBlankLineCountInsideCodeBlock, detailView.Buffer())
	}
}

func TestBrowserMode_GivenAnInlineCommentCodeFence_WhenRenderingComments_ThenItDoesNotAddExtraPaddingLinesInsideTheCommentBox(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				InlineComments: []githubcli.PullRequestInlineComment{{
					Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
					Body:         "```go\nfunc render(value int) string {\n\treturn fmt.Sprintf(\"%d\", value + 42)\n}\n```",
					CreatedAt:    "2026-04-18T10:00:00Z",
					Path:         "internal/tui/render.go",
					Line:         43,
					OriginalLine: 43,
					Side:         "RIGHT",
					DiffHunk:     "@@ -42,1 +43,3 @@\n+func render(value int) string {\n+\treturn fmt.Sprintf(\"%d\", value + 42)\n+}",
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	metadataLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "@reviewer-inline")
	codeStartLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "func render")
	codeEndLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "}")
	if metadataLineIndex+1 != codeStartLineIndex {
		t.Fatalf("expected the code fence to start immediately after the metadata line, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.BufferLines()[codeEndLineIndex+1], "╰") {
		t.Fatalf("expected the code fence to end immediately before the bottom border, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenInlineThreadReplyCodeFence_WhenRenderingChanges_ThenItDoesNotAddExtraPaddingLinesInsideTheReplyCommentBox(t *testing.T) {
	diff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\nindex 0000000..1111111 100644\n--- a/internal/tui/render.go\n+++ b/internal/tui/render.go\n@@ -42,0 +43,3 @@\n+func render(value int) string {\n+\treturn fmt.Sprintf(\"%d\", value + 42)\n+}\n"}
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     45,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:      "Root comment",
				CreatedAt: "2026-04-18T10:00:00Z",
				DiffHunk:  "@@ -42,0 +43,3 @@\n+func render(value int) string {\n+\treturn fmt.Sprintf(\"%d\", value + 42)\n+}",
			},
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				Body:      "```go\nfunc render(value int) string {\n\treturn fmt.Sprintf(\"%d\", value + 42)\n}\n```",
				CreatedAt: "2026-04-18T10:30:00Z",
			},
		},
	}}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGuiWithSize(t, 120, 50)
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	metadataLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "@octocat")
	codeStartLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "func render")
	codeEndLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "}")
	if metadataLineIndex+1 != codeStartLineIndex {
		t.Fatalf("expected the reply code fence to start immediately after the metadata line, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.BufferLines()[codeEndLineIndex+1], "╰") {
		t.Fatalf("expected the reply code fence to end immediately before the bottom border, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenInlineThreadReplies_WhenRenderingChanges_ThenItOmitsTheReplySpacerAndLabel(t *testing.T) {
	diff := githubcli.PullRequestDiff{UnifiedDiff: "diff --git a/internal/tui/render.go b/internal/tui/render.go\nindex 0000000..1111111 100644\n--- a/internal/tui/render.go\n+++ b/internal/tui/render.go\n@@ -42,1 +43,1 @@\n-context line\n+new line\n"}
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     43,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
				Body:      "Root comment",
				CreatedAt: "2026-04-18T10:00:00Z",
				DiffHunk:  "@@ -42,1 +43,1 @@\n-context line\n+new line",
			},
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				Body:      "First reply",
				CreatedAt: "2026-04-18T10:30:00Z",
			},
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "maintainer"},
				Body:      "Last reply",
				CreatedAt: "2026-04-18T10:45:00Z",
			},
		},
	}}
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/changes",
				State:       "OPEN",
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Root comment": "Rendered root comment",
		"First reply":  "Rendered first reply",
		"Last reply":   "Rendered last reply",
	}}
	gui := given_headlessGuiWithSize(t, 120, 50)
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if strings.Contains(detailView.Buffer(), "Reply") {
		t.Fatalf("expected changes-tab inline replies to omit the standalone reply label, actual %q", detailView.Buffer())
	}
	firstReplyMetadataLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "@octocat")
	lastReplyMetadataLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "@maintainer")
	if !strings.HasPrefix(detailView.BufferLines()[firstReplyMetadataLineIndex-1], "├─╭") {
		t.Fatalf("expected the first reply to start on the connector line, actual %q", detailView.Buffer())
	}
	if !strings.HasPrefix(detailView.BufferLines()[lastReplyMetadataLineIndex-1], "╰─╭") {
		t.Fatalf("expected the last reply to start on the closing connector line, actual %q", detailView.Buffer())
	}
}

func TestBrowserMode_GivenInlineThreadConversations_WhenRendering_ThenItCollapsesResolvedThreadsKeepsSectionsTightAndShowsTheDiffPreview(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/conversations",
				State:       "OPEN",
				InlineCommentThreads: []githubcli.PullRequestReviewThread{
					{
						ID:         "thread-1",
						IsResolved: true,
						Path:       "internal/tui/render.go",
						Line:       43,
						DiffSide:   "RIGHT",
						Comments: []githubcli.PullRequestComment{{
							Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-resolved"},
							Body:      "Resolved thread body",
							CreatedAt: "2026-04-18T10:00:00Z",
							DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
						}},
					},
					{
						ID:       "thread-2",
						Path:     "internal/tui/model.go",
						Line:     60,
						DiffSide: "RIGHT",
						Comments: []githubcli.PullRequestComment{{
							Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-active"},
							Body:      "Active thread body",
							CreatedAt: "2026-04-18T10:30:00Z",
							DiffHunk:  "@@ -56,6 +56,7 @@\n one\n two\n three\n four\n-old value\n+new value\n tail",
						}},
					},
				},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{
		"Resolved thread body": "Rendered resolved thread body",
		"Active thread body":   "Rendered active thread body",
	}}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	resolvedHeaderLineIndex := given_viewLineIndexContaining(t, detailView, " internal/tui/render.go:43")
	activeHeaderLineIndex := given_viewLineIndexContaining(t, detailView, " internal/tui/model.go:60")
	if activeHeaderLineIndex != resolvedHeaderLineIndex+1 {
		t.Fatalf("expected the collapsed thread header to sit immediately before the next thread header without an outer border block, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered resolved thread body") {
		t.Fatalf("expected resolved threads to stay folded by default, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered active thread body") {
		t.Fatalf("expected unresolved threads to stay visible, actual %q", detailView.Buffer())
	}
	for _, expected := range []string{"@@ -56,6 +56,7 @@", "56 : 56 │ one", "57 : 57 │ two", "58 : 58 │ three", "59 : 59 │ four", "60 :    │ old value", "   : 60 │ new value"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the browser conversations tab to keep the diff preview line %q, actual %q", expected, detailView.Buffer())
		}
	}
	if strings.Contains(detailView.Buffer(), "61 : 61 │ tail") {
		t.Fatalf("expected the browser conversations tab to crop the diff preview after five lines plus the selected line, actual %q", detailView.Buffer())
	}
	if activeHeaderLineIndex == 0 || strings.HasPrefix(detailView.BufferLines()[activeHeaderLineIndex-1], "────") {
		t.Fatalf("expected the browser conversations tab to drop the outer thread separator, actual %q", detailView.Buffer())
	}
	metadataLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-active")
	if strings.Contains(detailView.BufferLines()[metadataLineIndex], "Unresolved") {
		t.Fatalf("expected the unresolved state to move off the comment metadata line, actual %q", detailView.BufferLines()[metadataLineIndex])
	}
}

func TestLayout_GivenOverviewHeaderMetadata_WhenRendering_ThenTheOverviewTabShowsColoredChurnAndLifecycleTimestamps(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 118, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body", UpdatedAt: "2026-04-18T12:30:00Z"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#118": {Title: "Styled PR", Number: 118, Body: "Body 118", Author: &githubcli.PullRequestAuthor{Login: "octocat"}, BaseRefName: "main", HeadRefName: "feature-118", State: "OPEN", CreatedAt: "2026-04-18T10:00:00Z", UpdatedAt: "2026-04-18T12:30:00Z", Additions: 12, Deletions: 3, ChangedFiles: 5},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 118": "Rendered body 118"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for _, expected := range []string{"Created by", "the 2026-04-18 10:00 UTC", "(last updated at 2026-04-18 12:30 UTC)", "+12", "-3", "Rendered body 118"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected detail buffer to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	countsLineIndex := given_viewLineIndexContaining(t, detailView, "+12")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, countsLineIndex, "+12", given_themeColorHex(t, theme.DiffAdditionHex), "description header additions")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, countsLineIndex, "-3", given_themeColorHex(t, theme.DiffDeletionHex), "description header deletions")
}

func TestLayout_GivenInlineCommentDiff_WhenRendering_ThenTheCommentsTabUsesTreeSitterSyntaxColorsAndExactChangeBackgrounds(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 109, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#109": {Title: "Styled PR", Number: 109, Body: "Body 109", BaseRefName: "main", HeadRefName: "feature-109", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "Inline diff body", CreatedAt: "2026-04-18T10:00:00Z", Path: "src/main/java/com/acme/VersionParser.java", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -43,1 +43,1 @@\n-return Versions.fromString(\"5.0.1\");\n+return Versions.fromString(\"5.1.0\");"}}},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 109": "Rendered body 109", "Inline diff body": "Rendered inline diff body"}}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	locationLineIndex := given_viewLineIndexContaining(t, detailView, detailInlineCommentLocationIcon+" src/main/java/com/acme/VersionParser.java:43")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, locationLineIndex, "+1", given_themeColorHex(t, theme.DiffAdditionHex), "inline addition count")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, locationLineIndex, "-1", given_themeColorHex(t, theme.DiffDeletionHex), "inline deletion count")

	deletionLineIndex := given_viewLineIndexContaining(t, detailView, `return Versions.fromString("5.0.1");`)
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, "return", given_themeColorHex(t, theme.SyntaxKeywordHex), "inline deletion keyword")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, "fromString", given_themeColorHex(t, theme.SyntaxFunctionHex), "inline deletion method")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, `"5.0.1"`, given_themeColorHex(t, theme.SyntaxStringHex), "inline deletion string")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, "0.1", given_themeColorHex(t, theme.DiffDeletionHighlightBackgroundHex), "inline deletion changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, `return Versions.fromString("5.`, given_themeColorHex(t, theme.DiffDeletionBackgroundHex), "inline deletion unchanged background")

	additionLineIndex := given_viewLineIndexContaining(t, detailView, `return Versions.fromString("5.1.0");`)
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, "return", given_themeColorHex(t, theme.SyntaxKeywordHex), "inline addition keyword")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, "fromString", given_themeColorHex(t, theme.SyntaxFunctionHex), "inline addition method")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, `"5.1.0"`, given_themeColorHex(t, theme.SyntaxStringHex), "inline addition string")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, "1.0", given_themeColorHex(t, theme.DiffAdditionHighlightBackgroundHex), "inline addition changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, `return Versions.fromString("5.`, given_themeColorHex(t, theme.DiffAdditionBackgroundHex), "inline addition unchanged background")
}

func TestLayout_GivenInlineComments_WhenRendering_ThenTheCommentsTabUsesAHighlightedAuthorBadgeInsideTheCommentBox(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 110, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#110": {Title: "Styled PR", Number: 110, Body: "Body 110", BaseRefName: "main", HeadRefName: "feature-110", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "Inline body", CreatedAt: "2026-04-18T10:00:00Z", Path: "internal/tui/render.go", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\","}}},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 110": "Rendered body 110", "Inline body": "Rendered inline body"}}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	authorLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-inline")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-inline", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "inline comment author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-inline", given_themeColorHex(t, theme.CommentAuthorBadgeHex), "inline comment author badge foreground")
	if !strings.Contains(detailView.BufferLines()[authorLineIndex], "2026-04-18 10:00 UTC") {
		t.Fatalf("expected the inline comment timestamp to stay on the metadata line, actual %q", detailView.BufferLines()[authorLineIndex])
	}
}

func TestLayout_GivenSuggestionFenceInlineComment_WhenRendering_ThenTheCommentsTabShowsTheCurrentLineBeforeTheSuggestedReplacementAndFillsTheCommentBoxInterior(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 115, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#115": {Title: "Styled PR", Number: 115, Body: "Body 115", BaseRefName: "main", HeadRefName: "feature-115", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "```suggestion\nfmt.Println(\"bonjour\")\n```", CreatedAt: "2026-04-18T10:00:00Z", Path: "internal/tui/render.go", Line: 43, OriginalLine: 43, Side: "RIGHT", DiffHunk: "@@ -43,1 +43,1 @@\n-fmt.Println(\"goodbye\")\n+fmt.Println(\"hello\")"}}},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	labelLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "Suggestion")
	removedLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, `fmt.Println("hello")`)
	addedLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, `fmt.Println("bonjour")`)
	if removedLineIndex >= addedLineIndex {
		t.Fatalf("expected the removed suggestion line to render before the added line, actual %q", detailView.Buffer())
	}
	if labelLineIndex+1 != removedLineIndex {
		t.Fatalf("expected the suggestion diff to start immediately after the label, actual %q", detailView.Buffer())
	}
	if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[removedLineIndex])); !strings.Contains(actualInnerText, `-fmt.Println("hello")`) {
		t.Fatalf("expected the suggestion block to show the current line %q, actual %q", `-fmt.Println("hello")`, actualInnerText)
	}
	if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[addedLineIndex])); !strings.Contains(actualInnerText, `+fmt.Println("bonjour")`) {
		t.Fatalf("expected the suggestion block to show the suggested line %q, actual %q", `+fmt.Println("bonjour")`, actualInnerText)
	}
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, removedLineIndex, "hello", given_themeColorHex(t, theme.DiffDeletionHighlightBackgroundHex), "inline suggestion deletion changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, removedLineIndex, `fmt.Println("`, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion deletion base background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, addedLineIndex, "bonjour", given_themeColorHex(t, theme.DiffAdditionHighlightBackgroundHex), "inline suggestion addition changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, addedLineIndex, `fmt.Println("`, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion addition base background")
	then_viewCommentBoxBorderDoesNotHaveBackgroundColor(t, gui, viewDetailName, addedLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "inline suggestion code block border background")
}

func TestLayout_GivenALongSuggestionFenceInlineComment_WhenRendering_ThenTheCommentsTabWrapsTheSuggestionInsideTheCommentBoxWithoutAGapAfterTheLabel(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 116, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#116": {Title: "Styled PR", Number: 116, Body: "Body 116", BaseRefName: "main", HeadRefName: "feature-116", State: "OPEN", InlineComments: []githubcli.PullRequestInlineComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "```suggestion\n- [ ] 21.10 Rename `observability` packages → `foo.bar.observability.*`; rename `s3-assets` → `foo.bar.s3assets.*`; rename `scheduled-jobs` → `foo.bar.scheduled_jobs.*`; rename `http-clients` → `foo.bar.http_clients.*`; update all imports repo-wide\n```", CreatedAt: "2026-04-18T10:00:00Z", Path: "openspec/changes/refactor-shared-modules/tasks.md", Line: 218, OriginalLine: 218, Side: "RIGHT", DiffHunk: "@@ -0,0 +218,1 @@\n+- [ ] 21.10 Rename `observability` packages → `foo.bar.observability.*`; rename `infrastructure/s3-assets` → `foo.bar.s3_assets.*`; rename `scheduled-jobs` → `foo.bar.scheduled_jobs.*`; rename `http-clients` → `foo.bar.http_clients.*`; update all imports repo-wide"}}},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 80, 40)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	labelLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "Suggestion")
	firstDeletionLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "-- [ ] 21.10 Rename")
	continuedDeletionLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "s3_assets.*")
	firstAdditionLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "+- [ ] 21.10 Rename")
	continuedAdditionLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "s3assets.*")
	if labelLineIndex+1 != firstDeletionLineIndex {
		t.Fatalf("expected the wrapped suggestion diff to start immediately after the label, actual %q", detailView.Buffer())
	}
	if firstDeletionLineIndex >= continuedDeletionLineIndex || continuedDeletionLineIndex >= firstAdditionLineIndex || firstAdditionLineIndex >= continuedAdditionLineIndex {
		t.Fatalf("expected the long suggestion to wrap across multiple visible lines in order, actual %q", detailView.Buffer())
	}
}

func TestLayout_GivenMarkdownDescriptionAndComments_WhenRendering_ThenTheDetailPaneShowsAStyledHeadingGreyCommentBorderAndHighlightedCommentAuthorBadge(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 110, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#110": {Title: "Styled PR", Number: 110, Body: "## Why\n\nParagraph body", BaseRefName: "main", HeadRefName: "feature-110", State: "OPEN", Comments: []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Ship it", CreatedAt: "2026-04-18T10:00:00Z"}}},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headingLineIndex := given_viewLineIndexContaining(t, detailView, "Why")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, headingLineIndex, "Why", given_themeColorHex(t, theme.MarkdownHeadingHex), "markdown heading")

	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	commentBorderLineIndex := given_viewLineIndexContaining(t, detailView, "╭")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, commentBorderLineIndex, "╭", given_themeColorHex(t, theme.InactiveBorderHex), "comment border")
	authorLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-one")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "comment author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.CommentAuthorBadgeHex), "comment author badge foreground")
	if !strings.Contains(detailView.BufferLines()[authorLineIndex], "2026-04-18 10:00 UTC") {
		t.Fatalf("expected the comment timestamp to stay on the metadata line, actual %q", detailView.BufferLines()[authorLineIndex])
	}
}

func TestLayout_GivenConnectedUserAndReviewerComments_WhenRenderingCommentsTab_ThenOnlyTheViewerKeepsTheHighlightedAuthorBadge(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 111, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#111": {
				Title:       "Styled PR",
				Number:      111,
				Body:        "Body 111",
				BaseRefName: "main",
				HeadRefName: "feature-111",
				State:       "OPEN",
				Comments: []githubcli.PullRequestComment{
					{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, Body: "Needs changes", CreatedAt: "2026-04-18T10:00:00Z"},
					{Author: &githubcli.PullRequestCommentAuthor{Login: "octocat"}, Body: "Already fixed", CreatedAt: "2026-04-18T10:30:00Z"},
				},
			},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.connectedUserLogin = "octocat"
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 111": "Rendered body 111", "Needs changes": "Rendered reviewer comment", "Already fixed": "Rendered viewer comment"}}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.nextDetailTab(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	reviewerLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-one")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, reviewerLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.PendingBackgroundHex), "reviewer author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, reviewerLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.PendingHex), "reviewer author badge foreground")
	viewerLineIndex := given_viewLineIndexContaining(t, detailView, "@octocat")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, viewerLineIndex, detailCommentsIcon+" @octocat", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "viewer author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, viewerLineIndex, detailCommentsIcon+" @octocat", given_themeColorHex(t, theme.CommentAuthorBadgeHex), "viewer author badge foreground")
}

func TestLayout_GivenMarkdownHeading_WhenRendering_ThenItsBackgroundFillsTheWholeVisibleLine(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 112, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#112": {Title: "Styled PR", Number: 112, Body: "# Ship it", BaseRefName: "main", HeadRefName: "feature-112", State: "OPEN"},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headingLineIndex := given_viewLineIndexContaining(t, detailView, "Ship it")
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, headingLineIndex, given_themeColorHex(t, theme.MarkdownHeadingBackgroundHex), "markdown heading background")
}

func TestLayout_GivenMarkdownCodeBlock_WhenRendering_ThenItsBackgroundFillsTheWholeVisibleLine(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 113, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#113": {Title: "Styled PR", Number: 113, Body: "```go\nfmt.Println(\"hi\")\n```", BaseRefName: "main", HeadRefName: "feature-113", State: "OPEN"},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	codeLineIndex := given_viewLineIndexContaining(t, detailView, `fmt.Println("hi")`)
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, codeLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "markdown code block background")
}

func TestLayout_GivenMarkdownCodeBlock_WhenRendering_ThenItAddsBlankBackgroundLinesAboveAndBelowTheCode(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 114, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#114": {Title: "Styled PR", Number: 114, Body: "```go\nfmt.Println(\"hi\")\n```", BaseRefName: "main", HeadRefName: "feature-114", State: "OPEN"},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGuiWithSize(t, 120, 50)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	codeLineIndex := given_viewLineIndexContaining(t, detailView, `fmt.Println("hi")`)
	if codeLineIndex < 2 || codeLineIndex >= len(detailView.BufferLines())-2 {
		t.Fatalf("expected blank outer spacing and padding lines around the code block, actual lines %q", detailView.BufferLines())
	}
	for _, lineIndex := range []int{codeLineIndex - 2, codeLineIndex - 1, codeLineIndex + 1, codeLineIndex + 2} {
		if strings.Trim(detailView.BufferLines()[lineIndex], " ⠀") != "" {
			t.Fatalf("expected the code block spacer line %d to stay blank, actual %q", lineIndex, detailView.BufferLines()[lineIndex])
		}
	}
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, codeLineIndex-1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "markdown code block top padding background")
	then_viewLineHasBackgroundColor(t, gui, viewDetailName, codeLineIndex+1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "markdown code block bottom padding background")
	actualDocument := subject.currentDetailDocument(detailView)
	for _, lineIndex := range []int{codeLineIndex - 2, codeLineIndex + 2} {
		if len(actualDocument.lineStylePrefixes[lineIndex]) != 0 {
			t.Fatalf("expected the outer spacing line %d to keep the default background, actual prefixes %q", lineIndex, actualDocument.lineStylePrefixes[lineIndex])
		}
	}
}

func TestLayout_GivenRenderedMarkdownDescription_WhenRendering_ThenVisibleLinesDoNotLeakRawANSISequences(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Styled PR", Number: 111, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#111": {Title: "Styled PR", Number: 111, Body: "No need to deploy the docker image and publish the doc to S3.\n\nFor now, we are just testing deploying the runtime on GCP, not AWS.", BaseRefName: "main", HeadRefName: "feature-111", State: "OPEN"},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for _, actualLine := range detailView.BufferLines() {
		if strings.Contains(actualLine, "[38;") || strings.Contains(actualLine, "[3;") || strings.Contains(actualLine, "[0m") {
			t.Fatalf("expected the visible detail lines to hide ANSI sequences, actual %q", actualLine)
		}
	}
}

func TestLayout_GivenAnotherSelectedPullRequestAfterScrolling_WhenRendering_ThenTheDetailOriginResets(t *testing.T) {
	longRenderedBody := strings.TrimSpace(strings.Repeat("rendered detail line\n", 80))
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 201, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 201"}),
		myPullRequestRow(githubcli.PullRequest{Title: "Second PR", Number: 202, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 202"}),
	})
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#201": {Title: "First PR", Number: 201, Body: "Body 201", BaseRefName: "main", HeadRefName: "feature-201", State: "OPEN"},
			"acme/widgets#202": {Title: "Second PR", Number: 202, Body: "Body 202", BaseRefName: "main", HeadRefName: "feature-202", State: "OPEN"},
		},
	}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.asyncRunner = inlineAsyncRunner{}
	subject.uiUpdater = immediateUIUpdater{}
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 201": longRenderedBody, "Body 202": longRenderedBody}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.openDetail(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	for range detailView.InnerHeight() + 4 {
		actualErr = subject.moveSelectionDown(gui, detailView)
		then_noError(t, actualErr)
	}
	_, originY := detailView.Origin()
	if originY < 1 {
		t.Fatalf("expected detail origin to move down, actual %d", originY)
	}

	actualErr = subject.closeDetail(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)
	actualErr = subject.layout(gui)
	then_noError(t, actualErr)

	_, originY = detailView.Origin()
	if originY != 0 {
		t.Fatalf("expected detail origin to reset to 0, actual %d", originY)
	}
	cursorX, cursorY := detailView.Cursor()
	if cursorX != 0 || cursorY != 0 {
		t.Fatalf("expected detail cursor 0,0 after resetting, actual %d,%d", cursorX, cursorY)
	}
}

func TestRefreshViews_GivenInvalidatedPullRequestDetail_WhenGhHasNotReturnedYet_ThenTheDetailPaneShowsALoadingState(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 301, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "fallback body"}),
	})
	loader := &fakePullRequestDetailLoader{}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	subject.pullRequestDetailCache["acme/widgets#301"] = pullRequestDetailResult{detail: githubcli.ToDomainPullRequestDetail(githubcli.PullRequestDetail{Title: "First PR", Number: 301, Body: "Cached detail body"})}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	subject.invalidatePullRequestDetail("acme/widgets", 301)
	actualErr = subject.refreshViews(gui)
	then_noError(t, actualErr)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued detail reload, actual %d", len(asyncRunner.runs))
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), string(loadingSpinnerFrames[0])) {
		t.Fatalf("expected detail body to show spinner %q, actual %q", string(loadingSpinnerFrames[0]), detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), pullRequestDetailLoadingTitle) {
		t.Fatalf("expected detail body to hide %q, actual %q", pullRequestDetailLoadingTitle, detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Running `gh pr view 301 -R acme/widgets --json ...`.") {
		t.Fatalf("expected detail body to keep only the spinner, actual %q", detailView.Buffer())
	}

	statusView, actualErr := gui.View("status-line")
	then_noError(t, actualErr)
	expectedStatus := string(loadingSpinnerFrames[0]) + " Running `gh pr view 301 -R acme/widgets --json ...`."
	if actual := strings.TrimSpace(statusView.Buffer()); actual != expectedStatus {
		t.Fatalf("expected status line %q, actual %q", expectedStatus, actual)
	}
	then_viewDoesNotExist(t, gui, viewDetailFooterName)
}

func TestReloadActivePullRequestsTab_GivenExistingPullRequests_WhenGhHasNotReturnedYet_ThenTheStatusLineShowsTheLoadingCommand(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequests(MyPullRequestsTab, []Item{{Title: "my-pr-1", Detail: "body-1"}, {Title: "my-pr-2", Detail: "body-2"}})
	loader := &fakePullRequestDetailLoader{}
	asyncRunner := &capturingAsyncRunner{}
	subject := given_programWithTestGitHubDeps(model, loader)
	subject.connectedUserLoadStarted = true
	subject.myPullRequestsLoadStarted = true
	subject.requestedPullRequestsLoadStarted = true
	subject.notificationsLoadStarted = true
	subject.asyncRunner = asyncRunner
	subject.uiUpdater = immediateUIUpdater{}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	subject.reloadActivePullRequestsTab(gui)

	if len(asyncRunner.runs) != 1 {
		t.Fatalf("expected one queued pull request reload, actual %d", len(asyncRunner.runs))
	}

	pullRequestsView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(pullRequestsView.Buffer(), "my-pr-1") {
		t.Fatalf("expected the existing pull request list to stay visible, actual %q", pullRequestsView.Buffer())
	}

	statusView, actualErr := gui.View("status-line")
	then_noError(t, actualErr)
	expectedStatus := string(loadingSpinnerFrames[0]) + " " + myPullRequestsLoadingDetail
	if actual := strings.TrimSpace(statusView.Buffer()); actual != expectedStatus {
		t.Fatalf("expected status line %q, actual %q", expectedStatus, actual)
	}
	then_statusLineKeyHintsAre(t, gui, "?: help, /: search, a: action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func given_viewLineIndexContaining(t *testing.T, view *gocui.View, segment string) int {
	t.Helper()

	for lineIndex, line := range view.BufferLines() {
		if strings.Contains(line, segment) {
			return lineIndex
		}
	}

	t.Fatalf("expected a detail line containing %q, actual %q", segment, strings.Join(view.BufferLines(), "\n"))
	return -1
}

type fakePullRequestDetailLoader struct {
	details                           map[string]githubcli.PullRequestDetail
	detailErrors                      map[string]error
	detailCalls                       []string
	listPullRequestCommands           [][]string
	listPullRequestsErr               error
	diffs                             map[string]githubcli.PullRequestDiff
	diffErrors                        map[string]error
	diffCalls                         []string
	fileTeamOwners                    map[string]map[string][]string
	fileTeamOwnerErrors               map[string]error
	fileTeamOwnerCalls                []string
	fileTeamOwnerPaths                [][]string
	commentCalls                      []string
	commentBodies                     []string
	commentErr                        error
	updatePullRequestCommentIDs       []string
	updatePullRequestCommentBodies    []string
	updatePullRequestCommentErr       error
	deletePullRequestCommentIDs       []string
	deletePullRequestCommentErr       error
	myPullRequests                    []githubcli.PullRequest
	requestedPullRequests             []githubcli.PullRequest
	notifications                     []githubcli.Notification
	notificationsErr                  error
	markNotificationReadIDs           []string
	markNotificationReadErr           error
	markNotificationDoneIDs           []string
	markNotificationDoneErr           error
	markAllNotificationsReadCalls     int
	markAllNotificationsReadAccepted  bool
	markAllNotificationsReadErr       error
	markAllNotificationsReadPollLoads int
	markAllNotificationsDoneIDs       [][]string
	markAllNotificationsDoneErr       error
	issueDetails                      map[string]githubcli.IssueDetail
	issueDetailErrors                 map[string]error
	issueDetailCalls                  []string
	releaseDetails                    map[string]githubcli.ReleaseDetail
	releaseDetailErrors               map[string]error
	releaseDetailCalls                []string
	approveCalls                      []string
	approveErr                        error
	reviewCommentCalls                []string
	reviewCommentBodies               []string
	reviewCommentErr                  error
	requestChangesCalls               []string
	requestChangesBodies              []string
	requestChangesErr                 error
	submitReviewIDs                   []string
	submitReviewEvents                []githubcli.PullRequestReviewEvent
	submitReviewBodies                []string
	submitReviewErr                   error
	reviewThreadReviewIDs             []string
	reviewThreadBodies                []string
	reviewThreadTargets               []githubcli.PullRequestReviewThreadTarget
	reviewThreadErr                   error
	reviewThreadReplyReviewIDs        []string
	reviewThreadReplyThreadIDs        []string
	reviewThreadReplyBodies           []string
	reviewThreadReplyErr              error
	updateReviewCommentIDs            []string
	updateReviewCommentBodies         []string
	updateReviewCommentErr            error
	deleteReviewCommentIDs            []string
	deleteReviewCommentErr            error
	resolveReviewThreadIDs            []string
	resolveReviewThreadErr            error
	unresolveReviewThreadIDs          []string
	unresolveReviewThreadErr          error
	addReactionSubjectIDs             []string
	addReactionContents               []githubcli.ReactionContent
	addReactionErr                    error
	removeReactionSubjectIDs          []string
	removeReactionContents            []githubcli.ReactionContent
	removeReactionErr                 error
	reviewKeyByPendingID              map[string]string
	getPendingReviewCalls             []string
	getPendingReviewErr               error
	deletePullRequestReviewIDs        []string
	deletePullRequestReviewErr        error
	openBrowserCalls                  []string
	openBrowserErr                    error
	assignableUsers                   map[string][]githubcli.PullRequestAuthor
	assignableUserCalls               []string
	assignableUserErr                 error
	searchAssignableUsers             map[string][]githubcli.PullRequestAuthor
	searchAssignableUserCalls         []string
	searchAssignableUserErr           error
	updateAssigneeCalls               []string
	updateAssigneeAdditions           [][]string
	updateAssigneeRemovals            [][]string
	updateAssigneeErr                 error
	requestReviewerCalls              []string
	requestReviewerLogins             []string
	requestReviewerErr                error
	editTitleCalls                    []string
	editTitleValues                   []string
	editTitleErr                      error
	editDescriptionCalls              []string
	editDescriptionBodies             []string
	editDescriptionErr                error
	markReadyForReviewCalls           []string
	markReadyForReviewErr             error
	convertToDraftCalls               []string
	convertToDraftErr                 error
	closePullRequestCalls             []string
	closePullRequestErr               error
	reopenPullRequestCalls            []string
	reopenPullRequestErr              error
	squashMergeCalls                  []string
	squashMergeErr                    error
	enableAutoMergeCalls              []string
	enableAutoMergeErr                error
	disableAutoMergeCalls             []string
	disableAutoMergeErr               error
	updateBranchCalls                 []string
	updateBranchErr                   error
	startReviewCalls                  []string
	startReviewID                     string
	startReviewErr                    error
	buildRuns                         map[string]string
	buildRunJobs                      map[string][]githubcli.PullRequestBuildRunJob
	buildLogs                         map[int]string
	buildRunCalls                     []string
	buildRunChecks                    []githubcli.PullRequestStatusCheck
	buildRunJobCalls                  []string
	buildRunJobChecks                 []githubcli.PullRequestStatusCheck
	buildLogCalls                     []int
	buildRunErr                       error
	buildRunJobsErr                   error
	buildLogErr                       error
	renderedMarkdownHTML              map[string]string
	renderMarkdownHTMLCalls           []string
	renderMarkdownHTMLErr             error
	authToken                         string
	authTokenCalls                    int
	authTokenErr                      error
}

func (loader *fakePullRequestDetailLoader) GetConnectedUser() (githubdomain.ConnectedUser, error) {
	return githubdomain.ConnectedUser{}, nil
}

func (loader *fakePullRequestDetailLoader) ListPullRequests(commandArguments []string) ([]githubdomain.PullRequestSummary, error) {
	loader.listPullRequestCommands = append(loader.listPullRequestCommands, append([]string(nil), commandArguments...))
	if loader.listPullRequestsErr != nil {
		return nil, loader.listPullRequestsErr
	}
	if slices.Contains(commandArguments, "--review-requested") {
		return githubcli.ToDomainPullRequests(append([]githubcli.PullRequest(nil), loader.requestedPullRequests...)), nil
	}

	return githubcli.ToDomainPullRequests(append([]githubcli.PullRequest(nil), loader.myPullRequests...)), nil
}

func (loader *fakePullRequestDetailLoader) ListNotifications() ([]githubdomain.Notification, error) {
	if loader.notificationsErr != nil {
		return nil, loader.notificationsErr
	}
	if loader.markAllNotificationsReadPollLoads > 0 {
		loader.markAllNotificationsReadPollLoads--
		if loader.markAllNotificationsReadPollLoads == 0 {
			loader.markLoadedNotificationsRead()
		}
	}
	return githubcli.ToDomainNotifications(append([]githubcli.Notification(nil), loader.notifications...)), nil
}

func (loader *fakePullRequestDetailLoader) MarkNotificationRead(threadID string) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	loader.markNotificationReadIDs = append(loader.markNotificationReadIDs, trimmedThreadID)
	if loader.markNotificationReadErr != nil {
		return loader.markNotificationReadErr
	}
	loader.markNotificationRead(trimmedThreadID)
	return nil
}

func (loader *fakePullRequestDetailLoader) MarkNotificationDone(threadID string) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	loader.markNotificationDoneIDs = append(loader.markNotificationDoneIDs, trimmedThreadID)
	if loader.markNotificationDoneErr != nil {
		return loader.markNotificationDoneErr
	}
	loader.removeNotification(trimmedThreadID)
	return nil
}

func (loader *fakePullRequestDetailLoader) MarkAllNotificationsRead() (githubdomain.NotificationBulkReadResult, error) {
	loader.markAllNotificationsReadCalls++
	if loader.markAllNotificationsReadErr != nil {
		return githubdomain.NotificationBulkReadResult{}, loader.markAllNotificationsReadErr
	}
	if loader.markAllNotificationsReadAccepted {
		loader.markAllNotificationsReadPollLoads = maxInt(loader.markAllNotificationsReadPollLoads, 1)
		return githubdomain.NotificationBulkReadResult{Accepted: true}, nil
	}
	loader.markLoadedNotificationsRead()
	return githubdomain.NotificationBulkReadResult{}, nil
}

func (loader *fakePullRequestDetailLoader) MarkAllNotificationsDone(notifications []githubdomain.Notification) (int, error) {
	ids := make([]string, 0, len(notifications))
	for _, notification := range notifications {
		trimmedThreadID := strings.TrimSpace(notification.ID)
		if trimmedThreadID == "" {
			continue
		}
		ids = append(ids, trimmedThreadID)
	}
	loader.markAllNotificationsDoneIDs = append(loader.markAllNotificationsDoneIDs, append([]string(nil), ids...))
	if loader.markAllNotificationsDoneErr != nil {
		return 0, loader.markAllNotificationsDoneErr
	}
	for _, threadID := range ids {
		loader.removeNotification(threadID)
	}
	return len(ids), nil
}

func (loader *fakePullRequestDetailLoader) GetIssueDetail(repository string, number int) (githubdomain.IssueDetail, error) {
	key := repository + "#" + strconv.Itoa(number)
	loader.issueDetailCalls = append(loader.issueDetailCalls, key)
	if loader.issueDetailErrors != nil {
		if err, ok := loader.issueDetailErrors[key]; ok {
			return githubdomain.IssueDetail{}, err
		}
	}
	if loader.issueDetails != nil {
		if detail, ok := loader.issueDetails[key]; ok {
			return githubcli.ToDomainIssueDetail(detail), nil
		}
	}
	return githubdomain.IssueDetail{}, nil
}

func (loader *fakePullRequestDetailLoader) GetReleaseDetail(repository string, id int) (githubdomain.ReleaseDetail, error) {
	key := repository + "#" + strconv.Itoa(id)
	loader.releaseDetailCalls = append(loader.releaseDetailCalls, key)
	if loader.releaseDetailErrors != nil {
		if err, ok := loader.releaseDetailErrors[key]; ok {
			return githubdomain.ReleaseDetail{}, err
		}
	}
	if loader.releaseDetails != nil {
		if detail, ok := loader.releaseDetails[key]; ok {
			return githubcli.ToDomainReleaseDetail(detail), nil
		}
	}
	return githubdomain.ReleaseDetail{}, nil
}

func (loader *fakePullRequestDetailLoader) GetPullRequestDetail(repository string, number int) (githubdomain.PullRequestDetail, error) {
	key := repository + "#" + strconv.Itoa(number)
	loader.detailCalls = append(loader.detailCalls, key)
	if loader.detailErrors != nil {
		if err, ok := loader.detailErrors[key]; ok {
			return githubdomain.PullRequestDetail{}, err
		}
	}
	if loader.details != nil {
		if detail, ok := loader.details[key]; ok {
			return githubcli.ToDomainPullRequestDetail(detail), nil
		}
	}
	return githubdomain.PullRequestDetail{}, nil
}

func (loader *fakePullRequestDetailLoader) GetPullRequestDiff(repository string, number int) (githubdomain.PullRequestDiff, error) {
	key := repository + "#" + strconv.Itoa(number)
	loader.diffCalls = append(loader.diffCalls, key)
	if loader.diffErrors != nil {
		if err, ok := loader.diffErrors[key]; ok {
			return githubdomain.PullRequestDiff{}, err
		}
	}
	if loader.diffs != nil {
		if diff, ok := loader.diffs[key]; ok {
			return githubcli.ToDomainPullRequestDiff(diff), nil
		}
	}
	return githubdomain.PullRequestDiff{}, nil
}

func (loader *fakePullRequestDetailLoader) GetPullRequestFileTeamOwners(repository string, number int, filePaths []string) (map[string][]string, error) {
	key := repository + "#" + strconv.Itoa(number)
	loader.fileTeamOwnerCalls = append(loader.fileTeamOwnerCalls, key)
	loader.fileTeamOwnerPaths = append(loader.fileTeamOwnerPaths, append([]string(nil), filePaths...))
	if loader.fileTeamOwnerErrors != nil {
		if err, ok := loader.fileTeamOwnerErrors[key]; ok {
			return nil, err
		}
	}
	if loader.fileTeamOwners == nil {
		return nil, nil
	}
	teamOwnersByPath, ok := loader.fileTeamOwners[key]
	if !ok {
		return nil, nil
	}

	actual := make(map[string][]string, len(teamOwnersByPath))
	for path, teamOwners := range teamOwnersByPath {
		actual[path] = append([]string(nil), teamOwners...)
	}
	return actual, nil
}

func (loader *fakePullRequestDetailLoader) CommentOnPullRequest(repository string, number int, body string) error {
	loader.commentCalls = append(loader.commentCalls, repository+"#"+strconv.Itoa(number))
	loader.commentBodies = append(loader.commentBodies, body)
	if loader.commentErr != nil {
		return loader.commentErr
	}
	loader.addPullRequestComment(repository, number, body)
	return nil
}

func (loader *fakePullRequestDetailLoader) UpdatePullRequestComment(commentID string, body string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	loader.updatePullRequestCommentIDs = append(loader.updatePullRequestCommentIDs, trimmedCommentID)
	loader.updatePullRequestCommentBodies = append(loader.updatePullRequestCommentBodies, body)
	if loader.updatePullRequestCommentErr != nil {
		return loader.updatePullRequestCommentErr
	}
	loader.updatePullRequestComment(trimmedCommentID, body)
	return nil
}

func (loader *fakePullRequestDetailLoader) DeletePullRequestComment(commentID string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	loader.deletePullRequestCommentIDs = append(loader.deletePullRequestCommentIDs, trimmedCommentID)
	if loader.deletePullRequestCommentErr != nil {
		return loader.deletePullRequestCommentErr
	}
	loader.deletePullRequestComment(trimmedCommentID)
	return nil
}

func (loader *fakePullRequestDetailLoader) ApprovePullRequest(repository string, number int) error {
	loader.approveCalls = append(loader.approveCalls, repository+"#"+strconv.Itoa(number))
	return loader.approveErr
}

func (loader *fakePullRequestDetailLoader) ReviewPullRequestWithComment(repository string, number int, body string) error {
	loader.reviewCommentCalls = append(loader.reviewCommentCalls, repository+"#"+strconv.Itoa(number))
	loader.reviewCommentBodies = append(loader.reviewCommentBodies, body)
	return loader.reviewCommentErr
}

func (loader *fakePullRequestDetailLoader) RequestChangesOnPullRequest(repository string, number int, body string) error {
	loader.requestChangesCalls = append(loader.requestChangesCalls, repository+"#"+strconv.Itoa(number))
	loader.requestChangesBodies = append(loader.requestChangesBodies, body)
	return loader.requestChangesErr
}

func (loader *fakePullRequestDetailLoader) SubmitPullRequestReview(pullRequestReviewID string, event githubdomain.PullRequestReviewEvent, body string) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	loader.submitReviewIDs = append(loader.submitReviewIDs, trimmedReviewID)
	loader.submitReviewEvents = append(loader.submitReviewEvents, githubcli.PullRequestReviewEventFromDomain(event))
	loader.submitReviewBodies = append(loader.submitReviewBodies, body)
	if loader.submitReviewErr != nil {
		return loader.submitReviewErr
	}
	loader.clearPendingPullRequestReview(trimmedReviewID)
	return nil
}

func (loader *fakePullRequestDetailLoader) AddPullRequestReviewThread(pullRequestReviewID string, body string, target githubdomain.PullRequestReviewThreadTarget) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	legacyTarget := githubcli.PullRequestReviewThreadTargetFromDomain(target)
	loader.reviewThreadReviewIDs = append(loader.reviewThreadReviewIDs, trimmedReviewID)
	loader.reviewThreadBodies = append(loader.reviewThreadBodies, body)
	loader.reviewThreadTargets = append(loader.reviewThreadTargets, legacyTarget)
	if loader.reviewThreadErr != nil {
		return loader.reviewThreadErr
	}
	if loader.diffs != nil {
		if key, ok := loader.reviewKeyByPendingID[trimmedReviewID]; ok {
			diff := loader.diffs[key]
			diff.Threads = append(diff.Threads, githubcli.PullRequestReviewThread{
				ID:            "thread-" + strconv.Itoa(len(loader.reviewThreadReviewIDs)),
				Path:          legacyTarget.Path,
				Line:          legacyTarget.Line,
				StartLine:     legacyTarget.StartLine,
				DiffSide:      legacyTarget.Side,
				StartDiffSide: legacyTarget.StartSide,
				Comments: []githubcli.PullRequestComment{{
					Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
					Body:      body,
					CreatedAt: "2026-04-20T12:00:00Z",
					State:     "PENDING",
				}},
			})
			loader.diffs[key] = diff
		}
	}
	return nil
}

func (loader *fakePullRequestDetailLoader) AddPullRequestReviewThreadReply(pullRequestReviewID string, pullRequestReviewThreadID string, body string) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	trimmedThreadID := strings.TrimSpace(pullRequestReviewThreadID)
	loader.reviewThreadReplyReviewIDs = append(loader.reviewThreadReplyReviewIDs, trimmedReviewID)
	loader.reviewThreadReplyThreadIDs = append(loader.reviewThreadReplyThreadIDs, trimmedThreadID)
	loader.reviewThreadReplyBodies = append(loader.reviewThreadReplyBodies, body)
	if loader.reviewThreadReplyErr != nil {
		return loader.reviewThreadReplyErr
	}
	loader.addReviewThreadReply(trimmedReviewID, trimmedThreadID, body)
	return nil
}

func (loader *fakePullRequestDetailLoader) UpdatePullRequestReviewComment(commentID string, body string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	loader.updateReviewCommentIDs = append(loader.updateReviewCommentIDs, trimmedCommentID)
	loader.updateReviewCommentBodies = append(loader.updateReviewCommentBodies, body)
	if loader.updateReviewCommentErr != nil {
		return loader.updateReviewCommentErr
	}
	loader.updateReviewComment(trimmedCommentID, body)
	return nil
}

func (loader *fakePullRequestDetailLoader) DeletePullRequestReviewComment(commentID string) error {
	trimmedCommentID := strings.TrimSpace(commentID)
	loader.deleteReviewCommentIDs = append(loader.deleteReviewCommentIDs, trimmedCommentID)
	if loader.deleteReviewCommentErr != nil {
		return loader.deleteReviewCommentErr
	}
	loader.deleteReviewComment(trimmedCommentID)
	return nil
}

func (loader *fakePullRequestDetailLoader) ResolvePullRequestReviewThread(threadID string) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	loader.resolveReviewThreadIDs = append(loader.resolveReviewThreadIDs, trimmedThreadID)
	if loader.resolveReviewThreadErr != nil {
		return loader.resolveReviewThreadErr
	}
	loader.updateReviewThreadState(trimmedThreadID, true)
	return nil
}

func (loader *fakePullRequestDetailLoader) UnresolvePullRequestReviewThread(threadID string) error {
	trimmedThreadID := strings.TrimSpace(threadID)
	loader.unresolveReviewThreadIDs = append(loader.unresolveReviewThreadIDs, trimmedThreadID)
	if loader.unresolveReviewThreadErr != nil {
		return loader.unresolveReviewThreadErr
	}
	loader.updateReviewThreadState(trimmedThreadID, false)
	return nil
}

func (loader *fakePullRequestDetailLoader) AddReaction(subjectID string, content githubdomain.ReactionContent) error {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	legacyContent := githubcli.ReactionContentFromDomain(content)
	loader.addReactionSubjectIDs = append(loader.addReactionSubjectIDs, trimmedSubjectID)
	loader.addReactionContents = append(loader.addReactionContents, legacyContent)
	if loader.addReactionErr != nil {
		return loader.addReactionErr
	}
	loader.addReaction(trimmedSubjectID, legacyContent)
	return nil
}

func (loader *fakePullRequestDetailLoader) RemoveReaction(subjectID string, content githubdomain.ReactionContent) error {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	legacyContent := githubcli.ReactionContentFromDomain(content)
	loader.removeReactionSubjectIDs = append(loader.removeReactionSubjectIDs, trimmedSubjectID)
	loader.removeReactionContents = append(loader.removeReactionContents, legacyContent)
	if loader.removeReactionErr != nil {
		return loader.removeReactionErr
	}
	loader.removeReaction(trimmedSubjectID, legacyContent)
	return nil
}

func (loader *fakePullRequestDetailLoader) OpenPullRequestInBrowser(repository string, number int) error {
	loader.openBrowserCalls = append(loader.openBrowserCalls, repository+"#"+strconv.Itoa(number))
	return loader.openBrowserErr
}

func (loader *fakePullRequestDetailLoader) ListAssignableUsers(repository string) ([]githubdomain.PullRequestAuthor, error) {
	trimmedRepository := strings.TrimSpace(repository)
	loader.assignableUserCalls = append(loader.assignableUserCalls, trimmedRepository)
	if loader.assignableUserErr != nil {
		return nil, loader.assignableUserErr
	}
	if loader.assignableUsers != nil {
		if actual, ok := loader.assignableUsers[trimmedRepository]; ok {
			return githubcli.ToDomainPullRequestAuthors(append([]githubcli.PullRequestAuthor(nil), actual...)), nil
		}
	}
	return nil, nil
}

func (loader *fakePullRequestDetailLoader) SearchAssignableUsers(repository string, query string) ([]githubdomain.PullRequestAuthor, error) {
	trimmedRepository := strings.TrimSpace(repository)
	trimmedQuery := strings.TrimSpace(query)
	loader.searchAssignableUserCalls = append(loader.searchAssignableUserCalls, trimmedRepository+"|"+trimmedQuery)
	if loader.searchAssignableUserErr != nil {
		return nil, loader.searchAssignableUserErr
	}
	if loader.searchAssignableUsers != nil {
		if actual, ok := loader.searchAssignableUsers[trimmedRepository+"|"+trimmedQuery]; ok {
			return githubcli.ToDomainPullRequestAuthors(append([]githubcli.PullRequestAuthor(nil), actual...)), nil
		}
	}
	if loader.assignableUsers != nil {
		if actual, ok := loader.assignableUsers[trimmedRepository]; ok {
			return githubcli.ToDomainPullRequestAuthors(append([]githubcli.PullRequestAuthor(nil), actual...)), nil
		}
	}
	return nil, nil
}

func (loader *fakePullRequestDetailLoader) UpdatePullRequestAssignees(repository string, number int, addLogins []string, removeLogins []string) error {
	loader.updateAssigneeCalls = append(loader.updateAssigneeCalls, repository+"#"+strconv.Itoa(number))
	loader.updateAssigneeAdditions = append(loader.updateAssigneeAdditions, append([]string(nil), addLogins...))
	loader.updateAssigneeRemovals = append(loader.updateAssigneeRemovals, append([]string(nil), removeLogins...))
	if loader.updateAssigneeErr != nil {
		return loader.updateAssigneeErr
	}

	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		updatedAssignees := append([]githubcli.PullRequestAuthor(nil), detail.Assignees...)
		for _, login := range removeLogins {
			trimmedLogin := strings.TrimSpace(login)
			filteredAssignees := updatedAssignees[:0]
			for _, assignee := range updatedAssignees {
				if strings.TrimSpace(assignee.Login) == trimmedLogin {
					continue
				}
				filteredAssignees = append(filteredAssignees, assignee)
			}
			updatedAssignees = filteredAssignees
		}
		for _, login := range addLogins {
			trimmedLogin := strings.TrimSpace(login)
			if trimmedLogin == "" {
				continue
			}
			alreadyAssigned := false
			for _, assignee := range updatedAssignees {
				if strings.TrimSpace(assignee.Login) == trimmedLogin {
					alreadyAssigned = true
					break
				}
			}
			if alreadyAssigned {
				continue
			}
			assignee := githubcli.PullRequestAuthor{Login: trimmedLogin}
			for _, candidate := range loader.assignableUsers[strings.TrimSpace(repository)] {
				if strings.TrimSpace(candidate.Login) == trimmedLogin {
					assignee = candidate
					break
				}
			}
			updatedAssignees = append(updatedAssignees, assignee)
		}
		detail.Assignees = updatedAssignees
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) RequestPullRequestReviewer(repository string, number int, reviewerLogin string) error {
	trimmedLogin := strings.TrimSpace(reviewerLogin)
	loader.requestReviewerCalls = append(loader.requestReviewerCalls, repository+"#"+strconv.Itoa(number))
	loader.requestReviewerLogins = append(loader.requestReviewerLogins, trimmedLogin)
	if loader.requestReviewerErr != nil {
		return loader.requestReviewerErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		for _, existing := range pullRequest.ReviewRequests {
			if strings.TrimSpace(existing.RequestedReviewer.Login) == trimmedLogin {
				return
			}
		}
		pullRequest.ReviewRequests = append(pullRequest.ReviewRequests, githubcli.PullRequestReviewRequest{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: trimmedLogin}})
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		for _, existing := range detail.ReviewRequests {
			if strings.TrimSpace(existing.RequestedReviewer.Login) == trimmedLogin {
				return
			}
		}
		detail.ReviewRequests = append(detail.ReviewRequests, githubcli.PullRequestReviewRequest{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: trimmedLogin}})
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) EditPullRequestTitle(repository string, number int, title string) error {
	loader.editTitleCalls = append(loader.editTitleCalls, repository+"#"+strconv.Itoa(number))
	loader.editTitleValues = append(loader.editTitleValues, title)
	if loader.editTitleErr != nil {
		return loader.editTitleErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.Title = title
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.Title = title
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) EditPullRequestDescription(repository string, number int, body string) error {
	loader.editDescriptionCalls = append(loader.editDescriptionCalls, repository+"#"+strconv.Itoa(number))
	loader.editDescriptionBodies = append(loader.editDescriptionBodies, body)
	if loader.editDescriptionErr != nil {
		return loader.editDescriptionErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.Body = body
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.Body = body
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) MarkPullRequestReadyForReview(repository string, number int) error {
	loader.markReadyForReviewCalls = append(loader.markReadyForReviewCalls, repository+"#"+strconv.Itoa(number))
	if loader.markReadyForReviewErr != nil {
		return loader.markReadyForReviewErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.State = "OPEN"
		pullRequest.IsDraft = false
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.State = "OPEN"
		detail.IsDraft = false
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) ConvertPullRequestToDraft(repository string, number int) error {
	loader.convertToDraftCalls = append(loader.convertToDraftCalls, repository+"#"+strconv.Itoa(number))
	if loader.convertToDraftErr != nil {
		return loader.convertToDraftErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.State = "OPEN"
		pullRequest.IsDraft = true
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.State = "OPEN"
		detail.IsDraft = true
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) ClosePullRequest(repository string, number int) error {
	loader.closePullRequestCalls = append(loader.closePullRequestCalls, repository+"#"+strconv.Itoa(number))
	if loader.closePullRequestErr != nil {
		return loader.closePullRequestErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.State = "CLOSED"
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.State = "CLOSED"
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) ReopenPullRequest(repository string, number int) error {
	loader.reopenPullRequestCalls = append(loader.reopenPullRequestCalls, repository+"#"+strconv.Itoa(number))
	if loader.reopenPullRequestErr != nil {
		return loader.reopenPullRequestErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.State = "OPEN"
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.State = "OPEN"
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) SquashMergePullRequest(repository string, number int) error {
	loader.squashMergeCalls = append(loader.squashMergeCalls, repository+"#"+strconv.Itoa(number))
	if loader.squashMergeErr != nil {
		return loader.squashMergeErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.State = "MERGED"
		pullRequest.IsDraft = false
		pullRequest.ReviewDecision = ""
		pullRequest.ReviewRequests = nil
		pullRequest.Mergeable = ""
		pullRequest.MergeStateStatus = ""
		pullRequest.AutoMergeRequest = nil
		pullRequest.StatusCheckRollupState = ""
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.State = "MERGED"
		detail.IsDraft = false
		detail.Mergeable = ""
		detail.MergeStateStatus = ""
		detail.AutoMergeRequest = nil
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) EnablePullRequestAutoMerge(repository string, number int) error {
	loader.enableAutoMergeCalls = append(loader.enableAutoMergeCalls, repository+"#"+strconv.Itoa(number))
	if loader.enableAutoMergeErr != nil {
		return loader.enableAutoMergeErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.AutoMergeRequest = &githubcli.PullRequestAutoMergeRequest{}
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.AutoMergeRequest = &githubcli.PullRequestAutoMergeRequest{}
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) DisablePullRequestAutoMerge(repository string, number int) error {
	loader.disableAutoMergeCalls = append(loader.disableAutoMergeCalls, repository+"#"+strconv.Itoa(number))
	if loader.disableAutoMergeErr != nil {
		return loader.disableAutoMergeErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		pullRequest.AutoMergeRequest = nil
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.AutoMergeRequest = nil
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) UpdatePullRequestBranch(repository string, number int) error {
	loader.updateBranchCalls = append(loader.updateBranchCalls, repository+"#"+strconv.Itoa(number))
	if loader.updateBranchErr != nil {
		return loader.updateBranchErr
	}

	loader.updatePullRequestSummary(repository, number, func(pullRequest *githubcli.PullRequest) {
		if strings.EqualFold(strings.TrimSpace(pullRequest.MergeStateStatus), "BEHIND") {
			pullRequest.MergeStateStatus = ""
		}
	})
	loader.updatePullRequestDetail(repository, number, func(detail *githubcli.PullRequestDetail) {
		detail.OutOfDateWithBase = false
		if strings.EqualFold(strings.TrimSpace(detail.MergeStateStatus), "BEHIND") {
			detail.MergeStateStatus = ""
		}
	})
	return nil
}

func (loader *fakePullRequestDetailLoader) StartPendingPullRequestReview(repository string, number int) (string, error) {
	loader.startReviewCalls = append(loader.startReviewCalls, repository+"#"+strconv.Itoa(number))
	if loader.startReviewErr != nil {
		return "", loader.startReviewErr
	}
	reviewID := strings.TrimSpace(loader.startReviewID)
	if reviewID == "" {
		reviewID = "PRR_pending"
	}
	if loader.reviewKeyByPendingID == nil {
		loader.reviewKeyByPendingID = map[string]string{}
	}
	loader.reviewKeyByPendingID[reviewID] = repository + "#" + strconv.Itoa(number)
	return reviewID, nil
}

func (loader *fakePullRequestDetailLoader) GetPendingPullRequestReviewID(repository string, number int) (string, bool, error) {
	key := strings.TrimSpace(repository) + "#" + strconv.Itoa(number)
	loader.getPendingReviewCalls = append(loader.getPendingReviewCalls, key)
	if loader.getPendingReviewErr != nil {
		return "", false, loader.getPendingReviewErr
	}
	return loader.pendingPullRequestReviewID(key)
}

func (loader *fakePullRequestDetailLoader) DeletePullRequestReview(pullRequestReviewID string) error {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	loader.deletePullRequestReviewIDs = append(loader.deletePullRequestReviewIDs, trimmedReviewID)
	if loader.deletePullRequestReviewErr != nil {
		return loader.deletePullRequestReviewErr
	}
	loader.clearPendingPullRequestReview(trimmedReviewID)
	return nil
}

func (loader *fakePullRequestDetailLoader) pendingPullRequestReviewID(key string) (string, bool, error) {
	trimmedKey := strings.TrimSpace(key)
	for reviewID, reviewKey := range loader.reviewKeyByPendingID {
		if strings.TrimSpace(reviewKey) != trimmedKey {
			continue
		}
		trimmedReviewID := strings.TrimSpace(reviewID)
		if trimmedReviewID != "" {
			return trimmedReviewID, true, nil
		}
	}
	return "", false, nil
}

func (loader *fakePullRequestDetailLoader) clearPendingPullRequestReview(pullRequestReviewID string) {
	trimmedReviewID := strings.TrimSpace(pullRequestReviewID)
	if trimmedReviewID == "" || loader.reviewKeyByPendingID == nil {
		return
	}
	delete(loader.reviewKeyByPendingID, trimmedReviewID)
}

func (loader *fakePullRequestDetailLoader) GetPullRequestBuildRun(repository string, check githubdomain.PullRequestStatusCheck) (string, error) {
	legacyCheck := githubcli.PullRequestStatusCheckFromDomain(check)
	loader.buildRunCalls = append(loader.buildRunCalls, strings.TrimSpace(repository))
	loader.buildRunChecks = append(loader.buildRunChecks, legacyCheck)
	if loader.buildRunErr != nil {
		return "", loader.buildRunErr
	}
	if loader.buildRuns != nil {
		if actual, ok := loader.buildRuns[strings.TrimSpace(legacyCheck.Link)]; ok {
			return actual, nil
		}
	}
	return "", githubcli.ErrMissingPullRequestBuildLink
}

func (loader *fakePullRequestDetailLoader) GetPullRequestBuildRunJobs(repository string, check githubdomain.PullRequestStatusCheck) ([]githubdomain.PullRequestBuildRunJob, error) {
	legacyCheck := githubcli.PullRequestStatusCheckFromDomain(check)
	loader.buildRunJobCalls = append(loader.buildRunJobCalls, strings.TrimSpace(repository))
	loader.buildRunJobChecks = append(loader.buildRunJobChecks, legacyCheck)
	if loader.buildRunJobsErr != nil {
		return nil, loader.buildRunJobsErr
	}
	if loader.buildRunJobs != nil {
		if actual, ok := loader.buildRunJobs[strings.TrimSpace(legacyCheck.Link)]; ok {
			return githubcli.ToDomainPullRequestBuildRunJobs(append([]githubcli.PullRequestBuildRunJob(nil), actual...)), nil
		}
	}
	return nil, nil
}

func (loader *fakePullRequestDetailLoader) GetPullRequestBuildRunJobLog(_ string, jobDatabaseID int) (string, error) {
	loader.buildLogCalls = append(loader.buildLogCalls, jobDatabaseID)
	if loader.buildLogErr != nil {
		return "", loader.buildLogErr
	}
	if loader.buildLogs != nil {
		if actual, ok := loader.buildLogs[jobDatabaseID]; ok {
			return actual, nil
		}
	}
	return "", githubcli.ErrMissingPullRequestBuildLink
}

func (loader *fakePullRequestDetailLoader) GetPullRequestBuildRunJobLogForCheck(repository string, check githubdomain.PullRequestStatusCheck) (githubdomain.PullRequestBuildRunJob, string, error) {
	jobs, err := loader.GetPullRequestBuildRunJobs(repository, check)
	if err != nil {
		return githubdomain.PullRequestBuildRunJob{}, "", err
	}

	trimmedCheckName := strings.TrimSpace(check.Name)
	for _, job := range jobs {
		if strings.EqualFold(strings.TrimSpace(job.Name), trimmedCheckName) {
			actualLog, actualErr := loader.GetPullRequestBuildRunJobLog(repository, job.DatabaseID)
			return job, actualLog, actualErr
		}
	}
	if len(jobs) == 1 {
		actualLog, actualErr := loader.GetPullRequestBuildRunJobLog(repository, jobs[0].DatabaseID)
		return jobs[0], actualLog, actualErr
	}
	return githubdomain.PullRequestBuildRunJob{}, "", githubcli.ErrPullRequestBuildRunJobNotFound
}

func (loader *fakePullRequestDetailLoader) RenderMarkdownHTML(repository string, markdown string) (string, error) {
	key := strings.TrimSpace(repository) + "|" + strings.TrimSpace(markdown)
	loader.renderMarkdownHTMLCalls = append(loader.renderMarkdownHTMLCalls, key)
	if loader.renderMarkdownHTMLErr != nil {
		return "", loader.renderMarkdownHTMLErr
	}
	if loader.renderedMarkdownHTML != nil {
		if actual, ok := loader.renderedMarkdownHTML[key]; ok {
			return actual, nil
		}
	}
	return "", nil
}

func (loader *fakePullRequestDetailLoader) GetAuthToken() (string, error) {
	loader.authTokenCalls++
	if loader.authTokenErr != nil {
		return "", loader.authTokenErr
	}
	return strings.TrimSpace(loader.authToken), nil
}

func (loader *fakePullRequestDetailLoader) updatePullRequestSummary(repository string, number int, update func(*githubcli.PullRequest)) {
	for index := range loader.myPullRequests {
		if loader.myPullRequests[index].Repository.NameWithOwner == repository && loader.myPullRequests[index].Number == number {
			update(&loader.myPullRequests[index])
		}
	}
	for index := range loader.requestedPullRequests {
		if loader.requestedPullRequests[index].Repository.NameWithOwner == repository && loader.requestedPullRequests[index].Number == number {
			update(&loader.requestedPullRequests[index])
		}
	}
}

func (loader *fakePullRequestDetailLoader) updatePullRequestDetail(repository string, number int, update func(*githubcli.PullRequestDetail)) {
	if loader.details == nil {
		return
	}
	key := repository + "#" + strconv.Itoa(number)
	detail, ok := loader.details[key]
	if !ok {
		return
	}
	update(&detail)
	loader.details[key] = detail
}

func (loader *fakePullRequestDetailLoader) markLoadedNotificationsRead() {
	for index := range loader.notifications {
		loader.notifications[index].Unread = false
	}
}

func (loader *fakePullRequestDetailLoader) markNotificationRead(threadID string) {
	trimmedThreadID := strings.TrimSpace(threadID)
	for index := range loader.notifications {
		if strings.TrimSpace(loader.notifications[index].ID) != trimmedThreadID {
			continue
		}
		loader.notifications[index].Unread = false
		return
	}
}

func (loader *fakePullRequestDetailLoader) removeNotification(threadID string) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" || len(loader.notifications) == 0 {
		return
	}
	filteredNotifications := loader.notifications[:0]
	for _, notification := range loader.notifications {
		if strings.TrimSpace(notification.ID) == trimmedThreadID {
			continue
		}
		filteredNotifications = append(filteredNotifications, notification)
	}
	loader.notifications = append([]githubcli.Notification(nil), filteredNotifications...)
}

func (loader *fakePullRequestDetailLoader) addReaction(subjectID string, content githubcli.ReactionContent) {
	loader.updateReactionGroups(subjectID, func(groups []githubcli.ReactionGroup) []githubcli.ReactionGroup {
		return given_reactionGroupsWithAddedReaction(groups, content)
	})
}

func (loader *fakePullRequestDetailLoader) removeReaction(subjectID string, content githubcli.ReactionContent) {
	loader.updateReactionGroups(subjectID, func(groups []githubcli.ReactionGroup) []githubcli.ReactionGroup {
		return given_reactionGroupsWithRemovedReaction(groups, content)
	})
}

func (loader *fakePullRequestDetailLoader) updateReactionGroups(subjectID string, update func([]githubcli.ReactionGroup) []githubcli.ReactionGroup) {
	trimmedSubjectID := strings.TrimSpace(subjectID)
	if trimmedSubjectID == "" || update == nil {
		return
	}

	for key, detail := range loader.details {
		updated := false
		if strings.TrimSpace(detail.ID) == trimmedSubjectID {
			detail.ReactionGroups = update(detail.ReactionGroups)
			updated = true
		}
		for index := range detail.Comments {
			if strings.TrimSpace(detail.Comments[index].ID) != trimmedSubjectID {
				continue
			}
			detail.Comments[index].ReactionGroups = update(detail.Comments[index].ReactionGroups)
			updated = true
		}
		for index := range detail.InlineComments {
			if strings.TrimSpace(detail.InlineComments[index].ID) != trimmedSubjectID {
				continue
			}
			detail.InlineComments[index].ReactionGroups = update(detail.InlineComments[index].ReactionGroups)
			updated = true
		}
		for threadIndex := range detail.InlineCommentThreads {
			for commentIndex := range detail.InlineCommentThreads[threadIndex].Comments {
				if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ID) != trimmedSubjectID {
					continue
				}
				detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ReactionGroups = update(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ReactionGroups)
				updated = true
			}
		}
		if updated {
			loader.details[key] = detail
		}
	}
	for key, diff := range loader.diffs {
		updated := false
		for threadIndex := range diff.Threads {
			for commentIndex := range diff.Threads[threadIndex].Comments {
				if strings.TrimSpace(diff.Threads[threadIndex].Comments[commentIndex].ID) != trimmedSubjectID {
					continue
				}
				diff.Threads[threadIndex].Comments[commentIndex].ReactionGroups = update(diff.Threads[threadIndex].Comments[commentIndex].ReactionGroups)
				updated = true
			}
		}
		if updated {
			loader.diffs[key] = diff
		}
	}
}

func given_reactionGroupsWithAddedReaction(groups []githubcli.ReactionGroup, content githubcli.ReactionContent) []githubcli.ReactionGroup {
	updatedGroups := append([]githubcli.ReactionGroup(nil), groups...)
	for index := range updatedGroups {
		if strings.TrimSpace(string(updatedGroups[index].Content)) != strings.TrimSpace(string(content)) {
			continue
		}
		updatedGroups[index].TotalCount++
		updatedGroups[index].ViewerHasReacted = true
		return updatedGroups
	}
	return append(updatedGroups, githubcli.ReactionGroup{Content: content, TotalCount: 1, ViewerHasReacted: true})
}

func given_reactionGroupsWithRemovedReaction(groups []githubcli.ReactionGroup, content githubcli.ReactionContent) []githubcli.ReactionGroup {
	updatedGroups := append([]githubcli.ReactionGroup(nil), groups...)
	filteredGroups := updatedGroups[:0]
	for _, group := range updatedGroups {
		if strings.TrimSpace(string(group.Content)) != strings.TrimSpace(string(content)) {
			filteredGroups = append(filteredGroups, group)
			continue
		}
		if group.TotalCount <= 1 {
			continue
		}
		group.TotalCount--
		group.ViewerHasReacted = false
		filteredGroups = append(filteredGroups, group)
	}
	return append([]githubcli.ReactionGroup(nil), filteredGroups...)
}

func (loader *fakePullRequestDetailLoader) addReviewThreadReply(reviewID string, threadID string, body string) {
	trimmedThreadID := strings.TrimSpace(threadID)
	if trimmedThreadID == "" {
		return
	}

	commentID := "PRRC_reply_" + strconv.Itoa(len(loader.reviewThreadReplyBodies))
	state := ""
	if strings.TrimSpace(reviewID) != "" {
		state = "PENDING"
	}
	reply := githubcli.PullRequestComment{
		ID:              commentID,
		ViewerDidAuthor: true,
		Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
		Body:            body,
		CreatedAt:       "2026-04-20T12:30:00Z",
		State:           state,
	}

	for key, detail := range loader.details {
		updated := false
		for threadIndex := range detail.InlineCommentThreads {
			if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].ID) != trimmedThreadID {
				continue
			}
			detail.InlineCommentThreads[threadIndex].Comments = append(detail.InlineCommentThreads[threadIndex].Comments, reply)
			updated = true
		}
		if updated {
			loader.details[key] = detail
		}
	}
	for key, diff := range loader.diffs {
		updated := false
		for threadIndex := range diff.Threads {
			if strings.TrimSpace(diff.Threads[threadIndex].ID) != trimmedThreadID {
				continue
			}
			diff.Threads[threadIndex].Comments = append(diff.Threads[threadIndex].Comments, reply)
			updated = true
		}
		if updated {
			loader.diffs[key] = diff
		}
	}
}

func (loader *fakePullRequestDetailLoader) addPullRequestComment(repository string, number int, body string) {
	key := strings.TrimSpace(repository) + "#" + strconv.Itoa(number)
	detail, ok := loader.details[key]
	if !ok {
		return
	}
	commentID := "PRC_" + strconv.Itoa(len(loader.commentBodies))
	detail.Comments = append(detail.Comments, githubcli.PullRequestComment{
		ID:              commentID,
		ViewerDidAuthor: true,
		Author:          &githubcli.PullRequestCommentAuthor{Login: "octocat"},
		Body:            body,
		CreatedAt:       "2026-04-20T12:15:00Z",
	})
	loader.details[key] = detail
}

func (loader *fakePullRequestDetailLoader) updatePullRequestComment(commentID string, body string) {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return
	}
	for key, detail := range loader.details {
		updated := false
		for commentIndex := range detail.Comments {
			if strings.TrimSpace(detail.Comments[commentIndex].ID) != trimmedCommentID {
				continue
			}
			detail.Comments[commentIndex].Body = body
			updated = true
		}
		if updated {
			loader.details[key] = detail
		}
	}
}

func (loader *fakePullRequestDetailLoader) deletePullRequestComment(commentID string) {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return
	}
	for key, detail := range loader.details {
		updated := false
		comments := detail.Comments[:0]
		for _, comment := range detail.Comments {
			if strings.TrimSpace(comment.ID) == trimmedCommentID {
				updated = true
				continue
			}
			comments = append(comments, comment)
		}
		detail.Comments = comments
		if updated {
			loader.details[key] = detail
		}
	}
}

func (loader *fakePullRequestDetailLoader) updateReviewComment(commentID string, body string) {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return
	}
	for key, detail := range loader.details {
		updated := false
		for threadIndex := range detail.InlineCommentThreads {
			for commentIndex := range detail.InlineCommentThreads[threadIndex].Comments {
				if strings.TrimSpace(detail.InlineCommentThreads[threadIndex].Comments[commentIndex].ID) != trimmedCommentID {
					continue
				}
				detail.InlineCommentThreads[threadIndex].Comments[commentIndex].Body = body
				updated = true
			}
		}
		for commentIndex := range detail.InlineComments {
			if strings.TrimSpace(detail.InlineComments[commentIndex].ID) != trimmedCommentID {
				continue
			}
			detail.InlineComments[commentIndex].Body = body
			updated = true
		}
		if updated {
			loader.details[key] = detail
		}
	}
	for key, diff := range loader.diffs {
		updated := false
		for threadIndex := range diff.Threads {
			for commentIndex := range diff.Threads[threadIndex].Comments {
				if strings.TrimSpace(diff.Threads[threadIndex].Comments[commentIndex].ID) != trimmedCommentID {
					continue
				}
				diff.Threads[threadIndex].Comments[commentIndex].Body = body
				updated = true
			}
		}
		if updated {
			loader.diffs[key] = diff
		}
	}
}

func (loader *fakePullRequestDetailLoader) deleteReviewComment(commentID string) {
	trimmedCommentID := strings.TrimSpace(commentID)
	if trimmedCommentID == "" {
		return
	}
	for key, detail := range loader.details {
		updated := false
		filteredThreads := detail.InlineCommentThreads[:0]
		for _, thread := range detail.InlineCommentThreads {
			comments := thread.Comments[:0]
			for _, comment := range thread.Comments {
				if strings.TrimSpace(comment.ID) == trimmedCommentID {
					updated = true
					continue
				}
				comments = append(comments, comment)
			}
			thread.Comments = comments
			if len(thread.Comments) > 0 {
				filteredThreads = append(filteredThreads, thread)
			} else if updated {
				continue
			}
		}
		detail.InlineCommentThreads = filteredThreads
		comments := detail.InlineComments[:0]
		for _, comment := range detail.InlineComments {
			if strings.TrimSpace(comment.ID) == trimmedCommentID {
				updated = true
				continue
			}
			comments = append(comments, comment)
		}
		detail.InlineComments = comments
		if updated {
			loader.details[key] = detail
		}
	}
	for key, diff := range loader.diffs {
		updated := false
		filteredThreads := diff.Threads[:0]
		for _, thread := range diff.Threads {
			comments := thread.Comments[:0]
			for _, comment := range thread.Comments {
				if strings.TrimSpace(comment.ID) == trimmedCommentID {
					updated = true
					continue
				}
				comments = append(comments, comment)
			}
			thread.Comments = comments
			if len(thread.Comments) > 0 {
				filteredThreads = append(filteredThreads, thread)
			} else if updated {
				continue
			}
		}
		diff.Threads = filteredThreads
		if updated {
			loader.diffs[key] = diff
		}
	}
}

func (loader *fakePullRequestDetailLoader) updateReviewThreadState(threadID string, resolved bool) {
	if trimmedThreadID := strings.TrimSpace(threadID); trimmedThreadID == "" {
		return
	} else {
		for key, detail := range loader.details {
			updated := false
			for index := range detail.InlineCommentThreads {
				if strings.TrimSpace(detail.InlineCommentThreads[index].ID) != trimmedThreadID {
					continue
				}
				detail.InlineCommentThreads[index].IsResolved = resolved
				updated = true
			}
			if updated {
				loader.details[key] = detail
			}
		}
		for key, diff := range loader.diffs {
			updated := false
			for index := range diff.Threads {
				if strings.TrimSpace(diff.Threads[index].ID) != trimmedThreadID {
					continue
				}
				diff.Threads[index].IsResolved = resolved
				updated = true
			}
			if updated {
				loader.diffs[key] = diff
			}
		}
	}
}

type inlineAsyncRunner struct{}

func (inlineAsyncRunner) Go(run func()) {
	run()
}

type capturingAsyncRunner struct {
	runs []func()
}

func (runner *capturingAsyncRunner) Go(run func()) {
	runner.runs = append(runner.runs, run)
}

type immediateUIUpdater struct{}

func (immediateUIUpdater) Apply(gui *gocui.Gui, update func(*gocui.Gui) error) {
	_ = update(gui)
}
