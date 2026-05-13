package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenTUIProviderBoundaryFiles_WhenScanning_ThenGithubcliTransitionFilesAreGone(t *testing.T) {
	for _, path := range []string{"provider_boundary.go", "provider_transition.go"} {
		_, actualErr := os.Stat(filepath.Join(given_guardPackageRoot(t), path))
		if !errors.Is(actualErr, os.ErrNotExist) {
			t.Fatalf("expected %q to be deleted, actual error %v", path, actualErr)
		}
	}

	actualMatches := given_forbiddenTextMatchesInGoFiles(t, ".", []string{"github.com/l-lin/lazygh/internal/githubcli"})
	remainingMatches := make([]string, 0)
	for _, match := range actualMatches {
		if strings.Contains(match, "_test.go") {
			continue
		}
		remainingMatches = append(remainingMatches, match)
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected production TUI files to stop importing githubcli, actual %v", remainingMatches)
	}
}
