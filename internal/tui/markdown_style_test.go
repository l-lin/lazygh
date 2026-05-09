package tui

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/l-lin/lazygh/internal/theme"
)

func TestUseDarkMarkdownStyle_GivenLightActiveText_WhenDeciding_ThenItPrefersTheDarkBase(t *testing.T) {
	actual := useDarkMarkdownStyle("#C0CAF5")

	if !actual {
		t.Fatal("expected the markdown renderer to prefer the dark base style")
	}
}

func TestUseDarkMarkdownStyle_GivenDarkActiveText_WhenDeciding_ThenItPrefersTheLightBase(t *testing.T) {
	actual := useDarkMarkdownStyle("#343B58")

	if actual {
		t.Fatal("expected the markdown renderer to prefer the light base style")
	}
}

func TestGlamourMarkdownRenderer_GivenAFirstLevelHeadingBackgroundOverride_WhenRendering_ThenItUsesTheConfiguredThemeColors(t *testing.T) {
	const (
		expectedHeadingHex           = "#7AA2F7"
		expectedHeadingBackgroundHex = "#223249"
	)

	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.Palette{
		MarkdownHeadingHex:           expectedHeadingHex,
		MarkdownHeadingBackgroundHex: expectedHeadingBackgroundHex,
	})
	renderer := glamourMarkdownRenderer{}

	actual, actualErr := renderer.Render("# Ship it", 40)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 40)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, "Ship it")
	then_linePrefixContainsBackgroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "Ship it", expectedHeadingBackgroundHex, "markdown h1 background")
	then_linePrefixContainsForegroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "Ship it", expectedHeadingHex, "markdown h1 foreground")
}

func TestGlamourMarkdownRenderer_GivenASecondLevelHeading_WhenRendering_ThenItKeepsTheDefaultGlamourPrefixWithThemeColor(t *testing.T) {
	renderer := glamourMarkdownRenderer{}

	actual, actualErr := renderer.Render("## Why\n\nParagraph body", 40)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 40)
	actualHeading := string(actualDocument.lines[0])
	if actualHeading != "## Why" {
		t.Fatalf("expected visible heading %q, actual %q", "## Why", actualHeading)
	}
	then_linePrefixContainsForegroundHex(t, actualDocument.lineStylePrefixes[0], actualHeading, "Why", theme.MarkdownHeadingHex, "markdown h2 foreground")
}

func TestGlamourMarkdownRenderer_GivenThemeSyntaxPalette_WhenRenderingCodeFence_ThenItUsesThemeSyntaxColors(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	theme.ApplyPalette(theme.Palette{
		ActiveTextHex:             "#EDEDED",
		InactiveTextHex:           "#CDCDCD",
		InactiveBorderHex:         "#7F7F7F",
		InactiveTitleHex:          "#6F6F6F",
		SelectedLineBackgroundHex: "#232323",
		MarkdownHeadingHex:        "#4A90E2",
		MarkdownLinkHex:           "#2EC4B6",
		MarkdownCodeHex:           "#FF9F1C",
		SyntaxKeywordHex:          "#F15BB5",
		SyntaxFunctionHex:         "#00BBF9",
		SyntaxTypeHex:             "#9B5DE5",
		SyntaxPropertyHex:         "#00F5D4",
		SyntaxStringHex:           "#70E000",
		SyntaxNumberHex:           "#FFD166",
		SyntaxCommentHex:          "#8E9AAF",
		DiffAdditionHex:           "#80ED99",
		DiffDeletionHex:           "#FF6B6B",
	})
	renderer := glamourMarkdownRenderer{}

	actual, actualErr := renderer.Render("```go\nfunc render(value int) string {\n\treturn fmt.Sprintf(\"%d\", value + 42)\n}\n```", 80)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 80)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, `return fmt.Sprintf("%d", value + 42)`)
	then_linePrefixContainsForegroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "return", theme.SyntaxKeywordHex, "markdown code keyword")
	then_linePrefixContainsForegroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "Sprintf", theme.SyntaxFunctionHex, "markdown code function")
	then_linePrefixContainsForegroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, `"%d"`, theme.SyntaxStringHex, "markdown code string")
	then_linePrefixContainsForegroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, "42", theme.SyntaxNumberHex, "markdown code number")
	then_linePrefixContainsBackgroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, `return fmt.Sprintf("%d", value + 42)`, theme.SelectedLineBackgroundHex, "markdown code background")
}

