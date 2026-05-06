package tui

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"codeberg.org/l-lin/lazygh/internal/githubcli"
	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestRenderPullRequestDetailHeader_GivenRichMetadata_WhenFormatting_ThenItShowsACompactHeaderWithIcons(t *testing.T) {
	summary := githubcli.PullRequest{
		Title:      "Fallback title",
		Number:     42,
		Repository: githubcli.Repository{NameWithOwner: "acme/widgets"},
	}
	detail := githubcli.PullRequestDetail{
		Title:     "Add a real detail pane",
		Number:    42,
		Author:    &githubcli.PullRequestAuthor{Login: "octocat"},
		State:     "OPEN",
		CreatedAt: "2026-04-18T10:00:00Z",
		UpdatedAt: "2026-04-18T12:30:00Z",
		Labels:    []githubcli.PullRequestLabel{{Name: "bug"}, {Name: "backend"}},
		Assignees: []githubcli.PullRequestAuthor{{Login: "assignee-one"}, {Login: "assignee-two"}},
		ReviewRequests: []githubcli.PullRequestReviewRequest{
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "User", Login: "reviewer-requested"}},
			{RequestedReviewer: githubcli.PullRequestRequestedReviewer{TypeName: "Team", Slug: "platform", Organization: &githubcli.PullRequestReviewRequestOrganization{Login: "acme"}}},
		},
		BaseRefName: "main",
		HeadRefName: "feature/detail",
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer"},
			Body:      "Looks good",
			CreatedAt: "2026-04-18T13:00:00Z",
		}},
		StatusCheckRollup: []githubcli.PullRequestStatusCheck{
			{Name: "lint", Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Name: "test", Status: "COMPLETED", Conclusion: "FAILURE"},
		},
		Additions:    12,
		Deletions:    3,
		ChangedFiles: 5,
	}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	actual := make([]string, 0, len(actualDocument.lines))
	for _, line := range actualDocument.lines {
		actual = append(actual, string(line))
	}
	actualText := strings.Join(actual, "\n")

	for _, expected := range []string{
		detailRepositoryIcon + " acme/widgets#42",
		detailAuthorIcon + " @octocat",
		"Add a real detail pane",
		"Created: 2026-04-18 10:00 UTC",
		"Updated: 2026-04-18 12:30 UTC",
		detailBranchIcon + " main ← feature/detail",
		detailStatusIcon + " OPEN",
		detailChecksIcon + " 1 passing, 1 failing",
		detailCommentsIcon + " 1 comment",
		"+12",
		"-3",
		detailLabelIcon + " bug",
		detailLabelIcon + " backend",
		detailAssigneesIcon + " @assignee-one",
		detailAssigneesIcon + " @assignee-two",
		detailReviewRequestsIcon + " @reviewer-requested",
		detailReviewRequestsIcon + " @acme/platform",
	} {
		if !strings.Contains(actualText, expected) {
			t.Fatalf("expected header to contain %q, actual %q", expected, actualText)
		}
	}
	if strings.Contains(actualText, "  ·  ") {
		t.Fatalf("expected header metadata to avoid dot separators, actual %q", actualText)
	}
}

func TestRenderPullRequestDetailHeader_GivenInlineCommentThreadsAndRestInlineComments_WhenFormatting_ThenItCountsThreadCommentsWithoutDoubleCounting(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Number: 42,
		Comments: []githubcli.PullRequestComment{{
			Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer"},
			Body:      "General feedback",
			CreatedAt: "2026-04-18T13:00:00Z",
		}},
		InlineComments: []githubcli.PullRequestInlineComment{{Body: "First inline"}, {Body: "Second inline"}},
		InlineCommentThreads: []githubcli.PullRequestReviewThread{{
			ID:         "thread-1",
			IsResolved: true,
			Comments:   []githubcli.PullRequestComment{{Body: "First inline"}, {Body: "Second inline"}},
		}},
	}

	actual := renderPullRequestDetailHeader(summary, detail)

	if !strings.Contains(actual, detailCommentsIcon+" 3 comments") {
		t.Fatalf("expected the header to count thread comments once, actual %q", actual)
	}
}

