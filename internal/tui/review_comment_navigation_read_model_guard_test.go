package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenReviewCommentNavigationFile_WhenScanning_ThenLegacyProgramSelectorHelpersMoveBehindTheSnapshot(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), "review_navigation.go"))
	then_noError(t, actualErr)

	forbiddenPatterns := []string{
		"func (program *Program) currentReviewCommentPosition(",
		"func (program *Program) reviewSessionCommentTarget(",
		"func (program *Program) reviewSessionCommentLocations(",
	}
	for _, forbiddenPattern := range forbiddenPatterns {
		if strings.Contains(string(contents), forbiddenPattern) {
			t.Fatalf("expected review_navigation.go to move %q behind the review comment snapshot", forbiddenPattern)
		}
	}
}
