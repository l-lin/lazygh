package github

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestArchitectureGuard_GivenGithubDomainFiles_WhenScanning_ThenTheyDoNotImportGithubcli(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGithubDomainGoFiles(t, ".", []string{"github.com/l-lin/lazygh/internal/githubcli"})

	if len(actualMatches) != 0 {
		t.Fatalf("expected github domain files to stay transport-neutral, actual %v", actualMatches)
	}
}

func given_forbiddenTextMatchesInGithubDomainGoFiles(t *testing.T, root string, forbidden []string) []string {
	t.Helper()

	packageRoot := given_githubDomainPackageRoot(t)
	scanRoot := filepath.Clean(filepath.Join(packageRoot, root))
	if filepath.IsAbs(root) {
		scanRoot = filepath.Clean(root)
	}
	actualMatches := make([]string, 0)
	actualErr := filepath.WalkDir(scanRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "testdata" {
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
		t.Fatalf("walk github domain files: %v", actualErr)
	}

	slices.Sort(actualMatches)
	return actualMatches
}

func given_githubDomainPackageRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve github domain test file path")
	}
	return filepath.Dir(file)
}
