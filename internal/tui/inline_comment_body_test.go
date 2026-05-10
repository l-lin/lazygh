package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/githubcli"
	"github.com/l-lin/lazygh/internal/theme"
)

func TestPrepareInlineCommentMarkdown_GivenSuggestionFence_WhenFormatting_ThenItConvertsItToALabeledDiffFence(t *testing.T) {
	actual := prepareInlineCommentMarkdown("Please apply this:\n\n```suggestion\nfirst := 1\n\nsecond := 2\n```")

	expected := strings.Join([]string{
		"Please apply this:",
		"",
		"**Suggestion**",
		"",
		"```diff",
		"+first := 1",
		"+",
		"+second := 2",
		"```",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected prepared markdown %q, actual %q", expected, actual)
	}
}

func TestPrepareInlineCommentMarkdownWithSuggestionContext_GivenSuggestionFenceAndTargetedDiffHunk_WhenFormatting_ThenItShowsTheCurrentLinesBeforeTheSuggestedLines(t *testing.T) {
	suggestionContext := githubcli.PullRequestInlineComment{
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -43,1 +43,1 @@\n-fmt.Println(\"goodbye\")\n+fmt.Println(\"hello\")",
	}

	actual := prepareInlineCommentMarkdownWithSuggestionContext("Please apply this:\n\n```suggestion\nfmt.Println(\"bonjour\")\n```", suggestionContext)

	expected := strings.Join([]string{
		"Please apply this:",
		"",
		"**Suggestion**",
		"",
		"```diff",
		"-fmt.Println(\"hello\")",
		"+fmt.Println(\"bonjour\")",
		"```",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected prepared markdown %q, actual %q", expected, actual)
	}
}

func TestPrepareInlineCommentMarkdown_GivenLanguageSuggestionFence_WhenFormatting_ThenItConvertsItToALabeledDiffFence(t *testing.T) {
	actual := prepareInlineCommentMarkdown("```go suggestion\nfmt.Println(\"hello\")\n```")

	expected := strings.Join([]string{
		"**Suggestion**",
		"",
		"```diff",
		"+fmt.Println(\"hello\")",
		"```",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected prepared markdown %q, actual %q", expected, actual)
	}
}

func TestPrepareInlineCommentMarkdown_GivenRegularCodeFence_WhenFormatting_ThenItAddsAVisibleCodeBlockLabel(t *testing.T) {
	actual := prepareInlineCommentMarkdown("```go\nfmt.Println(\"hello\")\n```")

	expected := strings.Join([]string{
		"**Code block · go**",
		"",
		"```go",
		"fmt.Println(\"hello\")",
		"```",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected prepared markdown %q, actual %q", expected, actual)
	}
}

func TestRenderInlineCommentBody_GivenSuggestionFence_WhenRendering_ThenItUsesASuggestionPlaceholderMarkdown(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}

	actual := renderInlineCommentBody("```suggestion\nfmt.Println(\"hello\")\n```", renderer, 72)

	if actual != "Rendered inline comment" {
		t.Fatalf("expected rendered inline comment %q, actual %q", "Rendered inline comment", actual)
	}
	expectedMarkdown := strings.Join([]string{
		"**Suggestion**",
		"",
		inlineCommentSuggestionMarker(0),
	}, "\n")
	if renderer.lastMarkdown != expectedMarkdown {
		t.Fatalf("expected markdown renderer input %q, actual %q", expectedMarkdown, renderer.lastMarkdown)
	}
	if renderer.lastWidth != 72 {
		t.Fatalf("expected markdown renderer width %d, actual %d", 72, renderer.lastWidth)
	}
}

func TestRenderInlineCommentBodyForInlineComment_GivenSuggestionFence_WhenRendering_ThenItHighlightsOnlyTheChangedSegments(t *testing.T) {
	comment := githubcli.PullRequestInlineComment{
		Body:         "```suggestion\nfmt.Println(\"bonjour\")\n```",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -43,1 +43,1 @@\n-fmt.Println(\"goodbye\")\n+fmt.Println(\"hello\")",
	}

	actualDocument := newDetailDocument(renderRoundedCommentBox(renderInlineCommentBodyForInlineComment(comment, glamourMarkdownRenderer{}, 72), 72), 72)
	deletionLineIndex, deletionLine := given_detailDocumentLineContaining(t, actualDocument, `fmt.Println("hello")`)
	additionLineIndex, additionLine := given_detailDocumentLineContaining(t, actualDocument, `fmt.Println("bonjour")`)

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, "hello", backgroundColorEscape(theme.DiffDeletionHighlightBackgroundHex), "suggestion deletion changed background")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, `fmt.Println("`, backgroundColorEscape(theme.DiffDeletionHighlightBackgroundHex), "suggestion deletion unchanged background")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, `fmt.Println("`, backgroundColorEscape(theme.SelectedLineBackgroundHex), "suggestion deletion base background")

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, "bonjour", backgroundColorEscape(theme.DiffAdditionHighlightBackgroundHex), "suggestion addition changed background")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, `fmt.Println("`, backgroundColorEscape(theme.DiffAdditionHighlightBackgroundHex), "suggestion addition unchanged background")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, `fmt.Println("`, backgroundColorEscape(theme.SelectedLineBackgroundHex), "suggestion addition base background")
}
