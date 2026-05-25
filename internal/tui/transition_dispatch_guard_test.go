package tui

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyShellGlueCallsLayoutOrRefreshHelpers(t *testing.T) {
	allowedFiles := map[string]bool{
		"dispatch.go":      true,
		"shell_refresh.go": true,
	}

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`return program\.(?:layout|refreshViewsIfGUI|afterStateChange)\(gui\)`), func(path string) bool {
		base := filepath.Base(path)
		return strings.HasSuffix(base, ".go") && !strings.HasSuffix(base, "_test.go")
	})

	remainingMatches := make([]string, 0, len(actualMatches))
	for _, match := range actualMatches {
		base := filepath.Base(strings.Split(match, ":")[0])
		if allowedFiles[base] {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected direct layout/refresh helper returns to stay confined to shell glue, actual %v", remainingMatches)
	}
}
