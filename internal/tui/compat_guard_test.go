package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenTUIFiles_WhenScanning_ThenLegacyConstructorAndCompatLoaderAreGone(t *testing.T) {
	for _, path := range []string{"deps_compat.go", "deps_compat_adapters.go"} {
		_, actualErr := os.Stat(filepath.Join(given_guardPackageRoot(t), path))
		if !errors.Is(actualErr, os.ErrNotExist) {
			t.Fatalf("expected %q to be deleted, actual error %v", path, actualErr)
		}
	}

	actualMatches := given_forbiddenTextMatchesInGoFiles(t, ".", []string{
		"NewProgramWithModelAnd" + "Loader(",
		"appDepsFromCompatibility" + "Loader(",
	})

	if len(actualMatches) != 0 {
		t.Fatalf("expected no legacy constructor or compat-loader references, actual %v", actualMatches)
	}
}

func given_forbiddenTextMatchesInGoFiles(t *testing.T, root string, forbidden []string) []string {
	t.Helper()

	packageRoot := given_guardPackageRoot(t)
	scanRoot := given_guardScanRoot(t, root)
	actualMatches := make([]string, 0)
	actualErr := filepath.WalkDir(scanRoot, func(path string, entry os.DirEntry, walkErr error) error {
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
		if filepath.Ext(path) != ".go" {
			return nil
		}

		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		relativePath, relErr := filepath.Rel(packageRoot, path)
		if relErr != nil {
			return relErr
		}
		for _, pattern := range forbidden {
			if strings.Contains(string(contents), pattern) {
				actualMatches = append(actualMatches, fmt.Sprintf("%s contains %q", filepath.ToSlash(relativePath), pattern))
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
