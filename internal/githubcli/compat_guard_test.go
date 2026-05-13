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
	for _, path := range []string{"client.go", "provider_facade.go"} {
		_, actualErr := os.Stat(filepath.Join(given_guardPackageRoot(t), path))
		if !errors.Is(actualErr, os.ErrNotExist) {
			t.Fatalf("expected %q to be deleted, actual error %v", path, actualErr)
		}
	}

	actualMatches := given_forbiddenTextMatchesInGithubcliSourceGoFiles(t, ".", []string{
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

func TestRefactorGuard_GivenGithubcliTestFiles_WhenScanning_ThenBroadClientTestWrappersAreGone(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGithubcliTestGoFiles(t, ".", []string{
		"type " + "testClient struct",
		"New" + "Client(",
		"NewClientWith" + "Runner(",
	})

	if len(actualMatches) != 0 {
		t.Fatalf("expected githubcli tests to use focused services instead of client wrappers, actual %v", actualMatches)
	}
}

func given_forbiddenTextMatchesInGithubcliSourceGoFiles(t *testing.T, root string, forbidden []string) []string {
	t.Helper()

	return given_forbiddenTextMatchesInGithubcliFiles(t, root, forbidden, func(path string) bool {
		return filepath.Ext(path) == ".go" && !strings.HasSuffix(path, "_test.go")
	})
}

func given_forbiddenTextMatchesInGithubcliTestGoFiles(t *testing.T, root string, forbidden []string) []string {
	t.Helper()

	return given_forbiddenTextMatchesInGithubcliFiles(t, root, forbidden, func(path string) bool {
		return filepath.Ext(path) == ".go" && strings.HasSuffix(path, "_test.go")
	})
}

func given_forbiddenTextMatchesInGithubcliFiles(t *testing.T, root string, forbidden []string, includeFile func(path string) bool) []string {
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
		if !includeFile(path) {
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
