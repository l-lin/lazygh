package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestGreySyntax_GivenContextCode_WhenRendering_ThenItHighlightsDefinitionsStringsAndNumbers(t *testing.T) {
	given_greySyntaxPalette(t)
	testCases := []struct {
		path, text, segment, foreground, background string
	}{
		{"main.go", `func sample() string { return "hello" }`, "sample", "#5C21A5", "#EADFF3"},
		{"main.go", `func sample() string { return "hello" }`, `"hello"`, "#1C5708", ""},
		{"main.go", `func count() int { return 42 }`, "42", "#1561B8", "#DCE8F7"},
		{"main.java", `String sample() { return "hello"; }`, "sample", "#5C21A5", "#EADFF3"},
		{"main.js", `function sample() { return 42; }`, "sample", "#5C21A5", "#EADFF3"},
		{"main.ts", `function sample(): number { return 42; }`, "sample", "#5C21A5", "#EADFF3"},
		{"main.kt", `fun sample(): Int { return 42 }`, "sample", "#5C21A5", "#EADFF3"},
		{"main.lua", `function sample() return 42 end`, "sample", "#5C21A5", "#EADFF3"},
		{"main.py", "def sample():\n    return 42", "sample", "#5C21A5", "#EADFF3"},
		{"main.rb", "def sample\n  42\nend", "sample", "#5C21A5", "#EADFF3"},
		{"main.sh", `sample() { echo hello; }`, "sample", "#5C21A5", "#EADFF3"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.path+testCase.segment, func(t *testing.T) {
			actual := newDetailDocument(renderSyntaxHighlightedCode(testCase.path, testCase.text, "", nil), 160)
			lineIndex, visibleLine := given_detailDocumentLineContaining(t, actual, testCase.segment)
			then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, testCase.segment, foregroundColorEscape(testCase.foreground), "syntax foreground")
			if testCase.background == "" {
				start := given_runeIndexInString(t, visibleLine, testCase.segment)
				for offset := range len([]rune(testCase.segment)) {
					if prefix := actual.lineStylePrefixes[lineIndex][start+offset]; strings.Contains(prefix, "[48;") {
						t.Fatalf("expected no string background, actual %q", prefix)
					}
				}
			} else {
				then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, testCase.segment, backgroundColorEscape(testCase.background), "syntax background")
			}
		})
	}
}

func TestGreySyntax_GivenFunctionCallAndComment_WhenRendering_ThenItKeepsTheirBackgroundUnset(t *testing.T) {
	given_greySyntaxPalette(t)
	for _, text := range []string{`sample(42)`, `// sample comment`} {
		actual := newDetailDocument(renderSyntaxHighlightedCode("main.go", text, "", nil), 160)
		lineIndex, visibleLine := given_detailDocumentLineContaining(t, actual, "sample")
		segmentStart := given_runeIndexInString(t, visibleLine, "sample")
		if strings.Contains(actual.lineStylePrefixes[lineIndex][segmentStart], "[48;") {
			t.Fatalf("expected no syntax background, actual %q", actual.lineStylePrefixes[lineIndex][segmentStart])
		}
	}
}

func TestGreySyntax_GivenChangedCode_WhenRendering_ThenDiffBackgroundsOverrideSyntaxBackgrounds(t *testing.T) {
	given_greySyntaxPalette(t)
	for _, kind := range []reviewDiffLineKind{reviewDiffAdditionLine, reviewDiffDeletionLine} {
		expected := theme.DiffAdditionBackgroundHex
		expectedHighlight := theme.DiffAdditionHighlightBackgroundHex
		if kind == reviewDiffDeletionLine {
			expected = theme.DiffDeletionBackgroundHex
			expectedHighlight = theme.DiffDeletionHighlightBackgroundHex
		}
		text := `func sample() { sample("hello", 42) }`
		start := strings.Index(text, "hello")
		actual := newDetailDocument(renderReviewDiffLine("main.go", reviewDiffLine{Kind: kind, Text: text}, 1,
			[]styledRuneRange{{start: start, end: start + len("hello"), prefix: backgroundColorEscape(expectedHighlight)}}, true), 160)
		lineIndex, visibleLine := given_detailDocumentLineContaining(t, actual, text)
		for _, segment := range []string{"sample", `"`, "42"} {
			then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, segment, backgroundColorEscape(expected), "diff background")
		}
		then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, "hello", backgroundColorEscape(expectedHighlight), "intraline background")
		then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, "sample", foregroundColorEscape(theme.SyntaxFunctionDefinitionHex), "definition foreground")
		for _, segment := range []string{"sample", `"`, "42", "hello"} {
			expectedBackground := expected
			if segment == "hello" {
				expectedBackground = expectedHighlight
			}
			then_greySyntaxBackgroundIsEffective(t, actual.lineStylePrefixes[lineIndex], visibleLine, segment, expectedBackground)
		}
	}
}

