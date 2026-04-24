package tui

import (
	"strings"
	"testing"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestRenderReviewDiffFile_GivenJavaCodeDiff_WhenFormatting_ThenItUsesTreeSitterSyntaxColors(t *testing.T) {
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

func TestRenderReviewDiffFile_GivenUnsupportedFileExtension_WhenFormatting_ThenItFallsBackToPlainDiffColors(t *testing.T) {
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
	expectedStylePrefix := foregroundColorEscape(theme.DiffAdditionForegroundHex) + backgroundColorEscape(theme.DiffAdditionBackgroundHex)
	if actualStylePrefix != expectedStylePrefix {
		t.Fatalf("expected unsupported file prefix %q, actual %q", expectedStylePrefix, actualStylePrefix)
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
