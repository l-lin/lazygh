package tui

import (
	"strings"
	"testing"

	githubdomain "github.com/l-lin/lazygh/internal/github"
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

func TestPrepareInlineCommentMarkdown_GivenRegularCodeFence_WhenFormatting_ThenItKeepsOnlyTheOriginalFence(t *testing.T) {
	actual := prepareInlineCommentMarkdown("```go\nfmt.Println(\"hello\")\n```")

	expected := strings.Join([]string{
		"```go",
		"fmt.Println(\"hello\")",
		"```",
	}, "\n")
	if actual != expected {
		t.Fatalf("expected prepared markdown %q, actual %q", expected, actual)
	}
}

func TestRenderInlineCommentBodyForInlineComment_GivenRegularCodeFence_WhenRendering_ThenItDoesNotShowACodeBlockPlaceholder(t *testing.T) {
	comment := githubcli.PullRequestInlineComment{
		Body: "```kotlin\nfun process(settings: list<notificationsetting>, frequency: notificationfrequency) {\n    settings.foreach { it.updatefrequency(frequency) }\n}\n```",
		Path: "settings/domain/src/main/kotlin/com/doctolib/patient_account/settings/domain/NotificationSettingsManager.kt",
	}

	actualDocument := newDetailDocumentWithWrap(renderRoundedCommentBox(renderInlineCommentBodyForInlineComment(comment, glamourMarkdownRenderer{}, 96), 96), 96, false)
	given_detailDocumentLineContaining(t, actualDocument, "fun process(settings")
	if strings.Contains(string(actualDocument.text), "Code block") {
		t.Fatalf("expected the inline comment body to omit the code block placeholder, actual %q", string(actualDocument.text))
	}
}

func TestRenderInlineThreadCommentBlock_GivenRegularCodeFence_WhenRendering_ThenItDoesNotShowACodeBlockPlaceholder(t *testing.T) {
	comment := githubdomain.PullRequestComment{
		Author:    &githubdomain.PullRequestCommentAuthor{Login: "reviewer-one"},
		Body:      "```kotlin\nfun process(settings: list<notificationsetting>, frequency: notificationfrequency) {\n    settings.foreach { it.updatefrequency(frequency) }\n}\n```",
		CreatedAt: "2026-04-20T10:00:00Z",
	}

	actualDocument := newDetailDocumentWithWrap(renderInlineThreadCommentBlock(comment, glamourMarkdownRenderer{}, 96, 0, 1), 96, false)
	given_detailDocumentLineContaining(t, actualDocument, "fun process(settings")
	if strings.Contains(string(actualDocument.text), "Code block") {
		t.Fatalf("expected the review thread comment to omit the code block placeholder, actual %q", string(actualDocument.text))
	}
}

func TestRenderInlineCommentBody_GivenSuggestionFence_WhenRendering_ThenItUsesASuggestionPlaceholderMarkdown(t *testing.T) {
	renderer := &fakeMarkdownRenderer{output: "Rendered inline comment"}

	actual := renderInlineCommentBody("```suggestion\nfmt.Println(\"hello\")\n```", renderer, 72)

	if actual != "Rendered inline comment" {
		t.Fatalf("expected rendered inline comment %q, actual %q", "Rendered inline comment", actual)
	}
	expectedMarkdown := inlineCommentSuggestionMarker(0)
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

func TestRenderInlineCommentBodyForInlineComment_GivenSuggestionFence_WhenRendering_ThenItUsesDiffForegroundColorsInsteadOfTreeSitterSyntaxColors(t *testing.T) {
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

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, "Println", foregroundColorEscape(theme.DiffDeletionHex), "suggestion deletion diff foreground")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, "Println", foregroundColorEscape(theme.SyntaxFunctionHex), "suggestion deletion syntax function foreground")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, `"hello"`, foregroundColorEscape(theme.DiffDeletionHex), "suggestion deletion string diff foreground")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionLine, `"hello"`, foregroundColorEscape(theme.SyntaxStringHex), "suggestion deletion syntax string foreground")

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, "Println", foregroundColorEscape(theme.DiffAdditionHex), "suggestion addition diff foreground")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, "Println", foregroundColorEscape(theme.SyntaxFunctionHex), "suggestion addition syntax function foreground")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, `"bonjour"`, foregroundColorEscape(theme.DiffAdditionHex), "suggestion addition string diff foreground")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionLine, `"bonjour"`, foregroundColorEscape(theme.SyntaxStringHex), "suggestion addition syntax string foreground")
}

