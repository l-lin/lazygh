package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenDetailImageSourceFiles_WhenScanning_ThenLegacyProgramImageSourceBuildersMoveBehindTheSnapshot(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), "detail_image_loader.go"))
	then_noError(t, actualErr)

	forbiddenPatterns := []string{
		"func (program *Program) currentDetailImageHTMLSources(",
		"func (program *Program) currentReviewSessionImageHTMLSources(",
		"func (program *Program) currentPullRequestImageHTMLSources(",
		"func (program *Program) pullRequestDescriptionImageHTMLSource(",
		"func (program *Program) pullRequestCommentsImageHTMLSources(",
		"func (program *Program) pullRequestCommitImageHTMLSources(",
		"func (program *Program) pullRequestDiffImageHTMLSources(",
		"func (program *Program) reviewDiffFileImageHTMLSources(",
		"func (program *Program) reviewDiffFileImageHTMLSourcesWithIndex(",
		"func (program *Program) pullRequestDiffFileIndex(",
		"func (program *Program) currentIssueImageHTMLSources(",
		"func (program *Program) currentReleaseImageHTMLSources(",
	}
	for _, forbiddenPattern := range forbiddenPatterns {
		if strings.Contains(string(contents), forbiddenPattern) {
			t.Fatalf("expected detail_image_loader.go to move %q behind the detail-image source snapshot", forbiddenPattern)
		}
	}
}
