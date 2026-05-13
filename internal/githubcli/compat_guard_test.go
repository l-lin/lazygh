package githubcli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenGithubcliFiles_WhenScanning_ThenCompatibilityClientFacadeAndRunGHWrappersAreGone(t *testing.T) {
	for _, path := range []string{"provider_facade.go"} {
		_, actualErr := os.Stat(path)
		if !errors.Is(actualErr, os.ErrNotExist) {
			t.Fatalf("expected %q to be deleted, actual error %v", path, actualErr)
		}
	}

	actualMatches := given_forbiddenTextMatchesInGithubcliGoFiles(t, ".", []string{
		"type " + "Client struct",
		"type " + "ProviderFacade struct",
		"New" + "Client(",
		"NewProvider" + "Facade(",
		"run" + "GH(",
		"runGHWith" + "Input(",
	})

	if len(actualMatches) != 0 {
		t.Fatalf("expected no githubcli compatibility shims, actual %v", actualMatches)
	}
}

func given_forbiddenTextMatchesInGithubcliGoFiles(t *testing.T, root string, forbidden []string) []string {
	t.Helper()

	actualMatches := make([]string, 0)
	actualErr := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		for _, pattern := range forbidden {
			if strings.Contains(string(contents), pattern) {
				actualMatches = append(actualMatches, fmt.Sprintf("%s contains %q", path, pattern))
			}
		}
		return nil
	})
	if actualErr != nil {
		t.Fatalf("walk go files: %v", actualErr)
	}

	slices.Sort(actualMatches)
	return actualMatches
}
