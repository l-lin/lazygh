package tui

import (
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestRenderStyledTextLineWithWidth_GivenCodeBlockBackground_WhenRendering_ThenItKeepsTheBackgroundAcrossTheRequestedWidth(t *testing.T) {
	codePrefix := foregroundColorEscape(theme.SyntaxKeywordHex) + backgroundColorEscape(theme.SelectedLineBackgroundHex)
	actual := renderStyledTextLineWithWidth(styledTextLine{
		runes:         []rune("return"),
		stylePrefixes: []string{codePrefix, codePrefix, codePrefix, codePrefix, codePrefix, codePrefix},
	}, 12) + "|"

	actualLine := splitStyledTextLines(actual)[0]
	if actualVisible := string(actualLine.runes); actualVisible != "return      |" {
		t.Fatalf("expected visible line %q, actual %q", "return      |", actualVisible)
	}

	then_styledTextLineRuneRangeHasBackgroundHex(t, actualLine, len([]rune("return")), len(actualLine.runes)-1, theme.SelectedLineBackgroundHex, "code block padding background")
}

func TestRenderRoundedCommentBox_GivenWrappedCodeBlockLines_WhenFormatting_ThenItKeepsTheBackgroundAcrossTheCommentBoxInterior(t *testing.T) {
	codePrefix := backgroundColorEscape(theme.SelectedLineBackgroundHex)
	styledBody := strings.Join([]string{
		codePrefix + "first wrapped line" + ansiReset,
		codePrefix + "second wrapped line" + ansiReset,
	}, "\n")

	actualDocument := newDetailDocument(renderRoundedCommentBox(styledBody, 40), 40)
	firstLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "first wrapped line")
	secondLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "second wrapped line")

	then_detailDocumentCommentBoxInteriorHasBackgroundHex(t, actualDocument, firstLineIndex, theme.SelectedLineBackgroundHex, "first wrapped code line background")
	then_detailDocumentCommentBoxInteriorHasBackgroundHex(t, actualDocument, secondLineIndex, theme.SelectedLineBackgroundHex, "second wrapped code line background")
	then_detailDocumentCommentBoxBorderDoesNotHaveBackgroundHex(t, actualDocument, firstLineIndex, theme.SelectedLineBackgroundHex, "wrapped code line border background")
}

func TestRenderRoundedCommentBox_GivenALongCodeBlockLine_WhenFormatting_ThenItWrapsInsideTheRequestedWidthAndKeepsTheBackground(t *testing.T) {
	codePrefix := backgroundColorEscape(theme.SelectedLineBackgroundHex)
	styledBody := codePrefix + "wrap this code block line around the comment box width" + ansiReset

	actualDocument := newDetailDocumentWithWrap(renderRoundedCommentBox(styledBody, 30), 30, false)
	firstLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "wrap this code block line")
	secondLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "around the comment box")
	thirdLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "width")

	if firstLineIndex >= secondLineIndex || secondLineIndex >= thirdLineIndex {
		t.Fatalf("expected the long code block line to wrap across multiple visible lines, actual %q", actualDocument.text)
	}
	for _, actualLineIndex := range []int{firstLineIndex, secondLineIndex, thirdLineIndex} {
		then_detailDocumentCommentBoxInteriorHasBackgroundHex(t, actualDocument, actualLineIndex, theme.SelectedLineBackgroundHex, "wrapped long code line background")
	}
	then_detailDocumentCommentBoxBorderDoesNotHaveBackgroundHex(t, actualDocument, firstLineIndex, theme.SelectedLineBackgroundHex, "wrapped long code line border background")
}

