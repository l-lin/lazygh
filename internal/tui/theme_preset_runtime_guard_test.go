package tui

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenThemePresetFiles_WhenScanning_ThenThemePresetStoreAccessStaysOnTheExplicitRuntimeSurface(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.themePresetStore`), func(path string) bool {
		switch filepath.Base(path) {
		case "theme_picker.go", "actions_popup_async_cmd.go", "theme_preset_runtime.go":
			return true
		default:
			return false
		}
	})

	for _, actualMatch := range actualMatches {
		if strings.Contains(actualMatch, "theme_preset_runtime.go") {
			continue
		}
		t.Fatalf("expected theme preset store reads to stay on the explicit runtime surface, actual %v", actualMatches)
	}
}
