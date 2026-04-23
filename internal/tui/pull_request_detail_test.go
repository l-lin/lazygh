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
		Title:       "Add a real detail pane",
		Number:      42,
		State:       "OPEN",
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
	}

	actual := renderPullRequestDetailHeader(summary, detail)

	for _, expected := range []string{
		detailRepositoryIcon + " acme/widgets#42",
		"Add a real detail pane",
		detailBranchIcon + " main ← feature/detail",
		detailStatusIcon + " OPEN",
		detailChecksIcon + " 1 passing, 1 failing",
		detailCommentsIcon + " 1 comment",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected header to contain %q, actual %q", expected, actual)
		}
	}
}

func TestRenderPullRequestDetailContentWithSeparator_GivenHeaderAndBody_WhenFormatting_ThenItPlacesAHorizontalRuleBetweenMetadataAndContent(t *testing.T) {
	actual := renderPullRequestDetailContentWithSeparator("repo\ntitle\nmeta", "Body", 12)

	expected := strings.Join([]string{"repo", "title", "meta", "------------", "Body"}, "\n")
	if actual != expected {
		t.Fatalf("expected detail content %q, actual %q", expected, actual)
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

	actual := renderPullRequestCommentsTab(comments, nil, renderer, 60)

	for _, expected := range []string{detailCommentsIcon + " @reviewer-one", "2026-04-18 13:00 UTC", "Rendered comment one", detailCommentsIcon + " @reviewer-two", "2026-04-18 14:15 UTC", "Rendered comment two"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected comments tab to contain %q, actual %q", expected, actual)
		}
	}
}

func TestRenderPullRequestCommentsTab_GivenComments_WhenFormatting_ThenItRendersEachCommentInsideAGreyRoundedBox(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered comment one"}
	comments := []githubcli.PullRequestComment{{Author: &githubcli.PullRequestCommentAuthor{Login: "reviewer-one"}, CreatedAt: "2026-04-18T13:00:00Z", Body: "**Ship it**"}}

	actual := renderPullRequestCommentsTab(comments, nil, renderer, 30)
	actualDocument := newDetailDocument(actual, 30)

	if actualHeader := string(actualDocument.lines[0]); actualHeader != detailCommentsIcon+" @reviewer-one · 2026-04-18 13:00 UTC" {
		t.Fatalf("expected comment header %q, actual %q", detailCommentsIcon+" @reviewer-one · 2026-04-18 13:00 UTC", actualHeader)
	}
	if actualTopBorder := string(actualDocument.lines[1]); !strings.HasPrefix(actualTopBorder, "╭") || !strings.HasSuffix(actualTopBorder, "╮") {
		t.Fatalf("expected a rounded top border, actual %q", actualTopBorder)
	}
	if actualBodyLine := string(actualDocument.lines[2]); !strings.HasPrefix(actualBodyLine, "│ Rendered comment one") {
		t.Fatalf("expected boxed comment body, actual %q", actualBodyLine)
	}
	if actualBottomBorder := string(actualDocument.lines[3]); !strings.HasPrefix(actualBottomBorder, "╰") || !strings.HasSuffix(actualBottomBorder, "╯") {
		t.Fatalf("expected a rounded bottom border, actual %q", actualBottomBorder)
	}
	if actualStylePrefix := actualDocument.lineStylePrefixes[1][0]; actualStylePrefix != foregroundColorEscape(theme.InactiveBorderHex) {
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

	actual := renderPullRequestCommentsTab(nil, inlineComments, renderer, 120)
	actualDocument := newDetailDocument(actual, 120)
	locationLineIndex, locationLine := given_detailDocumentLineContaining(t, actualDocument, "internal/tui/render.go:43")

	expectedVisibleLine := detailInlineCommentLocationIcon + " internal/tui/render.go:43  +1  -1"
	if locationLine != expectedVisibleLine {
		t.Fatalf("expected inline comment location %q, actual %q", expectedVisibleLine, locationLine)
	}
	additionIndex := given_runeIndexInString(t, locationLine, "+1")
	if actualStylePrefix := actualDocument.lineStylePrefixes[locationLineIndex][additionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffAdditionForegroundHex) {
		t.Fatalf("expected inline addition count prefix %q, actual %q", foregroundColorEscape(theme.DiffAdditionForegroundHex), actualStylePrefix)
	}
	deletionIndex := given_runeIndexInString(t, locationLine, "-1")
	if actualStylePrefix := actualDocument.lineStylePrefixes[locationLineIndex][deletionIndex]; actualStylePrefix != foregroundColorEscape(theme.DiffDeletionForegroundHex) {
		t.Fatalf("expected inline deletion count prefix %q, actual %q", foregroundColorEscape(theme.DiffDeletionForegroundHex), actualStylePrefix)
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

	actual := renderPullRequestCommentsTab(nil, inlineComments, renderer, 120)
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
	if actualHeading := string(actualDocument.lines[0]); actualHeading != "Why" {
		t.Fatalf("expected visible heading %q, actual %q", "Why", actualHeading)
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
}

func (renderer *fakeMarkdownRenderer) Render(markdown string, width int) (string, error) {
	renderer.lastMarkdown = markdown
	renderer.lastWidth = width
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
