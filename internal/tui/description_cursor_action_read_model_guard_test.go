package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenDescriptionCursorActionFiles_WhenScanning_ThenLegacyProgramLookupsMoveBehindTheSnapshot(t *testing.T) {
	forbiddenByFile := map[string][]string{
		"pull_request_build.go": {
			"func (program *Program) browserOverviewBuildEntryAtDetailCursorDocument(",
			"func (program *Program) currentPullRequestDescriptionSummaryAndDetail(",
			"func (program *Program) detailCursorActionsAvailable(",
		},
		"pull_request_reviewer.go": {
			"func (program *Program) browserOverviewReviewerEntryAtDetailCursorDocument(",
			"func (program *Program) currentPullRequestDescriptionSummaryAndDetail(",
			"func (program *Program) detailCursorActionsAvailable(",
		},
	}

	for relativePath, forbiddenPatterns := range forbiddenByFile {
		contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), relativePath))
		then_noError(t, actualErr)

		for _, forbiddenPattern := range forbiddenPatterns {
			if strings.Contains(string(contents), forbiddenPattern) {
				t.Fatalf("expected %q to move %q behind the description cursor action snapshot", relativePath, forbiddenPattern)
			}
		}
	}
}
