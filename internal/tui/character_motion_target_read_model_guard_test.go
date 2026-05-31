package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenCharacterMotionTargetSelectorFile_WhenScanning_ThenLegacyProgramWrappersMoveBehindTheSnapshot(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), "detail_character_motion_selectors.go"))
	then_noError(t, actualErr)

	forbiddenPatterns := []string{
		"func (program *Program) currentDetailCharacterMotionTargetRunes(",
		"func (program *Program) currentPullRequestBuildRunPopupCharacterMotionTargetRunes(",
	}
	for _, forbiddenPattern := range forbiddenPatterns {
		if strings.Contains(string(contents), forbiddenPattern) {
			t.Fatalf("expected detail_character_motion_selectors.go to move %q behind the character-motion target snapshot", forbiddenPattern)
		}
	}
}
