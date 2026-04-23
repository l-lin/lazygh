package tui

import (
	"strings"
	"testing"

	appconfig "codeberg.org/l-lin/lazygh/internal/config"
)

func TestHelpPopup_GivenConfiguredKeyOverrides_WhenTogglingHelp_ThenItShowsTheConfiguredKeys(t *testing.T) {
	model := given_model()
	model.FocusPullRequestsView()
	subject := given_programWithKeymapOverrides(model, appconfig.KeymapOverrides{
		"global": {
			"quit": {"ctrl+x"},
		},
		"main": {
			"open_search": {"s"},
		},
	})
	gui := given_headlessGui(t)
	defer gui.Close()
	subject.configureGUI(gui)

	actualErr := subject.layout(gui)
	then_noError(t, actualErr)
	actualErr = subject.toggleHelp(gui, nil)
	then_noError(t, actualErr)

	helpView, actualErr := gui.View(viewHelpName)
	then_noError(t, actualErr)
	actualBuffer := helpView.Buffer()
	then_helpEntryUsesKey(t, actualBuffer, "Search pull requests", "s")
	then_helpEntryUsesKey(t, actualBuffer, "Quit", "<c-x>")
}

func then_helpEntryUsesKey(t *testing.T, buffer string, description string, expectedKey string) {
	t.Helper()

	for _, line := range strings.Split(buffer, "\n") {
		if !strings.HasSuffix(strings.TrimSpace(line), description) {
			continue
		}

		actualKey := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), description))
		if actualKey != expectedKey {
			t.Fatalf("expected help key %q for %q, actual %q", expectedKey, description, actualKey)
		}
		return
	}

	t.Fatalf("expected help entry for %q in %q", description, buffer)
}