func TestRenderCommentBoxWithMetadata_GivenALongMarkdownCodeBlockLineWithErrorTokens_WhenFormatting_ThenItDoesNotInsertBlankLinesBetweenWrappedCodeSegments(t *testing.T) {
	markdown := strings.Join([]string{
		"(Fix with Cursor)",
		"",
		"```",
		"ApplicationContext failure threshold (1) exceeded: skipping repeated attempt to load context for [WebMergedContextConfiguration@31946dd4 testClass = foo.bar.ApplicationIntegrationTest, locations = [], classes = [foo.bar.Application, foo.bar.ApplicationIntegrationTest.CustomRestTemplateBuilderConfig], contextInitializerClasses = [], activeProfiles = [\"test\"], propertySourceDescriptors = [], propertySourceProperties = [\"org.springframework.boot.test.context.SpringBootTestContextBootstrapper=true\", \"server.port=0\"], contextCustomizers = [org.springframework.boot.test.context.filter.ExcludeFilterContextCustomizer@565983f3]]",
		"```",
	}, "\n")
	body := renderMarkdownWithFallback(markdown, glamourMarkdownRenderer{}, commentBoxInnerWidth(80), "")

	actualDocument := newDetailDocumentWithWrap(renderCommentBoxWithMetadata(nil, "", nil, body, 80), 80, false)
	introLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "(Fix with Cursor)")
	firstCodeLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "ApplicationContext failure threshold")
	lastCodeLineIndex, _ := given_detailDocumentLineContaining(t, actualDocument, "ExcludeFilterContextCustomizer@565983f3]]")

	actualBlankLineCountBeforeCode := 0
	for lineIndex := introLineIndex + 1; lineIndex < firstCodeLineIndex; lineIndex++ {
		if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, string(actualDocument.lines[lineIndex]))); actualInnerText == "" {
			actualBlankLineCountBeforeCode++
		}
	}
	if actualBlankLineCountBeforeCode != 2 {
		t.Fatalf("expected exactly 2 blank lines before the wrapped code block, actual %d in %q", actualBlankLineCountBeforeCode, actualDocument.text)
	}

	actualBlankLineCountInsideCodeBlock := 0
	for lineIndex := firstCodeLineIndex + 1; lineIndex < lastCodeLineIndex; lineIndex++ {
		if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, string(actualDocument.lines[lineIndex]))); actualInnerText == "" {
			actualBlankLineCountInsideCodeBlock++
		}
	}
	if actualBlankLineCountInsideCodeBlock != 0 {
		t.Fatalf("expected wrapped code block lines to stay adjacent, actual %d blank lines in %q", actualBlankLineCountInsideCodeBlock, actualDocument.text)
	}

	for _, lineIndex := range []int{lastCodeLineIndex + 1, lastCodeLineIndex + 2} {
		if actualInnerText := strings.TrimSpace(given_commentBoxInnerText(t, string(actualDocument.lines[lineIndex]))); actualInnerText != "" {
			t.Fatalf("expected exactly 2 blank lines after the wrapped code block, actual %q", actualInnerText)
		}
	}
}

func then_styledTextLineRuneRangeHasBackgroundHex(t *testing.T, line styledTextLine, startColumn int, endColumn int, expectedHex string, label string) {
	t.Helper()

	expectedRed, expectedGreen, expectedBlue, ok := parseHexColor(expectedHex)
	if !ok {
		t.Fatalf("expected a valid hex color, actual %q", expectedHex)
	}

	for column := startColumn; column < endColumn; column++ {
		if column < 0 || column >= len(line.stylePrefixes) {
			t.Fatalf("expected %s column %d to exist in styled line %q", label, column, string(line.runes))
		}
		actualRed, actualGreen, actualBlue, ok := sgrTrueColor(line.stylePrefixes[column], 48)
		if !ok {
			t.Fatalf("expected %s prefix to include a background color at column %d, actual %q", label, column, line.stylePrefixes[column])
		}
		if !rgbMatchesWithinTolerance(actualRed, actualGreen, actualBlue, expectedRed, expectedGreen, expectedBlue, 1) {
			t.Fatalf("expected %s to be close to %q at column %d, actual rgb [%d %d %d] in %q", label, expectedHex, column, actualRed, actualGreen, actualBlue, line.stylePrefixes[column])
		}
	}
}

