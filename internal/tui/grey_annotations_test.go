package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestGreyAnnotations_GivenJavaAndKotlinAnnotations_WhenRenderingDiffLines_ThenItUsesOrangeForegroundAndKeepsDiffBackgrounds(t *testing.T) {
	given_greySyntaxPalette(t)
	testCases := []struct {
		path, text, segment string
	}{
		{"main.java", "@Override", "@Override"},
		{"main.java", `@Import(Config.class)`, "@Import"},
		{"main.java", "@Import(", "@Import"},
		{"main.java", "@org.junit.Test", "@org.junit.Test"},
		{"main.kt", "@Test", "@Test"},
		{"main.kt", `@ExtendWith(SpringExtension::class)`, "@ExtendWith"},
		{"main.kt", "@org.junit.Test", "@org.junit.Test"},
		{"main.kt", "@Import(", "@Import"},
		{"main.kt", "@field:Label", "@field:Label"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.path+testCase.text, func(t *testing.T) {
			for _, kind := range []reviewDiffLineKind{reviewDiffContextLine, reviewDiffAdditionLine, reviewDiffDeletionLine} {
				actual := newDetailDocument(renderReviewDiffLine(testCase.path, reviewDiffLine{Kind: kind, Text: testCase.text}, 1, nil, true), 160)
				lineIndex, line := given_detailDocumentLineContaining(t, actual, testCase.text)
				expected := foregroundColorEscape("#A55000")
				then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], line, testCase.segment, expected, "annotation foreground")
				start := given_runeIndexInString(t, line, testCase.segment)
				for offset := range len([]rune(testCase.segment)) {
					prefix := actual.lineStylePrefixes[lineIndex][start+offset]
					if strings.LastIndex(prefix, "\x1b[38;") != strings.LastIndex(prefix, expected) {
						t.Fatalf("expected effective annotation foreground %q, actual %q", expected, prefix)
					}
					if kind == reviewDiffContextLine && strings.Contains(prefix, "[48;") {
						t.Fatalf("expected no annotation background, actual %q", prefix)
					}
				}
				if kind == reviewDiffAdditionLine {
					then_greySyntaxBackgroundIsEffective(t, actual.lineStylePrefixes[lineIndex], line, testCase.segment, theme.DiffAdditionBackgroundHex)
				}
				if kind == reviewDiffDeletionLine {
					then_greySyntaxBackgroundIsEffective(t, actual.lineStylePrefixes[lineIndex], line, testCase.segment, theme.DiffDeletionBackgroundHex)
				}
			}
		})
	}
}

func TestGreyAnnotations_GivenAnnotationArguments_WhenRendering_ThenItKeepsStringForegroundsAndOtherTypes(t *testing.T) {
	given_greySyntaxPalette(t)
	for _, path := range []string{"main.java", "main.kt"} {
		actual := newDetailDocument(renderSyntaxHighlightedCode(path, `@Label("hello")`, "", nil), 160)
		lineIndex, line := given_detailDocumentLineContaining(t, actual, "hello")
		then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], line, `"hello"`, foregroundColorEscape(theme.SyntaxStringHex), "annotation string foreground")
	}
	actual := newDetailDocument(renderSyntaxHighlightedCode("main.kt", "val value: String", "", nil), 160)
	lineIndex, line := given_detailDocumentLineContaining(t, actual, "String")
	then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], line, "String", foregroundColorEscape(theme.SyntaxTypeHex), "type foreground")
}

func TestGreyAnnotations_GivenCachedAnnotations_WhenChangingTheme_ThenItUsesTheCurrentForeground(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	text := "@Override"
	_ = renderSyntaxHighlightedCode("main.java", text, "", nil)
	given_greySyntaxPalette(t)
	actual := newDetailDocument(renderSyntaxHighlightedCode("main.java", text, "", nil), 160)
	lineIndex, line := given_detailDocumentLineContaining(t, actual, text)
	then_linePrefixContainsColor(t, actual.lineStylePrefixes[lineIndex], line, text, foregroundColorEscape("#A55000"), "current annotation foreground")
	theme.ResetPalette()
	actual = newDetailDocument(renderSyntaxHighlightedCode("main.java", text, "", nil), 160)
	lineIndex, line = given_detailDocumentLineContaining(t, actual, text)
	then_linePrefixDoesNotContainColor(t, actual.lineStylePrefixes[lineIndex], line, text, foregroundColorEscape("#A55000"), "previous annotation foreground")
}

func TestGreyAnnotations_GivenMarkdownCode_WhenRendering_ThenItUsesOrangeAnnotationForegrounds(t *testing.T) {
	given_greySyntaxPalette(t)
	renderer := glamourMarkdownRenderer{}
	for _, language := range []string{"java", "kotlin"} {
		actual, actualErr := renderer.Render("```"+language+"\n@Test\nclass Example {}\n```", 100)
		then_noError(t, actualErr)
		document := newDetailDocument(actual, 100)
		lineIndex, line := given_detailDocumentLineContaining(t, document, "@Test")
		then_linePrefixContainsForegroundHex(t, document.lineStylePrefixes[lineIndex], line, "@Test", "#A55000", "markdown annotation foreground")
	}
}
