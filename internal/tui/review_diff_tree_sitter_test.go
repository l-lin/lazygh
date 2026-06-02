package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestReviewDiffTreeSitter_GivenJavaCodeDiff_WhenFormatting_ThenItUsesTreeSitterSyntaxColors(t *testing.T) {
	file := reviewDiffFile{
		Path:       "src/main/java/com/acme/VersionParser.java",
		Additions:  1,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,0 +10,1 @@",
			Lines: []reviewDiffLine{{
				Kind:      reviewDiffAdditionLine,
				Text:      `return Versions.fromString("5.1.0");`,
				RightLine: 10,
				Side:      reviewDiffLineSideRight,
			}},
		}},
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, `return Versions.fromString("5.1.0");`)

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "return", foregroundColorEscape(theme.SyntaxKeywordHex), "java keyword")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "fromString", foregroundColorEscape(theme.SyntaxFunctionHex), "java method")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, `"5.1.0"`, foregroundColorEscape(theme.SyntaxStringHex), "java string")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, `"5.1.0"`, backgroundColorEscape(theme.DiffAdditionBackgroundHex), "diff addition background")
}

func TestReviewDiffTreeSitter_GivenGoCodeDiff_WhenFormatting_ThenItUsesTreeSitterSyntaxColors(t *testing.T) {
	file := reviewDiffFile{
		Path:       "internal/tui/render.go",
		Additions:  1,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -1,0 +1,1 @@",
			Lines: []reviewDiffLine{{
				Kind:      reviewDiffAdditionLine,
				Text:      `func addedLine() int { return 2 }`,
				RightLine: 1,
				Side:      reviewDiffLineSideRight,
			}},
		}},
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, `func addedLine() int { return 2 }`)

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "func", foregroundColorEscape(theme.SyntaxKeywordHex), "go keyword")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "addedLine", foregroundColorEscape(theme.SyntaxFunctionHex), "go function")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "2", foregroundColorEscape(theme.SyntaxNumberHex), "go number")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "addedLine", backgroundColorEscape(theme.DiffAdditionBackgroundHex), "diff addition background")
}

func TestReviewDiffTreeSitter_GivenCodeDiffsWithMoreSupportedLanguages_WhenFormatting_ThenItUsesTreeSitterSyntaxColors(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		line          string
		segment       string
		expectedColor string
	}{
		{name: "javascript keyword", path: "web/app.js", line: `const answer = formatValue("42");`, segment: "const", expectedColor: theme.SyntaxKeywordHex},
		{name: "typescript type", path: "web/app.ts", line: `const answer: number = 42;`, segment: "number", expectedColor: theme.SyntaxTypeHex},
		{name: "html tag", path: "templates/index.html", line: `<div class="hero">Hello</div>`, segment: "div", expectedColor: theme.SyntaxTypeHex},
		{name: "xml tag", path: "config/widget.xml", line: `<Widget enabled="true"/>`, segment: "Widget", expectedColor: theme.SyntaxTypeHex},
		{name: "json property", path: "config/app.json", line: `{"count": 2}`, segment: "count", expectedColor: theme.SyntaxPropertyHex},
		{name: "kotlin type", path: "src/main/kotlin/App.kt", line: `val answer: String = render("x")`, segment: "String", expectedColor: theme.SyntaxTypeHex},
		{name: "lua keyword", path: "scripts/init.lua", line: `local value = render("x")`, segment: "local", expectedColor: theme.SyntaxKeywordHex},
		{name: "make include", path: "Makefile", line: `include common.mk`, segment: "include", expectedColor: theme.SyntaxKeywordHex},
		{name: "python function", path: "tools/app.py", line: `items = render(42)`, segment: "render", expectedColor: theme.SyntaxFunctionHex},
		{name: "ruby keyword", path: "app/services/render.rb", line: `return render("x")`, segment: "return", expectedColor: theme.SyntaxKeywordHex},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			file := reviewDiffFile{
				Path:       testCase.path,
				Additions:  1,
				ChangeType: reviewDiffChangeTypeModified,
				Hunks: []reviewDiffHunk{{
					Header: "@@ -1,0 +1,1 @@",
					Lines: []reviewDiffLine{{
						Kind:      reviewDiffAdditionLine,
						Text:      testCase.line,
						RightLine: 1,
						Side:      reviewDiffLineSideRight,
					}},
				}},
			}

			actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
			lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, testCase.line)

			then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, testCase.segment, foregroundColorEscape(testCase.expectedColor), testCase.name)
		})
	}
}

