package tui

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
)

func given_statusLineView(t *testing.T, gui *gocui.Gui) *gocui.View {
	t.Helper()

	statusView, actualErr := gui.View(viewStatusLineName)
	then_noError(t, actualErr)
	return statusView
}

func given_statusLineKeyHintsView(t *testing.T, gui *gocui.Gui) *gocui.View {
	t.Helper()

	keyHintsView, actualErr := gui.View(viewStatusLineKeyHintsName)
	then_noError(t, actualErr)
	return keyHintsView
}

func then_statusLineContains(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	statusView := given_statusLineView(t, gui)
	if !strings.Contains(statusView.Buffer(), expected) {
		t.Fatalf("expected status line to contain %q, actual %q", expected, statusView.Buffer())
	}
}

func then_statusLineIs(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	statusView := given_statusLineView(t, gui)
	if actual := strings.TrimSpace(statusView.Buffer()); actual != expected {
		t.Fatalf("expected status line %q, actual %q", expected, actual)
	}
}

func then_statusLineKeyHintsAre(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	keyHintsView := given_statusLineKeyHintsView(t, gui)
	if actual := strings.TrimSpace(keyHintsView.Buffer()); actual != expected {
		t.Fatalf("expected status line key hints %q, actual %q", expected, actual)
	}
}

func then_statusLineKeyHintsAreRightAligned(t *testing.T, gui *gocui.Gui, expected string) {
	t.Helper()

	keyHintsView := given_statusLineKeyHintsView(t, gui)
	if actual := strings.TrimSpace(keyHintsView.Buffer()); actual != expected {
		t.Fatalf("expected status line key hints %q, actual %q", expected, actual)
	}
	if keyHintsView.InnerWidth() != utf8.RuneCountInString(expected) {
		t.Fatalf("expected key hints width %d, actual %d", utf8.RuneCountInString(expected), keyHintsView.InnerWidth())
	}

	_, _, maxX, _, actualErr := gui.ViewPosition(viewStatusLineName)
	then_noError(t, actualErr)
	_, _, keyHintsX1, _, actualErr := gui.ViewPosition(viewStatusLineKeyHintsName)
	then_noError(t, actualErr)
	if keyHintsX1 != maxX {
		t.Fatalf("expected key hints right edge %d, actual %d", maxX, keyHintsX1)
	}
}
