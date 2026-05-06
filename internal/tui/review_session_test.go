package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestReviewMode_GivenStartReviewActionSelected_WhenExecuting_ThenItRepurposesTheExistingThreePanesAndLoadsTheFileTreeAndFirstDiff(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 2,
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

	if !subject.reviewSession.active {
		t.Fatal("expected review mode to be active")
	}
	if subject.reviewSession.pendingReviewID != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", subject.reviewSession.pendingReviewID)
	}
	then_currentViewNameIs(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if metadataView.Title != "[1]-Metadata" {
		t.Fatalf("expected metadata view title %q, actual %q", "[1]-Metadata", metadataView.Title)
	}
	if !strings.Contains(metadataView.Buffer(), "Pending review: PRR_pending") || !strings.Contains(metadataView.Buffer(), "Changed files: 2") || !strings.Contains(metadataView.Buffer(), "+3") || !strings.Contains(metadataView.Buffer(), "-2") {
		t.Fatalf("expected metadata view to contain review stats, actual %q", metadataView.Buffer())
	}

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if filesView.Title != "[2]-Files" {
		t.Fatalf("expected files view title %q, actual %q", "[2]-Files", filesView.Title)
	}
	if !strings.Contains(filesView.Buffer(), "󰝰 internal/tui/") || !strings.Contains(filesView.Buffer(), " render.go") || !strings.Contains(filesView.Buffer(), " model.go") {
		t.Fatalf("expected files view to contain the iconified collapsed file tree, actual %q", filesView.Buffer())
	}
	if len(filesView.Tabs) != 0 {
		t.Fatalf("expected review files view to hide pull request tabs, actual %v", filesView.Tabs)
	}

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != "[0]-Diff" {
		t.Fatalf("expected detail view title %q, actual %q", "[0]-Diff", detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), detailInlineCommentLocationIcon+" internal/tui/render.go") || !strings.Contains(detailView.Buffer(), "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected detail view to contain the iconified first parsed diff file, actual %q", detailView.Buffer())
	}
	if len(detailView.Tabs) != 0 {
		t.Fatalf("expected review diff view to hide browser detail tabs, actual %v", detailView.Tabs)
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff calls %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
}

func TestReviewMode_GivenFilesWithInlineComments_WhenRenderingViewTwo_ThenItShowsCommentCountsBesideFileNames(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{
		{
			ID:       "thread-1",
			Path:     "internal/tui/render.go",
			Line:     3,
			DiffSide: "RIGHT",
			Comments: []githubcli.PullRequestComment{{Body: "First reply"}, {Body: "Second reply"}},
		},
		{
			ID:           "thread-2",
			Path:         "internal/tui/model.go",
			OriginalLine: 10,
			DiffSide:     "LEFT",
			Comments:     []githubcli.PullRequestComment{{Body: "Needs work"}},
		},
	}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
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

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(filesView.Buffer(), "render.go  2") || !strings.Contains(filesView.Buffer(), "model.go  1") {
		t.Fatalf("expected files view to contain per-file comment counts, actual %q", filesView.Buffer())
	}
	if strings.Contains(filesView.Buffer(), "internal/tui/ (") {
		t.Fatalf("expected directories to stay count-free, actual %q", filesView.Buffer())
	}
}