func then_detailDocumentCommentBoxInteriorHasBackgroundHex(t *testing.T, document detailDocument, lineIndex int, expectedHex string, label string) {
	t.Helper()

	visibleLine := string(document.lines[lineIndex])
	leftBorderIndex, rightBorderIndex := given_commentBoxInnerRuneRange(t, visibleLine)
	then_detailDocumentLineRuneRangeHasBackgroundHex(t, document, lineIndex, leftBorderIndex+1, rightBorderIndex, expectedHex, label)
}

func then_detailDocumentCommentBoxBorderDoesNotHaveBackgroundHex(t *testing.T, document detailDocument, lineIndex int, unexpectedHex string, label string) {
	t.Helper()

	visibleLine := string(document.lines[lineIndex])
	leftBorderIndex, rightBorderIndex := given_commentBoxInnerRuneRange(t, visibleLine)
	then_detailDocumentLineRuneDoesNotHaveBackgroundHex(t, document, lineIndex, leftBorderIndex, unexpectedHex, label+" left")
	then_detailDocumentLineRuneDoesNotHaveBackgroundHex(t, document, lineIndex, rightBorderIndex, unexpectedHex, label+" right")
}

func then_detailDocumentLineRuneRangeHasBackgroundHex(t *testing.T, document detailDocument, lineIndex int, startColumn int, endColumn int, expectedHex string, label string) {
	t.Helper()

	expectedRed, expectedGreen, expectedBlue, ok := parseHexColor(expectedHex)
	if !ok {
		t.Fatalf("expected a valid hex color, actual %q", expectedHex)
	}
	if lineIndex < 0 || lineIndex >= len(document.lineStylePrefixes) {
		t.Fatalf("expected line index %d in document with %d lines", lineIndex, len(document.lineStylePrefixes))
	}

	for column := startColumn; column < endColumn; column++ {
		if column < 0 || column >= len(document.lineStylePrefixes[lineIndex]) {
			t.Fatalf("expected %s column %d to exist on line %d", label, column, lineIndex)
		}
		actualRed, actualGreen, actualBlue, ok := sgrTrueColor(document.lineStylePrefixes[lineIndex][column], 48)
		if !ok {
			t.Fatalf("expected %s prefix to include a background color at line %d column %d, actual %q", label, lineIndex, column, document.lineStylePrefixes[lineIndex][column])
		}
		if !rgbMatchesWithinTolerance(actualRed, actualGreen, actualBlue, expectedRed, expectedGreen, expectedBlue, 1) {
			t.Fatalf("expected %s to be close to %q at line %d column %d, actual rgb [%d %d %d] in %q", label, expectedHex, lineIndex, column, actualRed, actualGreen, actualBlue, document.lineStylePrefixes[lineIndex][column])
		}
	}
}

func then_detailDocumentLineRuneDoesNotHaveBackgroundHex(t *testing.T, document detailDocument, lineIndex int, column int, unexpectedHex string, label string) {
	t.Helper()

	unexpectedRed, unexpectedGreen, unexpectedBlue, ok := parseHexColor(unexpectedHex)
	if !ok {
		t.Fatalf("expected a valid hex color, actual %q", unexpectedHex)
	}
	if lineIndex < 0 || lineIndex >= len(document.lineStylePrefixes) {
		t.Fatalf("expected line index %d in document with %d lines", lineIndex, len(document.lineStylePrefixes))
	}
	if column < 0 || column >= len(document.lineStylePrefixes[lineIndex]) {
		t.Fatalf("expected column %d on line %d", column, lineIndex)
	}

	actualRed, actualGreen, actualBlue, ok := sgrTrueColor(document.lineStylePrefixes[lineIndex][column], 48)
	if !ok {
		return
	}
	if rgbMatchesWithinTolerance(actualRed, actualGreen, actualBlue, unexpectedRed, unexpectedGreen, unexpectedBlue, 1) {
		t.Fatalf("expected %s to differ from %q at line %d column %d, actual rgb [%d %d %d] in %q", label, unexpectedHex, lineIndex, column, actualRed, actualGreen, actualBlue, document.lineStylePrefixes[lineIndex][column])
	}
}
