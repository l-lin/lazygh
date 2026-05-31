package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenBrowserChangesFiles_WhenScanning_ThenLegacyProgramToggleHelpersMoveBehindTheSnapshot(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), "browser_changes.go"))
	then_noError(t, actualErr)

	forbiddenPatterns := []string{
		"func (program *Program) browserCollapsedChangesFileIDs(",
		"func (program *Program) browserCollapsedChangesThreadIDs(",
		"func (program *Program) toggleBrowserChangesVisibility(",
		"func (program *Program) toggleBrowserChangesFileVisibility(",
		"func (program *Program) toggleBrowserChangesThreadVisibility(",
	}
	for _, forbiddenPattern := range forbiddenPatterns {
		if strings.Contains(string(contents), forbiddenPattern) {
			t.Fatalf("expected browser_changes.go to move %q behind the browser changes snapshot", forbiddenPattern)
		}
	}
}