func TestReviewMode_GivenTheMetadataPaneSelected_WhenRendering_ThenViewZeroShowsThePullRequestDescription(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				Author:         &githubcli.PullRequestAuthor{Login: "octocat"},
				Labels:         []githubcli.PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
				Assignees:      []githubcli.PullRequestAuthor{{Login: "assignee-one"}},
				ReviewRequests: []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}}},
				BaseRefName:    "main",
				HeadRefName:    "feature/review",
				State:          "OPEN",
				CreatedAt:      "2026-04-18T10:00:00Z",
				UpdatedAt:      "2026-04-18T12:30:00Z",
				Additions:      12,
				Deletions:      3,
				ChangedFiles:   2,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionPullRequestDiff(),
		},
	}
	model := given_pullRequestCommentModel()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "First PR", Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 42", URL: "https://github.com/acme/widgets/pull/42"}),
	})
	subject := given_pullRequestCommentProgram(model, loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Body 42": "Rendered body 42"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusUserView(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDescriptionTitle {
		t.Fatalf("expected detail view title %q, actual %q", reviewModeDescriptionTitle, detailView.Title)
	}
	for _, expected := range []string{detailRepositoryIcon + " acme/widgets#42", detailAuthorIcon + " @octocat", "Created: 2026-04-18 10:00 UTC", "Updated: 2026-04-18 12:30 UTC", "+12", "-3", detailLabelIcon + " bug", detailAssigneesIcon + " @assignee-one", detailReviewRequestsIcon + " @reviewer-requested", "Rendered body 42"} {
		if !strings.Contains(detailView.Buffer(), expected) {
			t.Fatalf("expected the review detail pane to contain %q, actual %q", expected, detailView.Buffer())
		}
	}
	if strings.Contains(detailView.Buffer(), "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected the review detail pane to hide the diff while metadata is selected, actual %q", detailView.Buffer())
	}

	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDescriptionTitle {
		t.Fatalf("expected detail view title %q when focusing view 0 from metadata, actual %q", reviewModeDescriptionTitle, detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), detailAuthorIcon+" @octocat") || !strings.Contains(detailView.Buffer(), "Rendered body 42") {
		t.Fatalf("expected the review detail pane to keep showing the rendered description with metadata, actual %q", detailView.Buffer())
	}

	actualErr = subject.focusPullRequestsView(gui, nil)
	then_noError(t, actualErr)
	if detailView.Title != reviewModeDiffTitle {
		t.Fatalf("expected detail view title %q after returning to files, actual %q", reviewModeDiffTitle, detailView.Title)
	}
	if !strings.Contains(detailView.Buffer(), "@@ -1,2 +1,3 @@") {
		t.Fatalf("expected the review detail pane to restore the diff after returning to files, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered body 42") {
		t.Fatalf("expected the review detail pane to hide the description after returning to files, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenTheSelectedFileDiff_WhenRendering_ThenViewZeroUsesDiffColorsForCountsHunksAndChangedLines(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 2,
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

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headerLineIndex := given_viewLineIndexContaining(t, detailView, "internal/tui/render.go")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, headerLineIndex, "+2", given_themeColorHex(t, theme.DiffAdditionForegroundHex), "review diff addition count")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, headerLineIndex, "-1", given_themeColorHex(t, theme.DiffDeletionForegroundHex), "review diff deletion count")

	hunkHeaderLineIndex := given_viewLineIndexContaining(t, detailView, "@@ -1,2 +1,3 @@")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, hunkHeaderLineIndex, "@@ -1,2 +1,3 @@", given_themeColorHex(t, theme.DiffHunkHeaderHex), "review diff hunk header")

	contextLineIndex := given_viewLineIndexContaining(t, detailView, "context")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, contextLineIndex, "1 : 1 │", given_themeColorHex(t, theme.DiffLineNumberHex), "review diff line numbers")

	deletionLineIndex := given_viewLineIndexContaining(t, detailView, "old line")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, deletionLineIndex, "old line", given_themeColorHex(t, theme.DiffDeletionForegroundHex), "review diff deletion text")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, "old", given_themeColorHex(t, theme.DiffDeletionHighlightBackgroundHex), "review diff deletion changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, deletionLineIndex, " line", given_themeColorHex(t, theme.DiffDeletionBackgroundHex), "review diff deletion unchanged background")

	additionLineIndex := given_viewLineIndexContaining(t, detailView, "new line")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, additionLineIndex, "new line", given_themeColorHex(t, theme.DiffAdditionForegroundHex), "review diff addition text")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, "new", given_themeColorHex(t, theme.DiffAdditionHighlightBackgroundHex), "review diff addition changed background")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, additionLineIndex, " line", given_themeColorHex(t, theme.DiffAdditionBackgroundHex), "review diff addition unchanged background")
}

func TestReviewMode_GivenInlineReviewThreads_WhenRendering_ThenTheAuthorBadgeUsesTheInlineCommentHeaderColors(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 2,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": diff,
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	authorLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-one")
	then_viewLineSegmentHasBackgroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.CommentAuthorBadgeBackgroundHex), "review thread author badge background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, authorLineIndex, detailCommentsIcon+" @reviewer-one", given_themeColorHex(t, theme.CommentAuthorBadgeForegroundHex), "review thread author badge foreground")
	if !strings.Contains(detailView.BufferLines()[authorLineIndex], "2026-04-20 10:00 UTC") {
		t.Fatalf("expected the review thread timestamp to stay on the metadata line, actual %q", detailView.BufferLines()[authorLineIndex])
	}
}