func TestReviewDiffTreeSitter_GivenModifiedJavaLine_WhenFormatting_ThenItUsesHardBackgroundOnlyOnChangedSegments(t *testing.T) {
	file := reviewDiffFile{
		Path:       "src/main/java/com/acme/VersionParser.java",
		Additions:  1,
		Deletions:  1,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -10,1 +10,1 @@",
			Lines: []reviewDiffLine{
				{Kind: reviewDiffDeletionLine, Text: `return Versions.fromString("5.0.1");`, LeftLine: 10, Side: reviewDiffLineSideLeft},
				{Kind: reviewDiffAdditionLine, Text: `return Versions.fromString("5.1.0");`, RightLine: 10, Side: reviewDiffLineSideRight},
			},
		}},
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
	deletionLineIndex, deletionVisibleLine := given_detailDocumentLineContaining(t, actualDocument, `return Versions.fromString("5.0.1");`)
	additionLineIndex, additionVisibleLine := given_detailDocumentLineContaining(t, actualDocument, `return Versions.fromString("5.1.0");`)

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionVisibleLine, "0.1", backgroundColorEscape(theme.DiffDeletionHighlightBackgroundHex), "deletion changed segment background")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionVisibleLine, `return Versions.fromString("5.`, backgroundColorEscape(theme.DiffDeletionHighlightBackgroundHex), "deletion unchanged segment background")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[deletionLineIndex], deletionVisibleLine, `return Versions.fromString("5.`, backgroundColorEscape(theme.DiffDeletionBackgroundHex), "deletion base background")

	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionVisibleLine, "1.0", backgroundColorEscape(theme.DiffAdditionHighlightBackgroundHex), "addition changed segment background")
	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionVisibleLine, `return Versions.fromString("5.`, backgroundColorEscape(theme.DiffAdditionHighlightBackgroundHex), "addition unchanged segment background")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[additionLineIndex], additionVisibleLine, `return Versions.fromString("5.`, backgroundColorEscape(theme.DiffAdditionBackgroundHex), "addition base background")
}

func TestReviewDiffTreeSitter_GivenUnsupportedFileExtension_WhenFormatting_ThenItFallsBackToPlainDiffColors(t *testing.T) {
	file := reviewDiffFile{
		Path:       "notes/version.custom",
		Additions:  1,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: "@@ -1,0 +1,1 @@",
			Lines: []reviewDiffLine{{
				Kind:      reviewDiffAdditionLine,
				Text:      `mysteryValue();`,
				RightLine: 1,
				Side:      reviewDiffLineSideRight,
			}},
		}},
	}

	actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, `mysteryValue();`)
	segmentIndex := given_runeIndexInString(t, visibleLine, "mysteryValue")
	actualStylePrefix := actualDocument.lineStylePrefixes[lineIndex][segmentIndex]
	expectedStylePrefix := foregroundColorEscape(theme.DiffAdditionHex) + backgroundColorEscape(theme.DiffAdditionBackgroundHex)
	if actualStylePrefix != expectedStylePrefix {
		t.Fatalf("expected unsupported file prefix %q, actual %q", expectedStylePrefix, actualStylePrefix)
	}
}

func TestRenderReviewDiffLine_GivenVeryLargeInputFile_WhenFormatting_ThenItFallsBackToPlainDiffColorsAndKeepsIntralineHighlights(t *testing.T) {
	file := given_veryLargeJavaReviewDiffFile()

	actualDocument := newDetailDocument(renderReviewDiffFile(file, nil, 160), 160)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, `return Versions.fromString("5.1.0");`)

	then_linePrefixDoesNotContainColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "fromString", foregroundColorEscape(theme.SyntaxFunctionHex), "large-file fallback syntax function")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "fromString", foregroundColorEscape(theme.DiffAdditionHex), "large-file fallback diff foreground")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, `return Versions.fromString("5.`, backgroundColorEscape(theme.DiffAdditionBackgroundHex), "large-file fallback diff background")
	then_linePrefixContainsColor(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "1.0", backgroundColorEscape(theme.DiffAdditionHighlightBackgroundHex), "large-file fallback intraline background")
}

func given_veryLargeJavaReviewDiffFile() reviewDiffFile {
	lines := []reviewDiffLine{
		{Kind: reviewDiffDeletionLine, Text: `return Versions.fromString("5.0.1");`, LeftLine: 1, Side: reviewDiffLineSideLeft},
		{Kind: reviewDiffAdditionLine, Text: `return Versions.fromString("5.1.0");`, RightLine: 1, Side: reviewDiffLineSideRight},
	}
	for lineNumber := 2; lineNumber <= reviewDiffSyntaxHighlightLargeFileLineCount; lineNumber++ {
		lines = append(lines, reviewDiffLine{Kind: reviewDiffAdditionLine, Text: fmt.Sprintf(`return Versions.fromString("%d.0.0");`, lineNumber), RightLine: lineNumber, Side: reviewDiffLineSideRight})
	}

	return reviewDiffFile{
		Path:       "src/main/java/com/acme/VersionParser.java",
		Additions:  len(lines) - 1,
		Deletions:  1,
		ChangeType: reviewDiffChangeTypeModified,
		Hunks: []reviewDiffHunk{{
			Header: fmt.Sprintf("@@ -1,1 +1,%d @@", len(lines)),
			Lines:  lines,
		}},
	}
}

func then_linePrefixContainsColor(t *testing.T, linePrefixes []string, visibleLine string, segment string, expectedColor string, label string) {
	t.Helper()

	segmentStart := given_runeIndexInString(t, visibleLine, segment)
	for offset := range len([]rune(segment)) {
		actualStylePrefix := linePrefixes[segmentStart+offset]
		if !strings.Contains(actualStylePrefix, expectedColor) {
			t.Fatalf("expected %s prefix to contain %q at offset %d, actual %q", label, expectedColor, offset, actualStylePrefix)
		}
	}
}

func then_linePrefixDoesNotContainColor(t *testing.T, linePrefixes []string, visibleLine string, segment string, unexpectedColor string, label string) {
	t.Helper()

	segmentStart := given_runeIndexInString(t, visibleLine, segment)
	for offset := range len([]rune(segment)) {
		actualStylePrefix := linePrefixes[segmentStart+offset]
		if strings.Contains(actualStylePrefix, unexpectedColor) {
			t.Fatalf("expected %s prefix to avoid %q at offset %d, actual %q", label, unexpectedColor, offset, actualStylePrefix)
		}
	}
}