func TestRenderPullRequestDetailHeader_GivenChurnCounts_WhenFormatting_ThenItUsesTheDiffPalette(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Number: 42, Additions: 12, Deletions: 3, ChangedFiles: 5}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	lineIndex, line := given_detailDocumentLineContaining(t, actualDocument, "+12")
	additionIndex := given_runeIndexInString(t, line, "+12")
	deletionIndex := given_runeIndexInString(t, line, "-3")

	if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][additionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffAdditionHex) {
		t.Fatalf("expected additions prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][deletionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffDeletionHex) {
		t.Fatalf("expected deletions prefix %q, actual %q", foregroundColorEscape(theme.DiffDeletionHex), actualStylePrefix)
	}
}

func TestRenderPullRequestDetailHeader_GivenSummaryOnlyUpdatedTimestamp_WhenFormatting_ThenItOmitsMissingCreatedTimeAndChurn(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, UpdatedAt: "2026-04-18T12:30:00Z"}
	detail := githubcli.PullRequestDetail{Number: 42}

	actual := renderPullRequestDetailHeader(summary, detail)

	if strings.Contains(actual, "Created:") {
		t.Fatalf("expected the header to omit the missing created timestamp, actual %q", actual)
	}
	if !strings.Contains(actual, "Updated: 2026-04-18 12:30 UTC") {
		t.Fatalf("expected the header to contain the updated timestamp, actual %q", actual)
	}
	if strings.Contains(actual, "  +0") || strings.Contains(actual, "  -0") {
		t.Fatalf("expected the header to omit placeholder churn counts, actual %q", actual)
	}
}

func TestRenderPullRequestDetailContentWithSeparator_GivenHeaderAndBody_WhenFormatting_ThenItPlacesAHorizontalRuleBetweenMetadataAndContent(t *testing.T) {
	actualDocument := newDetailDocument(renderPullRequestDetailContentWithSeparator("repo\ntitle\nmeta", "Body", 12), 12)
	actual := strings.Join([]string{string(actualDocument.lines[0]), string(actualDocument.lines[1]), string(actualDocument.lines[2]), string(actualDocument.lines[3]), string(actualDocument.lines[4])}, "\n")

	expected := strings.Join([]string{"repo", "title", "meta", "────────────", "Body"}, "\n")
	if actual != expected {
		t.Fatalf("expected detail content %q, actual %q", expected, actual)
	}
}

func TestRenderPullRequestDetailContentWithSeparator_GivenHeaderAndBody_WhenFormatting_ThenItStylesTheSeparatorLikeABorder(t *testing.T) {
	actualDocument := newDetailDocument(renderPullRequestDetailContentWithSeparator("repo", "Body", 12), 12)
	separatorLineIndex, separatorLine := given_detailDocumentLineContaining(t, actualDocument, "────────────")

	if separatorLine != "────────────" {
		t.Fatalf("expected separator line %q, actual %q", "────────────", separatorLine)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[separatorLineIndex][0]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
		t.Fatalf("expected separator border prefix %q, actual %q", foregroundColorEscape(theme.InactiveBorderHex), actualStylePrefix)
	}
}

func TestRenderPullRequestDescription_GivenMarkdownBody_WhenFormatting_ThenItUsesTheMarkdownRendererAndWrapWidth(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered markdown body"}
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Title: "Add a real detail pane", Number: 42, Body: "## Summary\n\n- render markdown"}

	actual := renderPullRequestDescription(summary, detail, renderer, 48)

	if actual != "Rendered markdown body" {
		t.Fatalf("expected rendered description %q, actual %q", "Rendered markdown body", actual)
	}
	if renderer.lastWidth != 48 {
		t.Fatalf("expected width %d, actual %d", 48, renderer.lastWidth)
	}
	if renderer.lastMarkdown != "## Summary\n\n- render markdown" {
		t.Fatalf("expected markdown %q, actual %q", "## Summary\n\n- render markdown", renderer.lastMarkdown)
	}
}

func TestRenderPullRequestCommentsTab_GivenComments_WhenFormatting_ThenItKeepsUsernamesClearlyVisible(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"**Ship it**":   "Rendered comment one",
		"Needs changes": "Rendered comment two",
	}}
	comments := []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, CreatedAt: "2026-04-18T13:00:00Z", Body: "**Ship it**"}, {Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, CreatedAt: "2026-04-18T14:15:00Z", Body: "Needs changes"}}

	actual := renderPullRequestCommentsTab(comments, nil, nil, renderer, 60)

	for _, expected := range []string{detailCommentsIcon + " @reviewer-one", "2026-04-18 13:00 UTC", "Rendered comment one", detailCommentsIcon + " @reviewer-two", "2026-04-18 14:15 UTC", "Rendered comment two"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected comments tab to contain %q, actual %q", expected, actual)
		}
	}
}