func TestReviewMode_GivenInlineCommentCodeFence_WhenRendering_ThenItKeepsSyntaxColorsAndFillsTheCommentBoxInterior(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "```go\nfunc render(value int) string {\n\treturn fmt.Sprintf(\"%d\", value + 42)\n}\n```",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	codeStartLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "func render")
	codeLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, `return fmt.Sprintf("%d", value + 42)`)
	codeEndLineIndex := given_viewLineIndexContainingCommentBoxText(t, detailView, "}")
	if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[codeStartLineIndex-1])); actualInnerText != "" {
		t.Fatalf("expected the review inline comment code block top padding line to stay blank inside the comment box, actual %q", actualInnerText)
	}
	if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, detailView.BufferLines()[codeEndLineIndex+1])); actualInnerText != "" {
		t.Fatalf("expected the review inline comment code block bottom padding line to stay blank inside the comment box, actual %q", actualInnerText)
	}
	then_viewCommentBoxInteriorHasBackgroundColor(t, gui, viewDetailName, codeLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "review inline comment code block background")
	then_viewCommentBoxInteriorHasBackgroundColor(t, gui, viewDetailName, codeStartLineIndex-1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "review inline comment code block top padding background")
	then_viewCommentBoxInteriorHasBackgroundColor(t, gui, viewDetailName, codeEndLineIndex+1, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "review inline comment code block bottom padding background")
	then_viewCommentBoxBorderDoesNotHaveBackgroundColor(t, gui, viewDetailName, codeLineIndex, given_themeColorHex(t, theme.SelectedLineBackgroundHex), "review inline comment code block border background")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, codeLineIndex, "return", given_themeColorHex(t, theme.SyntaxKeywordHex), "review inline comment code keyword")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, codeLineIndex, "Sprintf", given_themeColorHex(t, theme.SyntaxFunctionHex), "review inline comment code function")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, codeLineIndex, `"%d"`, given_themeColorHex(t, theme.SyntaxStringHex), "review inline comment code string")
	then_viewLineSegmentHasForegroundColor(t, gui, viewDetailName, codeLineIndex, "42", given_themeColorHex(t, theme.SyntaxNumberHex), "review inline comment code number")
}

func TestReviewMode_GivenAnExpandedInlineConversation_WhenRendering_ThenItShowsTheThreadChevronAndTheRightSideLineNumber(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headerLineIndex := given_viewLineIndexContaining(t, detailView, " Comment on line R3")
	if strings.Contains(detailView.BufferLines()[headerLineIndex], "│") {
		t.Fatalf("expected the conversation header to hide the diff gutter, actual %q", detailView.BufferLines()[headerLineIndex])
	}
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the expanded thread body to stay visible, actual %q", detailView.Buffer())
	}
	commentBoxLineIndex := given_viewLineIndexContaining(t, detailView, "@reviewer-one")
	if strings.Contains(detailView.BufferLines()[commentBoxLineIndex], " : ") || strings.Contains(detailView.BufferLines()[commentBoxLineIndex], "│ │") {
		t.Fatalf("expected the comment box line to hide the diff gutter, actual %q", detailView.BufferLines()[commentBoxLineIndex])
	}
	if strings.Contains(detailView.Buffer(), "Conversation") {
		t.Fatalf("expected the old conversation label to disappear, actual %q", detailView.Buffer())
	}
	if strings.TrimSpace(detailView.BufferLines()[headerLineIndex-2]) != "" {
		t.Fatalf("expected a blank separator line above the conversation, actual %q", detailView.BufferLines()[headerLineIndex-2])
	}
	if !strings.HasPrefix(detailView.BufferLines()[headerLineIndex-1], "────") || strings.Contains(detailView.BufferLines()[headerLineIndex-1], "│") {
		t.Fatalf("expected a full-width top border above the conversation, actual %q", detailView.BufferLines()[headerLineIndex-1])
	}
	blankLineIndex := headerLineIndex + 1
	for blankLineIndex < len(detailView.BufferLines()) && strings.TrimSpace(detailView.BufferLines()[blankLineIndex]) != "" {
		blankLineIndex++
	}
	if blankLineIndex >= len(detailView.BufferLines()) {
		t.Fatalf("expected a blank separator line below the conversation, actual %q", strings.Join(detailView.BufferLines(), "\n"))
	}
	if !strings.HasPrefix(detailView.BufferLines()[blankLineIndex-1], "────") || strings.Contains(detailView.BufferLines()[blankLineIndex-1], "│") {
		t.Fatalf("expected a full-width bottom border below the conversation, actual %q", detailView.BufferLines()[blankLineIndex-1])
	}
}

