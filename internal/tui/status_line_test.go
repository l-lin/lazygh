package tui

import (
	"strings"
	"testing"

	"github.com/jesseduffield/gocui"
)

func given_statusLineView(t *testing.T, gui *gocui.Gui) *gocui.View {
	t.Helper()

	statusView, actualErr := gui.View(viewStatusLineName)
	then_noError(t, actualErr)
	return statusView
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
