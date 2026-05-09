package tui

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/jesseduffield/gocui"

	"github.com/l-lin/lazygh/internal/theme"
)

func then_viewLineSegmentHasSearchHighlightBackground(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) {
	t.Helper()

	then_viewLineSegmentHasBackgroundColor(t, gui, viewName, lineIndex, segment, given_searchHighlightColorHex(t), "search highlight")
}

func then_viewLineSegmentHasSelectedLineBackground(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) {
	t.Helper()

	then_viewLineSegmentHasBackgroundColor(t, gui, viewName, lineIndex, segment, given_selectedLineColorHex(t), "selected line")
}

func then_viewLineSegmentHasBackgroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string, expected int32, label string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		_, backgroundColor, _ := actualCell.Style.Decompose()
		actual := backgroundColor.TrueColor().Hex()
		if actual != expected {
			t.Fatalf("expected %s color %#x at %s line %d offset %d, actual %#x", label, expected, viewName, lineIndex, offset, actual)
		}
	}
}

func then_viewLineSegmentHasForegroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string, expected int32, label string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		foregroundColor, _, _ := actualCell.Style.Decompose()
		actual := foregroundColor.TrueColor().Hex()
		if actual != expected {
			t.Fatalf("expected %s color %#x at %s line %d offset %d, actual %#x", label, expected, viewName, lineIndex, offset, actual)
		}
	}
}

func then_viewLineSegmentDoesNotHaveForegroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string, unexpected int32, label string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		foregroundColor, _, _ := actualCell.Style.Decompose()
		actual := foregroundColor.TrueColor().Hex()
		if actual == unexpected {
			t.Fatalf("expected %s color to differ from %#x at %s line %d offset %d", label, unexpected, viewName, lineIndex, offset)
		}
	}
}

func then_viewLineSegmentHasForegroundContrastAtLeast(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string, backgroundHex string, minimum float64, label string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		foregroundColor, _, _ := actualCell.Style.Decompose()
		actual := foregroundColor.TrueColor().Hex()
		actualHex := fmt.Sprintf("#%06X", actual)
		if actualContrast := foregroundContrastRatio(actualHex, backgroundHex); actualContrast < minimum {
			t.Fatalf("expected %s contrast at least %.2f at %s line %d offset %d, actual %.2f with foreground %s on %s", label, minimum, viewName, lineIndex, offset, actualContrast, actualHex, backgroundHex)
		}
	}
}

func then_viewLineHasBackgroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, expected int32, label string) {
	t.Helper()

	actualErr := gui.ForceLayoutAndRedraw()
	then_noError(t, actualErr)

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	x0, y0, _, _, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)

	screen, ok := gocui.Screen.(tcell.SimulationScreen)
	if !ok {
		t.Fatal("expected a simulation screen")
	}

	cells, width, _ := screen.GetContents()
	for offset := range view.InnerWidth() {
		actualCell := cells[((y0+1+lineIndex)*width)+(x0+1+offset)]
		_, backgroundColor, _ := actualCell.Style.Decompose()
		actual := backgroundColor.TrueColor().Hex()
		if actual != expected {
			t.Fatalf("expected %s color %#x at %s line %d offset %d, actual %#x", label, expected, viewName, lineIndex, offset, actual)
		}
	}
}

func then_viewLineSegmentIsUnderlined(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		_, _, attributes := actualCell.Style.Decompose()
		if attributes&tcell.AttrUnderline == 0 {
			t.Fatalf("expected underlined text at %s line %d offset %d, actual attributes %#x", viewName, lineIndex, offset, attributes)
		}
	}
}

func then_viewLineSegmentIsNotUnderlined(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		_, _, attributes := actualCell.Style.Decompose()
		if attributes&tcell.AttrUnderline != 0 {
			t.Fatalf("expected non-underlined text at %s line %d offset %d, actual attributes %#x", viewName, lineIndex, offset, attributes)
		}
	}
}

func then_viewLineSegmentIsBold(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) {
	t.Helper()

	cells, width, x, y := given_screenCellsForViewSegment(t, gui, viewName, lineIndex, segment)
	for offset := range utf8.RuneCountInString(segment) {
		actualCell := cells[(y*width)+(x+offset)]
		_, _, attributes := actualCell.Style.Decompose()
		if attributes&tcell.AttrBold == 0 {
			t.Fatalf("expected bold text at %s line %d offset %d, actual attributes %#x", viewName, lineIndex, offset, attributes)
		}
	}
}