func TestReviewMode_GivenAResolvedInlineConversation_WhenRendering_ThenItStartsCollapsedAndShowsTheLeftSideLineNumber(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:           "thread-1",
		IsResolved:   true,
		Path:         "internal/tui/render.go",
		OriginalLine: 2,
		DiffSide:     "LEFT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	headerLineIndex := given_viewLineIndexContaining(t, detailView, " Comment on line L2 · resolved")
	if strings.Contains(detailView.BufferLines()[headerLineIndex], "│") {
		t.Fatalf("expected the collapsed conversation header to hide the diff gutter, actual %q", detailView.BufferLines()[headerLineIndex])
	}
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the resolved thread body to stay hidden while collapsed, actual %q", detailView.Buffer())
	}
	if strings.TrimSpace(detailView.BufferLines()[headerLineIndex-2]) != "" {
		t.Fatalf("expected a blank separator line above the collapsed conversation, actual %q", detailView.BufferLines()[headerLineIndex-2])
	}
	if !strings.HasPrefix(detailView.BufferLines()[headerLineIndex-1], "────") || strings.Contains(detailView.BufferLines()[headerLineIndex-1], "│") {
		t.Fatalf("expected a full-width top border above the collapsed conversation, actual %q", detailView.BufferLines()[headerLineIndex-1])
	}
	if !strings.HasPrefix(detailView.BufferLines()[headerLineIndex+1], "────") || strings.Contains(detailView.BufferLines()[headerLineIndex+1], "│") {
		t.Fatalf("expected a full-width bottom border below the collapsed conversation, actual %q", detailView.BufferLines()[headerLineIndex+1])
	}
	if strings.TrimSpace(detailView.BufferLines()[headerLineIndex+2]) != "" {
		t.Fatalf("expected a blank separator line below the collapsed conversation, actual %q", detailView.BufferLines()[headerLineIndex+2])
	}
}