func TestGlamourMarkdownRenderer_GivenDifferentThemeRenders_WhenRenderingCodeFenceTwice_ThenItUsesTheCurrentThemeBackgroundEachTime(t *testing.T) {
	t.Cleanup(theme.ResetPalette)
	renderer := glamourMarkdownRenderer{}
	markdown := "```go\nfmt.Println(\"hi\")\n```"

	theme.ApplyPalette(theme.Palette{SelectedLineBackgroundHex: "#232323"})
	_, actualErr := renderer.Render(markdown, 80)
	then_noError(t, actualErr)

	theme.ResetPalette()
	actual, actualErr := renderer.Render(markdown, 80)

	then_noError(t, actualErr)
	actualDocument := newDetailDocument(actual, 80)
	lineIndex, visibleLine := given_detailDocumentLineContaining(t, actualDocument, `fmt.Println("hi")`)
	then_linePrefixContainsBackgroundHex(t, actualDocument.lineStylePrefixes[lineIndex], visibleLine, `fmt.Println("hi")`, theme.SelectedLineBackgroundHex, "markdown code background after palette reset")
}

func then_linePrefixContainsForegroundHex(t *testing.T, linePrefixes []string, visibleLine string, segment string, expectedHex string, label string) {
	t.Helper()
	then_linePrefixContainsTrueColorHex(t, linePrefixes, visibleLine, segment, 38, expectedHex, label)
}

func then_linePrefixContainsBackgroundHex(t *testing.T, linePrefixes []string, visibleLine string, segment string, expectedHex string, label string) {
	t.Helper()
	then_linePrefixContainsTrueColorHex(t, linePrefixes, visibleLine, segment, 48, expectedHex, label)
}

func then_linePrefixContainsTrueColorHex(t *testing.T, linePrefixes []string, visibleLine string, segment string, colorMode int, expectedHex string, label string) {
	t.Helper()

	expectedRed, expectedGreen, expectedBlue, ok := parseHexColor(expectedHex)
	if !ok {
		t.Fatalf("expected a valid hex color, actual %q", expectedHex)
	}

	segmentStart := given_runeIndexInString(t, visibleLine, segment)
	for offset := range len([]rune(segment)) {
		actualStylePrefix := linePrefixes[segmentStart+offset]
		actualRed, actualGreen, actualBlue, ok := sgrTrueColor(actualStylePrefix, colorMode)
		if !ok {
			t.Fatalf("expected %s prefix to include truecolor mode %d at offset %d, actual %q", label, colorMode, offset, actualStylePrefix)
		}
		if !rgbMatchesWithinTolerance(actualRed, actualGreen, actualBlue, expectedRed, expectedGreen, expectedBlue, 1) {
			t.Fatalf("expected %s prefix to be close to %q at offset %d, actual rgb [%d %d %d] in %q", label, expectedHex, offset, actualRed, actualGreen, actualBlue, actualStylePrefix)
		}
	}
}

func sgrTrueColor(prefix string, colorMode int) (int, int, int, bool) {
	sequencePrefix := fmt.Sprintf("%d;2;", colorMode)
	sequenceIndex := strings.LastIndex(prefix, sequencePrefix)
	if sequenceIndex < 0 {
		return 0, 0, 0, false
	}

	components := strings.Split(prefix[sequenceIndex+len(sequencePrefix):], ";")
	if len(components) < 3 {
		return 0, 0, 0, false
	}

	red, ok := leadingDecimal(components[0])
	if !ok {
		return 0, 0, 0, false
	}
	green, ok := leadingDecimal(components[1])
	if !ok {
		return 0, 0, 0, false
	}
	blue, ok := leadingDecimal(components[2])
	if !ok {
		return 0, 0, 0, false
	}

	return red, green, blue, true
}

func leadingDecimal(value string) (int, bool) {
	endIndex := 0
	for endIndex < len(value) && value[endIndex] >= '0' && value[endIndex] <= '9' {
		endIndex++
	}
	if endIndex == 0 {
		return 0, false
	}

	actual, actualErr := strconv.Atoi(value[:endIndex])
	if actualErr != nil {
		return 0, false
	}

	return actual, true
}

func rgbMatchesWithinTolerance(actualRed int, actualGreen int, actualBlue int, expectedRed int, expectedGreen int, expectedBlue int, tolerance int) bool {
	return channelMatchesWithinTolerance(actualRed, expectedRed, tolerance) && channelMatchesWithinTolerance(actualGreen, expectedGreen, tolerance) && channelMatchesWithinTolerance(actualBlue, expectedBlue, tolerance)
}

func channelMatchesWithinTolerance(actual int, expected int, tolerance int) bool {
	delta := actual - expected
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance
}