func TestRenderPullRequestCommentsTab_GivenComments_WhenFormatting_ThenItRendersEachCommentInsideAGreyRoundedBoxWithTheAuthorAndDateOnTheSameLine(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered comment one"}
	comments := []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, CreatedAt: "2026-04-18T13:00:00Z", Body: "**Ship it**"}}

	actual := renderPullRequestCommentsTab(comments, nil, nil, renderer, 60)
	actualDocument := newDetailDocument(actual, 60)

	if actualTopBorder := string(actualDocument.lines[0]); !strings.HasPrefix(actualTopBorder, "╭") || !strings.HasSuffix(actualTopBorder, "╮") {
		t.Fatalf("expected a rounded top border, actual %q", actualTopBorder)
	}
	metadataLine := string(actualDocument.lines[1])
	if !strings.HasPrefix(metadataLine, "│ ") || !strings.HasSuffix(metadataLine, " │") {
		t.Fatalf("expected the metadata line to stay inside the rounded box, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, detailCommentsIcon+" @reviewer-one") {
		t.Fatalf("expected the metadata line to contain the comment author badge, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, "2026-04-18 13:00 UTC") {
		t.Fatalf("expected the metadata line to contain the comment timestamp, actual %q", metadataLine)
	}
	if strings.Contains(metadataLine, "·") {
		t.Fatalf("expected the metadata line to avoid dot separators, actual %q", metadataLine)
	}
	if actualBodyLine := string(actualDocument.lines[2]); !strings.HasPrefix(actualBodyLine, "│ Rendered comment one") {
		t.Fatalf("expected boxed comment body, actual %q", actualBodyLine)
	}
	if actualBottomBorder := string(actualDocument.lines[3]); !strings.HasPrefix(actualBottomBorder, "╰") || !strings.HasSuffix(actualBottomBorder, "╯") {
		t.Fatalf("expected a rounded bottom border, actual %q", actualBottomBorder)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[0][0]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
		t.Fatalf("expected the comment border prefix %q, actual %q", foregroundColorEscape(theme.InactiveBorderHex), actualStylePrefix)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineComments_WhenFormatting_ThenItShowsAFileIconLineRangeAndColoredChangeCounts(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineComments := []githubcli.PullRequestInlineComment{{
		Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
		CreatedAt:    "2026-04-18T14:15:00Z",
		Body:         "Needs more spacing",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
	}}

	actual := renderPullRequestCommentsTab(nil, nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	locationLineIndex, locationLine := given_detailDocumentLineContaining(t, actualDocument, "internal/tui/render.go:43")

	expectedVisibleLine := detailInlineCommentLocationIcon + " internal/tui/render.go:43  +1  -1"
	if locationLine != expectedVisibleLine {
		t.Fatalf("expected inline comment location %q, actual %q", expectedVisibleLine, locationLine)
	}
	additionIndex := given_runeIndexInString(t, locationLine, "+1")
	if actualStylePrefix := actualDocument.lineStylePrefixes[locationLineIndex][additionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffAdditionHex) {
		t.Fatalf("expected inline addition count prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
	}
	deletionIndex := given_runeIndexInString(t, locationLine, "-1")
	if actualStylePrefix := actualDocument.lineStylePrefixes[locationLineIndex][deletionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffDeletionHex) {
		t.Fatalf("expected inline deletion count prefix %q, actual %q", foregroundColorEscape(theme.DiffDeletionHex), actualStylePrefix)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineComments_WhenFormatting_ThenItRendersABatLikeDiffPreview(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineComments := []githubcli.PullRequestInlineComment{{
		Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
		CreatedAt:    "2026-04-18T14:15:00Z",
		Body:         "Needs more spacing",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
	}}

	actual := renderPullRequestCommentsTab(nil, nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)

	_, actualHunkHeader := given_detailDocumentLineContaining(t, actualDocument, "@@ -42,2 +42,2 @@")
	if actualHunkHeader != "@@ -42,2 +42,2 @@" {
		t.Fatalf("expected diff hunk header %q, actual %q", "@@ -42,2 +42,2 @@", actualHunkHeader)
	}
	_, actualContextLine := given_detailDocumentLineContaining(t, actualDocument, "\"deny\": []")
	if actualContextLine != "42 : 42 │ \"deny\": []" {
		t.Fatalf("expected diff context line %q, actual %q", "42 : 42 │ \"deny\": []", actualContextLine)
	}
	_, actualDeletionLine := given_detailDocumentLineContaining(t, actualDocument, "\"opusplan\"")
	if actualDeletionLine != "43 :    │ \"model\": \"opusplan\"," {
		t.Fatalf("expected diff deletion line %q, actual %q", "43 :    │ \"model\": \"opusplan\",", actualDeletionLine)
	}
	_, actualAdditionLine := given_detailDocumentLineContaining(t, actualDocument, "\"opus\"")
	if actualAdditionLine != "   : 43 │ \"model\": \"opus\"," {
		t.Fatalf("expected diff addition line %q, actual %q", "   : 43 │ \"model\": \"opus\",", actualAdditionLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenInlineComments_WhenFormatting_ThenItRendersTheAuthorAndDateOnTheSameLineInsideTheCommentBox(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}
	inlineComments := []githubcli.PullRequestInlineComment{{
		Author:       &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
		CreatedAt:    "2026-04-18T14:15:00Z",
		Body:         "Needs more spacing",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
	}}

	actual := renderPullRequestCommentsTab(nil, nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	topBorderLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "╭")
	metadataLineIndex, metadataLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-inline")
	bodyLineIndex, bodyLine := given_detailDocumentLineContaining(t, actualDocument, "Rendered inline comment")

	if metadataLineIndex != topBorderLineIndex+1 {
		t.Fatalf("expected the metadata line to render inside the box immediately after the top border, actual %q", metadataLine)
	}
	if bodyLineIndex != metadataLineIndex+1 {
		t.Fatalf("expected the body line to render after the metadata line, actual %q", bodyLine)
	}
	if !strings.Contains(metadataLine, detailCommentsIcon+" @reviewer-inline") {
		t.Fatalf("expected the metadata line to contain the author badge, actual %q", metadataLine)
	}
	if !strings.Contains(metadataLine, "2026-04-18 14:15 UTC") {
		t.Fatalf("expected the metadata line to keep the timestamp on the same line, actual %q", metadataLine)
	}
	if strings.Contains(metadataLine, "·") {
		t.Fatalf("expected the inline comment metadata line to avoid dot separators, actual %q", metadataLine)
	}
	if !strings.HasPrefix(metadataLine, "│ ") || !strings.HasSuffix(metadataLine, " │") {
		t.Fatalf("expected the metadata line to render inside the rounded box, actual %q", metadataLine)
	}
}

func TestRenderPullRequestCommentsTab_GivenResolvedInlineCommentThreads_WhenFormatting_ThenItShowsTheResolvedStateAndRendersRepliesTogether(t *testing.T) {
	renderer := &fakeMarkdownRenderer{outputs: map[string]string{
		"Needs more spacing":     "Rendered inline comment",
		"Fixed in the next push": "Rendered reply",
	}}
	inlineThreads := []githubcli.PullRequestReviewThread{{
		ID:         "thread-1",
		IsResolved: true,
		Path:       "internal/tui/render.go",
		Line:       43,
		DiffSide:   "RIGHT",
		Comments: []githubcli.PullRequestComment{
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "reviewer-inline"},
				CreatedAt: "2026-04-18T14:15:00Z",
				Body:      "Needs more spacing",
				DiffHunk:  "@@ -42,2 +42,2 @@\n \"deny\": []\n-\"model\": \"opusplan\",\n+\"model\": \"opus\",",
			},
			{
				Author:    &githubcli.PullRequestCommentAuthor{Login: "octocat"},
				CreatedAt: "2026-04-18T14:45:00Z",
				Body:      "Fixed in the next push",
			},
		},
	}}

	actual := renderPullRequestCommentsTab(nil, inlineThreads, nil, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	statusLineIndex, statusLine := given_detailDocumentLineContaining(t, actualDocument, "Resolved")
	resolvedIndex := given_runeIndexInString(t, statusLine, "Resolved")

	if !strings.Contains(statusLine, "Resolved") {
		t.Fatalf("expected the thread status to mention the resolved state, actual %q", statusLine)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[statusLineIndex][resolvedIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.DiffAdditionHex)) {
		t.Fatalf("expected the resolved state to use the addition color prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
	}
	if _, replyLine := given_detailDocumentLineContaining(t, actualDocument, "Rendered reply"); !strings.Contains(replyLine, "Rendered reply") {
		t.Fatalf("expected the reply to render in the same inline thread section, actual %q", replyLine)
	}
}

func TestRenderRoundedCommentBox_GivenMultiLineStyledText_WhenFormatting_ThenItReappliesTheVisibleStyleAtEachContentLineStart(t *testing.T) {
	styledBody := foregroundColorEscape(theme.MarkdownHeadingHex) + "Styled line one\nline two" + ansiReset

	actual := renderRoundedCommentBox(styledBody, 32)
	actualDocument := newDetailDocument(actual, 32)
	if actualStylePrefix := actualDocument.lineStylePrefixes[1][2]; actualStylePrefix != foregroundColorEscape(theme.MarkdownHeadingHex) {
		t.Fatalf("expected the first content line style prefix %q, actual %q", foregroundColorEscape(theme.MarkdownHeadingHex), actualStylePrefix)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[2][2]; actualStylePrefix != foregroundColorEscape(theme.MarkdownHeadingHex) {
		t.Fatalf("expected the second content line style prefix %q, actual %q", foregroundColorEscape(theme.MarkdownHeadingHex), actualStylePrefix)
	}
}

func TestGlamourMarkdownRenderer_GivenHeadingMarkdown_WhenRendering_ThenItKeepsHeadingStyledAndDoesNotAddDocumentIndent(t *testing.T) {
	renderer := glamourMarkdownRenderer{}

	actual, actualErr := renderer.Render("## Why\n\nParagraph body", 40)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 40)
	if actualHeading := string(actualDocument.lines[0]); actualHeading != "## Why" {
		t.Fatalf("expected visible heading %q, actual %q", "## Why", actualHeading)
	}
	if actualParagraph := string(actualDocument.lines[2]); actualParagraph != "Paragraph body" {
		t.Fatalf("expected visible paragraph %q, actual %q", "Paragraph body", actualParagraph)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[0][0]; actualStylePrefix == "" {
		t.Fatal("expected the heading to keep a style prefix")
	}
}

func TestRenderPullRequestDescription_GivenMarkdownRendererFailure_WhenFormatting_ThenItFallsBackToRawMarkdown(t *testing.T) {
	renderer := &fakeMarkdownRenderer{err: errors.New("boom")}
	summary := githubcli.PullRequest{Number: 9, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{Title: "Fallback body", Number: 9, Body: "## Summary\n\n- keep the source"}

	actual := renderPullRequestDescription(summary, detail, renderer, 40)

	for _, expected := range []string{"Markdown rendering failed", "## Summary", "- keep the source"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected description to contain %q, actual %q", expected, actual)
		}
	}
}

func TestDetailStatus_GivenDraftMetadata_WhenFormatting_ThenItPrefersDRAFT(t *testing.T) {
	summary := githubcli.PullRequest{State: "OPEN"}
	detail := githubcli.PullRequestDetail{State: "MERGED", IsDraft: true}

	actual := detailStatus(detail, summary)

	if actual != "DRAFT" {
		t.Fatalf("expected status %q, actual %q", "DRAFT", actual)
	}
}

func TestRenderPullRequestDetailHeader_GivenPullRequestStatuses_WhenFormatting_ThenItUsesStateSpecificStatusBadgeBackgrounds(t *testing.T) {
	testCases := []struct {
		name                  string
		summary               githubcli.PullRequest
		detail                githubcli.PullRequestDetail
		expectedStatus        string
		expectedBackgroundHex string
	}{
		{name: "open", summary: githubcli.PullRequest{State: "OPEN"}, detail: githubcli.PullRequestDetail{State: "OPEN"}, expectedStatus: "OPEN", expectedBackgroundHex: theme.PullRequestStatusOpenBackgroundHex},
		{name: "draft", summary: githubcli.PullRequest{State: "OPEN"}, detail: githubcli.PullRequestDetail{State: "OPEN", IsDraft: true}, expectedStatus: "DRAFT", expectedBackgroundHex: theme.PullRequestStatusDraftBackgroundHex},
		{name: "closed", summary: githubcli.PullRequest{State: "CLOSED"}, detail: githubcli.PullRequestDetail{State: "CLOSED"}, expectedStatus: "CLOSED", expectedBackgroundHex: theme.PullRequestStatusClosedBackgroundHex},
		{name: "merged", summary: githubcli.PullRequest{State: "MERGED"}, detail: githubcli.PullRequestDetail{State: "MERGED"}, expectedStatus: "MERGED", expectedBackgroundHex: theme.PullRequestStatusMergedBackgroundHex},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			header := renderPullRequestDetailHeader(githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}, State: testCase.summary.State, IsDraft: testCase.summary.IsDraft}, githubcli.PullRequestDetail{Number: 42, State: testCase.detail.State, IsDraft: testCase.detail.IsDraft})
			actualDocument := newDetailDocument(header, 120)
			lineIndex, line := given_detailDocumentLineContaining(t, actualDocument, testCase.expectedStatus)
			statusIndex := given_runeIndexInString(t, line, testCase.expectedStatus)

			if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][statusIndex]; !strings.Contains(actualStylePrefix, backgroundColorEscape(testCase.expectedBackgroundHex)) {
				t.Fatalf("expected status badge prefix to contain background %q, actual %q", backgroundColorEscape(testCase.expectedBackgroundHex), actualStylePrefix)
			}
		})
	}
}

func TestRenderPullRequestDetailHeader_GivenApprovalReviews_WhenFormatting_ThenItShowsTheCurrentApproversWithGreenCheckIcons(t *testing.T) {
	summary := githubcli.PullRequest{Number: 42, Repository: githubcli.Repository{NameWithOwner: "acme/widgets"}}
	detail := githubcli.PullRequestDetail{
		Number: 42,
		Reviews: []githubcli.PullRequestReview{
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, State: "APPROVED", SubmittedAt: "2026-04-21T10:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, State: "COMMENTED", SubmittedAt: "2026-04-21T11:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-two"}, State: "APPROVED", SubmittedAt: "2026-04-21T12:00:00Z"},
			{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-three"}, State: "APPROVED", SubmittedAt: "2026-04-21T09:00:00Z"},
		},
	}

	actualDocument := newDetailDocument(renderPullRequestDetailHeader(summary, detail), 120)
	lineIndex, approvalsLine := given_detailDocumentLineContaining(t, actualDocument, "@reviewer-two")
	firstIconIndex := given_runeIndexInString(t, approvalsLine, detailApprovalIcon+" @reviewer-two")
	secondIconIndex := given_runeIndexInString(t, approvalsLine, detailApprovalIcon+" @reviewer-three")

	for _, expected := range []string{detailApprovalIcon + " @reviewer-two", detailApprovalIcon + " @reviewer-three"} {
		if !strings.Contains(approvalsLine, expected) {
			t.Fatalf("expected approvals line to contain %q, actual %q", expected, approvalsLine)
		}
	}
	if strings.Contains(approvalsLine, "@reviewer-one") {
		t.Fatalf("expected approvals line to hide reviewers whose latest review is not an approval, actual %q", approvalsLine)
	}
	for _, iconIndex := range []int{firstIconIndex, secondIconIndex} {
		if actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][iconIndex]; !strings.Contains(actualStylePrefix, foregroundColorEscape(theme.DiffAdditionHex)) {
			t.Fatalf("expected approval icon prefix to contain the addition foreground %q, actual %q", foregroundColorEscape(theme.DiffAdditionHex), actualStylePrefix)
		}
	}
}

func TestCompactBranchLabel_GivenALongBranchName_WhenFormatting_ThenItKeepsBothEndsWithAnEllipsis(t *testing.T) {
	actual := compactBranchLabel("1234567890123ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	if actual != "1234567890123…MNOPQRSTUVWXYZ" {
		t.Fatalf("expected branch label %q, actual %q", "1234567890123…MNOPQRSTUVWXYZ", actual)
	}
}

type fakeMarkdownRenderer struct {
	output       string
	outputs      map[string]string
	err          error
	lastMarkdown string
	lastWidth    int
	callCount    int
}

func (renderer *fakeMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer.lastMarkdown = markdown
	renderer.lastWidth = width
	renderer.callCount++
	if renderer.err != nil {
		return "", renderer.err
	}
	if renderer.outputs != nil {
		if output, ok := renderer.outputs[markdown]; ok {
			return output, nil
		}
	}
	return renderer.output, nil
}

func given_detailDocumentLineContaining(t *testing.T, document detailDocument, segment string) (int, string) {
	t.Helper()

	for lineIndex, line := range document.lines {
		visibleLine := string(line)
		if strings.Contains(visibleLine, segment) {
			return lineIndex, visibleLine
		}
	}

	t.Fatalf("expected detail document to contain %q, actual %q", segment, document.text)
	return -1, ""
}

func given_runeIndexInString(t *testing.T, text string, segment string) int {
	t.Helper()

	byteIndex := strings.Index(text, segment)
	if byteIndex < 0 {
		t.Fatalf("expected %q to contain %q", text, segment)
	}
	return utf8.RuneCountInString(text[:byteIndex])
}