func TestReviewMode_GivenTheCursorOnAnInlineConversation_WhenPressingEnter_ThenItTogglesTheConversationVisibility(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Comment on line R3")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3") {
		t.Fatalf("expected the thread to collapse after pressing enter, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the collapsed thread body to be hidden, actual %q", detailView.Buffer())
	}

	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3") {
		t.Fatalf("expected the thread to expand after pressing enter again, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the expanded thread body to return, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenTheCursorOnAnInlineConversation_WhenPressingZA_ThenItTogglesTheConversationVisibility(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:       "thread-1",
		Path:     "internal/tui/render.go",
		Line:     3,
		DiffSide: "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Comment on line R3")

	prefixHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'z')
	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, 'a')
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3") {
		t.Fatalf("expected the first z to arm the motion without changing the thread, actual %q", detailView.Buffer())
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)

	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3") {
		t.Fatalf("expected za to collapse the thread, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected za to hide the thread body while collapsed, actual %q", detailView.Buffer())
	}
	then_viewDoesNotExist(t, gui, viewActionsPopupName)

	actualErr = prefixHandler(gui, detailView)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3") {
		t.Fatalf("expected za to expand the thread again, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected za to show the thread body again, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenTheCursorOnAResolvedCollapsedInlineConversation_WhenPressingEnter_ThenItExpandsAndRehidesTheConversation(t *testing.T) {
	diff := given_reviewSessionPullRequestDiff()
	diff.Threads = []githubcli.PullRequestReviewThread{{
		ID:         "thread-1",
		IsResolved: true,
		Path:       "internal/tui/render.go",
		Line:       3,
		DiffSide:   "RIGHT",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"},
			Body:      "Thread body",
			CreatedAt: "2026-04-20T10:00:00Z",
		}},
	}}
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending", diffs: map[string]githubcli.PullRequestDiff{"acme/widgets#42": diff}}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Thread body": "Rendered thread body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	actualErr = subject.focusDetailView(gui, nil)
	then_noError(t, actualErr)
	given_reviewModeDetailCursorOnLineContaining(t, gui, subject, "Comment on line R3 · resolved")

	toggleHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewDetailName, gocui.KeyEnter)
	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3 · resolved") {
		t.Fatalf("expected the resolved thread to expand after pressing enter, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the resolved thread body to appear after expanding, actual %q", detailView.Buffer())
	}

	actualErr = toggleHandler(gui, detailView)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), " Comment on line R3 · resolved") {
		t.Fatalf("expected the resolved thread to re-hide after pressing enter again, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "Rendered thread body") {
		t.Fatalf("expected the resolved thread body to hide again, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenApprovalMetadata_WhenRendering_ThenTheMetadataPaneFitsAllVisibleLines(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 2,
				Reviews:      []githubcli.PullRequestReview{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"}},
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

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	if metadataView.InnerHeight() < len(metadataView.BufferLines()) {
		t.Fatalf("expected metadata pane height %d to fit %d rendered lines", metadataView.InnerHeight(), len(metadataView.BufferLines()))
	}
}

func TestReviewMode_GivenReviewMetadata_WhenRendering_ThenViewOneShowsThePullRequestContextAndColoredDiffCounts(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:          "First PR",
				Number:         42,
				Body:           "Body 42",
				Author:         &githubcli.PullRequestAuthor{Login: "octocat"},
				Labels:         []githubcli.PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
				Assignees:      []githubcli.PullRequestAuthor{{Login: "assignee-one"}},
				ReviewRequests: []githubcli.PullRequestReviewRequest{{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Slug: "platform", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}}},
				BaseRefName:    "main",
				HeadRefName:    "feature/review",
				State:          "OPEN",
				ChangedFiles:   2,
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

	metadataView, actualErr := gui.View(viewUserName)
	then_noError(t, actualErr)
	for _, expected := range []string{"First PR", detailRepositoryIcon + " acme/widgets#42", detailAuthorIcon + " @octocat", detailBranchIcon + " main ← feature/review", detailStatusIcon + " OPEN", detailLabelIcon + " bug", detailAssigneesIcon + " @assignee-one", detailReviewRequestsIcon + " @acme/platform"} {
		if !strings.Contains(metadataView.Buffer(), expected) {
			t.Fatalf("expected review metadata to contain %q, actual %q", expected, metadataView.Buffer())
		}
	}

	countsLineIndex := given_viewLineIndexContaining(t, metadataView, "Changed files:")
	then_viewLineSegmentHasForegroundColor(t, gui, viewUserName, countsLineIndex, "+3", given_themeColorHex(t, theme.DiffAdditionForegroundHex), "review metadata additions")
	then_viewLineSegmentHasForegroundColor(t, gui, viewUserName, countsLineIndex, "-2", given_themeColorHex(t, theme.DiffDeletionForegroundHex), "review metadata deletions")
}

func TestReviewMode_GivenADeepSingleFilePath_WhenRenderingTheFilesPane_ThenTheFileRowShowsOnlyTheFilename(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 1,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": given_reviewSessionSingleDeepFilePullRequestDiff(),
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

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	if !strings.Contains(filesView.Buffer(), "󰝰 content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/") {
		t.Fatalf("expected files view to contain the collapsed directory row, actual %q", filesView.Buffer())
	}
	if !strings.Contains(filesView.Buffer(), " RecommendationStoreAdapter.java") {
		t.Fatalf("expected files view to contain the file basename row, actual %q", filesView.Buffer())
	}
	if strings.Contains(filesView.Buffer(), " content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/RecommendationStoreAdapter.java") {
		t.Fatalf("expected files view to keep full paths out of file rows, actual %q", filesView.Buffer())
	}
}

func TestReviewMode_GivenTheSelectedFileTreeRow_WhenRendering_ThenViewTwoKeepsItVisiblyMarked(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 2,
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

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	selectedLineIndex := given_viewLineIndexContaining(t, filesView, "render.go")
	then_viewLineSegmentHasSelectedLineBackground(t, gui, viewPullRequestsName, selectedLineIndex, "render.go")
}

func TestBrowserMode_GivenReviewRenderingSupport_WhenRefreshingThePullRequestDetail_ThenItKeepsTheExistingCommentsTabOutput(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:       "First PR",
				Number:      42,
				Body:        "Body 42",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
				InlineComments: []githubcli.PullRequestInlineComment{{
					Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
					Body:         "Inline body",
					CreatedAt:    "2026-04-20T10:00:00Z",
					Path:         "internal/tui/render.go",
					Line:         43,
					OriginalLine: 43,
					Side:         "RIGHT",
					DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
				}},
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#42": {
				Threads: []githubcli.PullRequestReviewThread{{
					ID:         "thread-1",
					Path:       "internal/tui/render.go",
					Line:       43,
					DiffSide:   "RIGHT",
					IsResolved: true,
					Comments:   []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"}, Body: "Inline body", CreatedAt: "2026-04-20T10:00:00Z"}},
				}},
			},
		},
	}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	subject.activeDetailTab = CommentsDetailTab
	subject.markdownRenderer = &fakeMarkdownRenderer{outputs: map[string]string{"Inline body": "Rendered inline body"}}
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	then_tabsAre(t, detailView, []string{DescriptionDetailTab.Label(), CommentsDetailTab.Label()}, 1)
	if strings.Contains(detailView.Buffer(), "Conversation · resolved") {
		t.Fatalf("expected browser mode to keep the existing comments-tab formatter, actual %q", detailView.Buffer())
	}
	if !strings.Contains(detailView.Buffer(), detailInlineCommentLocationIcon+" internal/tui/render.go:43") || !strings.Contains(detailView.Buffer(), "Rendered inline body") {
		t.Fatalf("expected browser mode comments tab to remain unchanged, actual %q", detailView.Buffer())
	}
}

