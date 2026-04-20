package tui

import (
	"testing"

	"codeberg.org/l-lin/lazygh/internal/theme"
)

func TestUpdatedANSIStylePrefix_GivenTrueColorBlackSequence_WhenTrackingStyles_ThenItKeepsZeroRGBComponents(t *testing.T) {
	actual := updatedANSIStylePrefix("", "\x1b[38;2;0;0;0m")

	if actual != "\x1b[38;2;0;0;0m" {
		t.Fatalf("expected style prefix %q, actual %q", "\x1b[38;2;0;0;0m", actual)
	}
}

func TestUpdatedANSIStylePrefix_GivenLeadingResetAndTrueColorSequence_WhenTrackingStyles_ThenItPreservesZeroRGBComponentsAfterReset(t *testing.T) {
	actual := updatedANSIStylePrefix("\x1b[1m", "\x1b[0;38;2;0;0;0m")

	if actual != "\x1b[38;2;0;0;0m" {
		t.Fatalf("expected reset style prefix %q, actual %q", "\x1b[38;2;0;0;0m", actual)
	}
}

func TestNewDetailDocument_GivenANSIStyledTextWithHyperlinkSequences_WhenWrapping_ThenItUsesVisibleRunesForRowsAndKeepsStylePrefixes(t *testing.T) {
	styledPrefix := foregroundColorEscape(theme.MarkdownHeadingHex)
	actual := newDetailDocument(styledPrefix+"AB"+ansiReset+"\x1b]8;id=1;https://example.com\aCD\x1b]8;;\aEF", 3)

	if actual.lineCount() != 1 {
		t.Fatalf("expected one visible line, actual %d", actual.lineCount())
	}
	if actualRowCount := actual.rowCount(); actualRowCount != 2 {
		t.Fatalf("expected two wrapped rows, actual %d", actualRowCount)
	}
	if actualVisibleLine := string(actual.lines[0]); actualVisibleLine != "ABCDEF" {
		t.Fatalf("expected visible line %q, actual %q", "ABCDEF", actualVisibleLine)
	}
	if actualFirstRow := actual.rows[0].text; actualFirstRow != "ABC" {
		t.Fatalf("expected first row %q, actual %q", "ABC", actualFirstRow)
	}
	if actualSecondRow := actual.rows[1].text; actualSecondRow != "DEF" {
		t.Fatalf("expected second row %q, actual %q", "DEF", actualSecondRow)
	}
	if actualStylePrefix := actual.lineStylePrefixes[0][0]; actualStylePrefix != styledPrefix {
		t.Fatalf("expected first rune style prefix %q, actual %q", styledPrefix, actualStylePrefix)
	}
	if actualStylePrefix := actual.lineStylePrefixes[0][1]; actualStylePrefix != styledPrefix {
		t.Fatalf("expected second rune style prefix %q, actual %q", styledPrefix, actualStylePrefix)
	}
	if actualStylePrefix := actual.lineStylePrefixes[0][2]; actualStylePrefix != "" {
		t.Fatalf("expected the reset rune style prefix %q, actual %q", "", actualStylePrefix)
	}
}