func TestGreySyntax_GivenChangedFunctionDefinitions_WhenRendering_ThenItKeepsTheDefinitionForeground(t *testing.T) {
	given_greySyntaxPalette(t)
	testCases := []struct {
		path, text string
	}{
		{"main.go", `func sample() { sample("hello", 42) }`},
		{"main.go", `func sample() {`},
		{"main.java", `public void sample() {`},
		{"main.kt", `fun sample() {`},
		{"main.js", `function sample() {`},
		{"main.py", `def sample():`},
		{"main.rb", `def sample`},
	}
	for _, testCase := range testCases {
		t.Run(testCase.path+testCase.text, func(t *testing.T) {
			for _, kind := range []reviewDiffLineKind{reviewDiffAdditionLine, reviewDiffDeletionLine} {
				actual := newDetailDocument(renderReviewDiffLine(testCase.path, reviewDiffLine{Kind: kind, Text: testCase.text}, 1, nil, true), 160)
				lineIndex, visibleLine := given_detailDocumentLineContaining(t, actual, testCase.text)
				expectedBackground := theme.DiffAdditionBackgroundHex
				previewKind := diffPreviewAdditionLine
				if kind == reviewDiffDeletionLine {
					expectedBackground = theme.DiffDeletionBackgroundHex
					previewKind = diffPreviewDeletionLine
				}
				then_greySyntaxBackgroundIsEffective(t, actual.lineStylePrefixes[lineIndex], visibleLine, "sample", expectedBackground)
				preview := newDetailDocument(renderDiffPreviewLine(testCase.path, diffPreviewLine{kind: previewKind, text: testCase.text}, 1, nil), 160)
				previewLineIndex, previewLine := given_detailDocumentLineContaining(t, preview, testCase.text)
				expected := foregroundColorEscape(theme.SyntaxFunctionDefinitionHex)
				then_linePrefixContainsColor(t, preview.lineStylePrefixes[previewLineIndex], previewLine, "sample", expected, "preview definition foreground")
				then_greySyntaxBackgroundIsEffective(t, preview.lineStylePrefixes[previewLineIndex], previewLine, "sample", expectedBackground)
				then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, "sample", expected, "definition foreground")
				start := given_runeIndexInString(t, visibleLine, "sample")
				for offset := range len("sample") {
					prefix := actual.lineStylePrefixes[lineIndex][start+offset]
					if strings.LastIndex(prefix, "\x1b[38;") != strings.LastIndex(prefix, expected) {
						t.Fatalf("expected effective foreground %q, actual %q", expected, prefix)
					}
				}
			}
		})
	}
}

func TestGreySyntax_GivenCachedCode_WhenChangingTheme_ThenItUsesTheCurrentPalette(t *testing.T) {
	text := `func cachedGreySample() { println("cached string", 456) }`
	_ = renderSyntaxHighlightedCode("main.go", text, "", nil)
	given_greySyntaxPalette(t)
	actual := newDetailDocument(renderSyntaxHighlightedCode("main.go", text, "", nil), 160)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actual, text)
	then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, "cachedGreySample", backgroundColorEscape("#EADFF3"), "current theme background")
	theme.ResetPalette()
	actual = newDetailDocument(renderSyntaxHighlightedCode("main.go", text, "", nil), 160)
	lineIndex, visibleLine = given_detailDocumentLineContaining(t, actual, text)
	then_linePrefixDoesNotContainColor(t, actual.lineStylePrefixes[lineIndex], visibleLine, "cachedGreySample", backgroundColorEscape("#EADFF3"), "previous theme background")
}

func TestGreySyntax_GivenMarkdownCode_WhenRendering_ThenItUsesSyntaxBackgrounds(t *testing.T) {
	given_greySyntaxPalette(t)
	renderer := glamourMarkdownRenderer{}
	actual, actualErr := renderer.Render("```go\nfunc sample() { println(\"hello\", 42) }\n```", 100)
	then_noError(t, actualErr)
	document := newDetailDocument(actual, 100)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, document, "func sample()")
	for segment, expected := range map[string]string{"sample": "#EADFF3", `"hello"`: theme.SelectedLineBackgroundHex, "42": "#DCE8F7"} {
		then_linePrefixContainsBackgroundHex(t, document.lineStylePrefixes[lineIndex], visibleLine, segment, expected, "markdown syntax background")
	}
}

func TestGreySyntax_GivenMarkdownCodeStyle_WhenResolving_ThenItUsesSyntaxBackgrounds(t *testing.T) {
	given_greySyntaxPalette(t)
	actual := baseMarkdownStyle()
	applyMarkdownCodeBlockTheme(&actual)
	for name, pair := range map[string][2]string{
		"definition": {*actual.CodeBlock.Chroma.NameFunction.BackgroundColor, "#EADFF3"},
		"string":     {*actual.CodeBlock.Chroma.LiteralString.BackgroundColor, theme.SelectedLineBackgroundHex},
		"number":     {*actual.CodeBlock.Chroma.LiteralNumber.BackgroundColor, "#DCE8F7"},
	} {
		if pair[0] != pair[1] {
			t.Errorf("expected %s background %q, actual %q", name, pair[1], pair[0])
		}
	}
}

func then_greySyntaxBackgroundIsEffective(t *testing.T, prefixes []string, line, segment, expected string) {
	t.Helper()
	start := given_runeIndexInString(t, line, segment)
	for offset := range len([]rune(segment)) {
		actual := prefixes[start+offset]
		if strings.LastIndex(actual, "\x1b[48;") != strings.LastIndex(actual, backgroundColorEscape(expected)) {
			t.Fatalf("expected effective background %q, actual %q", expected, actual)
		}
	}
}

func given_greySyntaxPalette(t *testing.T) {
	t.Helper()
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.ResolvePaletteWithPreset("grey", theme.Palette{}))
}