func TestReviewMode_GivenColoredFileTreeRows_WhenRendering_ThenDirectoriesAreGrayAndFilesReflectTheirChangeType(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#77": {
				Title:        "Colorful Tree PR",
				Number:       77,
				Body:         "Body 77",
				BaseRefName:  "main",
				HeadRefName:  "feature/review-tree-colors",
				State:        "OPEN",
				ChangedFiles: 3,
			},
		},
		diffs: map[string]githubcli.PullRequestDiff{
			"acme/widgets#77": given_reviewSessionColoredTreePullRequestDiff(),
		},
	}
	model := given_pullRequestCommentModel()
	model.SetPullRequestRows(MyPullRequestsTab, []PullRequestRow{
		myPullRequestRow(githubcli.PullRequest{Title: "Colorful Tree PR", Number: 77, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, URL: "https://github.com/acme/widgets/pull/77"}),
	})
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	filesView, actualErr := gui.View(viewPullRequestsName)
	then_noError(t, actualErr)
	directoryLineIndex := given_viewLineIndexContaining(t, filesView, "internal/tui/")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, directoryLineIndex, "󰝰", given_themeColorHex(t, theme.DiffLineNumberHex), "review tree directory icon")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, directoryLineIndex, "internal/tui/", given_themeColorHex(t, theme.ActiveTextHex), "review tree directory label")

	changedLineIndex := given_viewLineIndexContaining(t, filesView, "changed.go")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, changedLineIndex, "", given_themeColorHex(t, theme.ActiveTextHex), "review tree modified file icon")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, changedLineIndex, "changed.go", given_themeColorHex(t, theme.ActiveTextHex), "review tree modified file label")

	addedLineIndex := given_viewLineIndexContaining(t, filesView, "added.go")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, addedLineIndex, "", given_themeColorHex(t, theme.DiffAdditionForegroundHex), "review tree added file icon")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, addedLineIndex, "added.go", given_themeColorHex(t, theme.ActiveTextHex), "review tree added file label")

	deletedLineIndex := given_viewLineIndexContaining(t, filesView, "deleted.go")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, deletedLineIndex, "", given_themeColorHex(t, theme.DiffDeletionForegroundHex), "review tree deleted file icon")
	then_viewLineSegmentHasForegroundColor(t, gui, viewPullRequestsName, deletedLineIndex, "deleted.go", given_themeColorHex(t, theme.ActiveTextHex), "review tree deleted file label")
}

func TestReviewMode_GivenMovingTheViewTwoSelection_WhenRefreshingTheReviewPane_ThenViewZeroRendersTheSelectedFileDiff(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#42": {
				Title:        "First PR",
				Number:       42,
				Body:         "Body 42",
				BaseRefName:  "main",
				HeadRefName:  "feature/review",
				State:        "OPEN",
				ChangedFiles: 2,
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

	actualErr = subject.moveSelectionDown(gui, nil)
	then_noError(t, actualErr)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "internal/tui/model.go") || !strings.Contains(detailView.Buffer(), "+new model") {
		t.Fatalf("expected detail view to switch to the selected file diff, actual %q", detailView.Buffer())
	}
	if strings.Contains(detailView.Buffer(), "+another line") {
		t.Fatalf("expected the first file diff to disappear after selection changed, actual %q", detailView.Buffer())
	}
	if !reflect.DeepEqual(loader.diffCalls, []string{"acme/widgets#42"}) {
		t.Fatalf("expected diff calls to stay cached at %v, actual %v", []string{"acme/widgets#42"}, loader.diffCalls)
	}
}

func TestReviewMode_GivenTheFilesPane_WhenPressingGGOrG_ThenItMovesToTheFirstOrLastSelectableFile(t *testing.T) {
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
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

	selectableRows, ok := subject.reviewSessionSelectableRows()
	if !ok || len(selectableRows) < 2 {
		t.Fatalf("expected at least two selectable review rows, actual %v", selectableRows)
	}

	bottomHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'G')
	actualErr = bottomHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.reviewSession.selectedFileTreeRow != selectableRows[len(selectableRows)-1] {
		t.Fatalf("expected selected review row %d, actual %d", selectableRows[len(selectableRows)-1], subject.reviewSession.selectedFileTreeRow)
	}

	topHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'g')
	actualErr = topHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.reviewSession.selectedFileTreeRow != selectableRows[len(selectableRows)-1] {
		t.Fatalf("expected the first g to arm the motion without moving selection, actual %d", subject.reviewSession.selectedFileTreeRow)
	}

	actualErr = topHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.reviewSession.selectedFileTreeRow != selectableRows[0] {
		t.Fatalf("expected selected review row %d, actual %d", selectableRows[0], subject.reviewSession.selectedFileTreeRow)
	}
}