func then_viewFooterIsRenderedOnBottomBorder(t *testing.T, gui *gocui.Gui, viewName string, expected string) {
	t.Helper()

	actualErr := gui.ForceLayoutAndRedraw()
	then_noError(t, actualErr)

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	if view.Footer != expected {
		t.Fatalf("expected view %q footer %q, actual %q", viewName, expected, view.Footer)
	}

	x0, _, x1, y1, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)
	screen, ok := gocui.Screen.(tcell.SimulationScreen)
	if !ok {
		t.Fatal("expected a simulation screen")
	}
	cells, width, _ := screen.GetContents()
	startX := x1 - 1 - utf8.RuneCountInString(expected)
	if startX < x0 {
		t.Fatalf("expected footer %q to fit in view %q", expected, viewName)
	}

	for offset, expectedRune := range expected {
		actualCell := cells[(y1*width)+(startX+offset)]
		if len(actualCell.Runes) == 0 || actualCell.Runes[0] != expectedRune {
			actualRune := rune(0)
			if len(actualCell.Runes) > 0 {
				actualRune = actualCell.Runes[0]
			}
			t.Fatalf("expected footer rune %q at %s bottom border offset %d, actual %q", string(expectedRune), viewName, offset, string(actualRune))
		}
	}
}

func given_screenCellsForViewSegment(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, segment string) ([]tcell.SimCell, int, int, int) {
	t.Helper()

	actualErr := gui.ForceLayoutAndRedraw()
	then_noError(t, actualErr)

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)

	line, ok := view.Line(lineIndex)
	if !ok {
		t.Fatalf("expected view %q to have line %d", viewName, lineIndex)
	}

	byteStart := strings.Index(line, segment)
	if byteStart < 0 {
		t.Fatalf("expected view %q line %d to contain segment %q, actual %q", viewName, lineIndex, segment, line)
	}

	runeStart := utf8.RuneCountInString(line[:byteStart])
	x0, y0, _, _, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)

	screen, ok := gocui.Screen.(tcell.SimulationScreen)
	if !ok {
		t.Fatal("expected a simulation screen")
	}

	cells, width, _ := screen.GetContents()
	return cells, width, x0 + 1 + runeStart, y0 + 1 + lineIndex
}

func then_viewLineRuneRangeHasBackgroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, startRune int, endRune int, expected int32, label string) {
	t.Helper()

	actualErr := gui.ForceLayoutAndRedraw()
	then_noError(t, actualErr)

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	line, ok := view.Line(lineIndex)
	if !ok {
		t.Fatalf("expected view %q to have line %d", viewName, lineIndex)
	}
	lineRunes := []rune(line)
	if startRune < 0 || endRune > len(lineRunes) || startRune >= endRune {
		t.Fatalf("expected rune range [%d,%d) inside line %q", startRune, endRune, line)
	}

	x0, y0, _, _, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)
	screen, ok := gocui.Screen.(tcell.SimulationScreen)
	if !ok {
		t.Fatal("expected a simulation screen")
	}

	cells, width, _ := screen.GetContents()
	for runeIndex := startRune; runeIndex < endRune; runeIndex++ {
		actualCell := cells[((y0+1+lineIndex)*width)+(x0+1+runeIndex)]
		_, backgroundColor, _ := actualCell.Style.Decompose()
		actual := backgroundColor.TrueColor().Hex()
		if actual != expected {
			t.Fatalf("expected %s color %#x at %s line %d rune %d, actual %#x", label, expected, viewName, lineIndex, runeIndex, actual)
		}
	}
}

func then_viewLineRuneDoesNotHaveBackgroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, runeIndex int, unexpected int32, label string) {
	t.Helper()

	actualErr := gui.ForceLayoutAndRedraw()
	then_noError(t, actualErr)

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	line, ok := view.Line(lineIndex)
	if !ok {
		t.Fatalf("expected view %q to have line %d", viewName, lineIndex)
	}
	lineRunes := []rune(line)
	if runeIndex < 0 || runeIndex >= len(lineRunes) {
		t.Fatalf("expected rune index %d inside line %q", runeIndex, line)
	}

	x0, y0, _, _, actualErr := gui.ViewPosition(viewName)
	then_noError(t, actualErr)
	screen, ok := gocui.Screen.(tcell.SimulationScreen)
	if !ok {
		t.Fatal("expected a simulation screen")
	}

	cells, width, _ := screen.GetContents()
	actualCell := cells[((y0+1+lineIndex)*width)+(x0+1+runeIndex)]
	_, backgroundColor, _ := actualCell.Style.Decompose()
	actual := backgroundColor.TrueColor().Hex()
	if actual == unexpected {
		t.Fatalf("expected %s to avoid color %#x at %s line %d rune %d, actual %#x", label, unexpected, viewName, lineIndex, runeIndex, actual)
	}
}

func given_commentBoxInnerRuneRange(t *testing.T, line string) (int, int) {
	t.Helper()

	lineRunes := []rune(line)
	leftBorderIndex := -1
	rightBorderIndex := -1
	for index, character := range lineRunes {
		if character != '│' {
			continue
		}
		if leftBorderIndex < 0 {
			leftBorderIndex = index
		}
		rightBorderIndex = index
	}
	if leftBorderIndex < 0 || rightBorderIndex <= leftBorderIndex {
		t.Fatalf("expected line %q to contain a comment box interior", line)
	}
	return leftBorderIndex, rightBorderIndex
}

func given_commentBoxInnerText(t *testing.T, line string) string {
	t.Helper()

	leftBorderIndex, rightBorderIndex := given_commentBoxInnerRuneRange(t, line)
	return string([]rune(line)[leftBorderIndex+1 : rightBorderIndex])
}

func given_viewLineIndexContainingCommentBoxText(t *testing.T, view *gocui.View, expected string) int {
	t.Helper()

	for index, line := range view.BufferLines() {
		if strings.Contains(line, expected) && strings.Count(line, "│") >= 2 {
			return index
		}
	}

	t.Fatalf("expected a comment box line to contain %q, actual %q", expected, view.Buffer())
	return 0
}

func then_viewCommentBoxInteriorHasBackgroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, expected int32, label string) {
	t.Helper()

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	line, ok := view.Line(lineIndex)
	if !ok {
		t.Fatalf("expected view %q to have line %d", viewName, lineIndex)
	}
	leftBorderIndex, rightBorderIndex := given_commentBoxInnerRuneRange(t, line)
	then_viewLineRuneRangeHasBackgroundColor(t, gui, viewName, lineIndex, leftBorderIndex+1, rightBorderIndex, expected, label)
}

func then_viewCommentBoxBorderDoesNotHaveBackgroundColor(t *testing.T, gui *gocui.Gui, viewName string, lineIndex int, unexpected int32, label string) {
	t.Helper()

	view, actualErr := gui.View(viewName)
	then_noError(t, actualErr)
	line, ok := view.Line(lineIndex)
	if !ok {
		t.Fatalf("expected view %q to have line %d", viewName, lineIndex)
	}
	leftBorderIndex, rightBorderIndex := given_commentBoxInnerRuneRange(t, line)
	then_viewLineRuneDoesNotHaveBackgroundColor(t, gui, viewName, lineIndex, leftBorderIndex, unexpected, label+" left")
	then_viewLineRuneDoesNotHaveBackgroundColor(t, gui, viewName, lineIndex, rightBorderIndex, unexpected, label+" right")
}

func given_searchHighlightColorHex(t *testing.T) int32 {
	t.Helper()

	return given_themeColorHex(t, theme.SearchHighlightHex)
}

func given_selectedLineColorHex(t *testing.T) int32 {
	t.Helper()

	return given_themeColorHex(t, theme.SelectedLineBackgroundHex)
}

func given_themeColorHex(t *testing.T, hexColor string) int32 {
	t.Helper()

	red, green, blue, ok := parseHexColor(hexColor)
	if !ok {
		t.Fatalf("expected a valid theme color, actual %q", hexColor)
	}

	return int32((red << 16) | (green << 8) | blue)
}
