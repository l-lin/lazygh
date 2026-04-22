package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/jesseduffield/gocui"

	"codeberg.org/l-lin/lazygh/internal/theme"
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