func TestReviewMode_GivenFullscreenPullRequestBrowser_WhenStartingAndExiting_ThenItUsesThreePanesInReviewModeAndRestoresThePreviousFullscreenLayout(t *testing.T) {
	model := given_pullRequestCommentModel()
	model.FocusPullRequestsView()
	model.ShrinkFocusedPane()
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(model, loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected fullscreen layout before starting review mode, actual %v", subject.model.PaneLayoutSize())
	}

	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	if subject.model.PaneLayoutSize() == PaneLayoutFullscreen {
		t.Fatalf("expected review mode to leave fullscreen and show three panes, actual %v", subject.model.PaneLayoutSize())
	}
	then_viewExists(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewExists(t, gui, viewDetailName)

	exitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyEsc)
	actualErr = exitHandler(gui, nil)
	then_noError(t, actualErr)
	if subject.model.PaneLayoutSize() != PaneLayoutFullscreen {
		t.Fatalf("expected fullscreen layout after exiting review mode, actual %v", subject.model.PaneLayoutSize())
	}
	if subject.model.FullscreenPane() != FocusPullRequestsView {
		t.Fatalf("expected fullscreen pane %v, actual %v", FocusPullRequestsView, subject.model.FullscreenPane())
	}
	then_viewDoesNotExist(t, gui, viewUserName)
	then_viewExists(t, gui, viewPullRequestsName)
	then_viewDoesNotExist(t, gui, viewDetailName)
}

func TestReviewMode_GivenAnOpenPendingReview_WhenExiting_ThenItKeepsTheReviewOpenAndShowsResumeFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	exitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyEsc)
	actualErr = exitHandler(gui, nil)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected review mode to be inactive after exiting")
	}
	then_statusLineContains(t, gui, pendingPullRequestReviewKeptOpenMessage)
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, a: Action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)

	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	if subject.reviewSession.pendingReviewID != "PRR_pending" {
		t.Fatalf("expected pending review id %q, actual %q", "PRR_pending", subject.reviewSession.pendingReviewID)
	}
}

func TestReviewMode_GivenAnOpenPendingReview_WhenPressingQFromTheFileTree_ThenItKeepsTheReviewOpenAndShowsResumeFeedback(t *testing.T) {
	loader := &fakePullRequestDetailLoader{startReviewID: "PRR_pending"}
	subject := given_pullRequestCommentProgram(given_pullRequestCommentModel(), loader)
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)

	exitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, 'q')
	actualErr = exitHandler(gui, nil)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected review mode to be inactive after exiting")
	}
	then_statusLineContains(t, gui, pendingPullRequestReviewKeptOpenMessage)
	then_statusLineKeyHintsAre(t, gui, "?: Help, /: Search, a: Action")
	then_viewDoesNotExist(t, gui, viewPullRequestsFooterName)
}

