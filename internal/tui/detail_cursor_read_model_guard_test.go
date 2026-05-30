package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenDetailCursorFiles_WhenScanning_ThenLegacyProgramCursorSelectorsMoveBehindTheSnapshot(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), "detail_cursor_selectors.go"))
	then_noError(t, actualErr)

	forbiddenPatterns := []string{
		"func (program *Program) currentDetailCursorSelection(",
		"func (program *Program) currentPullRequestDescriptionCursorContext(",
		"func (program *Program) currentBrowserChangesCursorContext(",
		"func (program *Program) currentReviewDiffCursorContext(",
	}
	for _, forbiddenPattern := range forbiddenPatterns {
		if strings.Contains(string(contents), forbiddenPattern) {
			t.Fatalf("expected detail_cursor_selectors.go to move %q behind the detail cursor snapshot", forbiddenPattern)
		}
	}
}