func TestRenderInlineCommentBodyForInlineComment_GivenSuggestionFence_WhenRendering_ThenItKeepsNoVisibleBlankLineBeforeTheDiff(t *testing.T) {
	comment := githubcli.PullRequestInlineComment{
		Body:         "```suggestion\nfmt.Println(\"bonjour\")\n```",
		Path:         "internal/tui/render.go",
		Line:         43,
		OriginalLine: 43,
		Side:         "RIGHT",
		DiffHunk:     "@@ -43,1 +43,1 @@\n-fmt.Println(\"goodbye\")\n+fmt.Println(\"hello\")",
	}

	actualDocument := newDetailDocument(renderRoundedCommentBox(renderInlineCommentBodyForInlineComment(comment, glamourMarkdownRenderer{}, 72), 72), 72)
	labelLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "Suggestion")
	deletionLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, `fmt.Println("hello")`)

	actualBlankLineCount := 0
	for lineIndex := labelLineIndex + 1; lineIndex < deletionLineIndex; lineIndex++ {
		if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, string(actualDocument.lines[lineIndex]))); actualInnerText == "" {
			actualBlankLineCount++
		}
	}
	if actualBlankLineCount != 0 {
		t.Fatalf("expected no blank line between the suggestion label and the first diff line, actual %d in %q", actualBlankLineCount, actualDocument.text)
	}
}

func TestRenderInlineCommentBodyForInlineComment_GivenARealWorldLongSuggestionFence_WhenRendering_ThenItWrapsTheSuggestionLinesWithoutExtraPadding(t *testing.T) {
	comment := githubcli.PullRequestInlineComment{
		Body:         "```suggestion\n- [ ] 21.10 Rename `observability` packages → `foo.bar.infrastructure.observability.*`; rename `s3-assets` → `foo.bar.s3assets.*`; rename `scheduled-jobs` → `foo.bar.scheduled_jobs.*`; rename `http-clients` → `foo.bar.http_clients.*`; update all imports repo-wide\n```",
		Path:         "openspec/changes/refactor-shared-modules/tasks.md",
		Line:         218,
		OriginalLine: 218,
		Side:         "RIGHT",
		DiffHunk:     "@@ -0,0 +218,1 @@\n+- [ ] 21.10 Rename `observability` packages → `foo.bar.*`; rename `s3-assets` → `foo.bar.s3_assets.*`; rename `scheduled-jobs` → `foo.bar.scheduled_jobs.*`; rename `http-clients` → `foo.bar.http_clients.*`; update all imports repo-wide",
	}

	actualDocument := newDetailDocumentWithWrap(renderRoundedCommentBox(renderInlineCommentBodyForInlineComment(comment, glamourMarkdownRenderer{}, 60), 60), 60, false)
	labelLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "Suggestion")
	firstDeletionLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "-- [ ] 21.10 Rename")
	continuedDeletionLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "s3_assets.*")
	firstAdditionLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "+- [ ] 21.10 Rename")
	continuedAdditionLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "s3assets.*")

	if labelLineIndex+1 != firstDeletionLineIndex {
		t.Fatalf("expected the first suggestion diff line to follow the label immediately, actual %q", actualDocument.text)
	}
	if firstDeletionLineIndex >= continuedDeletionLineIndex || continuedDeletionLineIndex >= firstAdditionLineIndex || firstAdditionLineIndex >= continuedAdditionLineIndex {
		t.Fatalf("expected the long suggestion lines to wrap across multiple visible lines, actual %q", actualDocument.text)
	}
}