func TestReviewMode_GivenItStartedFromPullRequestDetail_WhenExiting_ThenItRestoresThePriorBrowserFocusSelectionAndDetailTab(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	model.SetPullRequestRows(RequestedPullRequestsTab, []PullRequestRow{
		requestedPullRequestRow(githubcli.PullRequest{Title: "Requested PR 1", Number: 7, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 7"}),
		requestedPullRequestRow(githubcli.PullRequest{Title: "Requested PR 2", Number: 8, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, Body: "Body 8"}),
	})
	model.NextPullRequestTab()
	model.MoveSelectionDown()
	model.OpenDetail()
	loader := &fakePullRequestDetailLoader{
		startReviewID: "PRR_pending",
		details: map[string]githubcli.PullRequestDetail{
			"acme/widgets#8": {
				Title:       "Requested PR 2",
				Number:      8,
				Body:        "Body 8",
				BaseRefName: "main",
				HeadRefName: "feature/review",
				State:       "OPEN",
			},
		},
	}
	subject := given_pullRequestCommentProgram(model, loader)
	subject.activeDetailTab = CommentsDetailTab
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = given_startingReviewMode(t, gui, subject)
	then_noError(t, actualErr)
	then_currentViewNameIs(t, gui, viewPullRequestsName)

	exitHandler := given_handlerForBinding(t, subject.keybindingSpecs(), viewPullRequestsName, gocui.KeyEsc)
	actualErr = exitHandler(gui, nil)
	then_noError(t, actualErr)

	if subject.reviewSession.active {
		t.Fatal("expected review mode to be inactive after exiting")
	}
	if subject.model.Focus() != FocusDetailView {
		t.Fatalf("expected focus %v, actual %v", FocusDetailView, subject.model.Focus())
	}
	if subject.model.ActivePullRequestTab() != RequestedPullRequestsTab {
		t.Fatalf("expected active tab %v, actual %v", RequestedPullRequestsTab, subject.model.ActivePullRequestTab())
	}
	if subject.model.SelectedPullRequestIndex(RequestedPullRequestsTab) != 1 {
		t.Fatalf("expected requested pull request selection 1, actual %d", subject.model.SelectedPullRequestIndex(RequestedPullRequestsTab))
	}
	if subject.activeDetailTab != CommentsDetailTab {
		t.Fatalf("expected detail tab %v, actual %v", CommentsDetailTab, subject.activeDetailTab)
	}
	then_currentViewNameIs(t, gui, viewDetailName)

	detailView, actualErr := gui.View(viewDetailName)
	then_noError(t, actualErr)
	if !strings.Contains(detailView.Buffer(), "No comments yet.") {
		t.Fatalf("expected browser detail content to be restored, actual %q", detailView.Buffer())
	}
}

func given_reviewSessionSingleDeepFilePullRequestDiff() githubcli.PullRequestDiff {
	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/RecommendationStoreAdapter.java b/content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/RecommendationStoreAdapter.java",
			"index 1111111..2222222 100644",
			"--- a/content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/RecommendationStoreAdapter.java",
			"+++ b/content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/RecommendationStoreAdapter.java",
			"@@ -1,1 +1,1 @@",
			"-old recommendation",
			"+new recommendation",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{
			{Path: "content/adapter/src/main/java/com/doctolib/healthcontent/adapters/recommendations/RecommendationStoreAdapter.java", ChangeType: "modified", Additions: 1, Deletions: 1},
		},
	}
}

func given_reviewSessionColoredTreePullRequestDiff() githubcli.PullRequestDiff {
	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/changed.go b/internal/tui/changed.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/changed.go",
			"+++ b/internal/tui/changed.go",
			"@@ -1,1 +1,1 @@",
			"-old changed content",
			"+new changed content",
			"diff --git a/internal/tui/added.go b/internal/tui/added.go",
			"new file mode 100644",
			"--- /dev/null",
			"+++ b/internal/tui/added.go",
			"@@ -0,0 +1,1 @@",
			"+added content",
			"diff --git a/internal/tui/deleted.go b/internal/tui/deleted.go",
			"deleted file mode 100644",
			"--- a/internal/tui/deleted.go",
			"+++ /dev/null",
			"@@ -1,1 +0,0 @@",
			"-deleted content",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{
			{Path: "internal/tui/changed.go", ChangeType: "modified", Additions: 1, Deletions: 1},
			{Path: "internal/tui/added.go", ChangeType: "added", Additions: 1, Deletions: 0},
			{Path: "internal/tui/deleted.go", ChangeType: "removed", Additions: 0, Deletions: 1},
		},
	}
}

func given_reviewSessionPullRequestDiff() githubcli.PullRequestDiff {
	return githubcli.PullRequestDiff{
		UnifiedDiff: strings.Join([]string{
			"diff --git a/internal/tui/render.go b/internal/tui/render.go",
			"index 1111111..2222222 100644",
			"--- a/internal/tui/render.go",
			"+++ b/internal/tui/render.go",
			"@@ -1,2 +1,3 @@",
			" context",
			"-old line",
			"+new line",
			"+another line",
			"diff --git a/internal/tui/model.go b/internal/tui/model.go",
			"index 3333333..4444444 100644",
			"--- a/internal/tui/model.go",
			"+++ b/internal/tui/model.go",
			"@@ -10,1 +10,1 @@",
			"-old model",
			"+new model",
		}, "\n"),
		Files: []githubcli.PullRequestDiffFile{
			{Path: "internal/tui/render.go", ChangeType: "modified", Additions: 2, Deletions: 1},
			{Path: "internal/tui/model.go", ChangeType: "modified", Additions: 1, Deletions: 1},
		},
	}
}

func given_startingReviewMode(t *testing.T, gui *gocui.Gui, subject *Program) error {
	t.Helper()

	actualErr := subject.openActionsPopup(gui, nil)
	if actualErr != nil {
		return actualErr
	}
	subject.model.UpdateActionsPopupSearch("start review", matchingActionsPopupIndexes(subject.currentActionsPopupActions(), "start review"))
	actualErr = subject.refreshViews(gui)
	if actualErr != nil {
		return actualErr
	}

	return subject.executeSelectedActionsPopupAction(gui, nil)
}
