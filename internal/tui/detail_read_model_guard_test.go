package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenDetailReadModelFiles_WhenScanning_ThenLegacyDetailReadMethodsMoveBehindTheSnapshot(t *testing.T) {
	forbiddenByFile := map[string][]string{
		"pull_request_detail_loader.go": {"func (program *Program) currentDetailIdentity("},
		"render_support.go":             {"func (program *Program) detailViewContent(", "func (program *Program) buildCurrentDetailDocument("},
		"view_selectors.go":             {"func (program *Program) currentDetailDocument("},
	}

	for relativePath, forbiddenPatterns := range forbiddenByFile {
		contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), relativePath))
		then_noError(t, actualErr)

		for _, forbiddenPattern := range forbiddenPatterns {
			if strings.Contains(string(contents), forbiddenPattern) {
				t.Fatalf("expected %q to move %q behind the detail read model snapshot", relativePath, forbiddenPattern)
			}
		}
	}
}
