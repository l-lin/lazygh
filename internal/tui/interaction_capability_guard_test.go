package tui

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestRefactorGuard_GivenInteractionCapabilityFiles_WhenScanning_ThenPopupHelpersAndDeadViewGlueStopReadingShellCapabilitiesDirectly(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.linkOpener == nil|program\.clipboardReader\.ReadText|func \(program \*Program\) clipboardPullRequestURL\(`), func(path string) bool {
		switch filepath.Base(path) {
		case "open_link.go", "notification_actions.go", "update_interaction.go", "view_url_prompt.go":
			return true
		default:
			return false
		}
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected popup helpers, interaction preflight, and dead view glue to stop reading shell capabilities directly, actual %v", actualMatches)
	}
}
