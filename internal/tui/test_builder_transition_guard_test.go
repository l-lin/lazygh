package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenTestAppDepsBuilder_WhenScanning_ThenLegacyGithubcliTestAdaptersAreGone(t *testing.T) {
	builderPath := filepath.Join(given_guardPackageRoot(t), "test_app_deps_builder_test.go")
	contents, actualErr := os.ReadFile(builderPath)
	if actualErr != nil {
		t.Fatalf("read test app deps builder: %v", actualErr)
	}

	actualMatches := make([]string, 0)
	for _, pattern := range []string{
		"type testLegacy",
		"github.com/l-lin/lazygh/internal/githubcli",
	} {
		if strings.Contains(string(contents), pattern) {
			actualMatches = append(actualMatches, pattern)
		}
	}
	if len(actualMatches) != 0 {
		t.Fatalf("expected test app deps builder to stop carrying legacy githubcli adapters, actual %v", actualMatches)
	}
}
