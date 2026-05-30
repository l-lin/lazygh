package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenBrowserDetailSectionFiles_WhenScanning_ThenLegacyProgramSectionBuildersMoveBehindTheSnapshot(t *testing.T) {
	forbiddenByFile := map[string][]string{
		"browser_detail_sections.go": {
			"func (program *Program) browserDetailSectionCollapsed(",
			"func (program *Program) currentPullRequestOverviewSections(",
			"func (program *Program) renderCurrentPullRequestOverview(",
			"func (program *Program) browserOverviewSectionAtCursor(",
			"func (program *Program) buildPullRequestConversationSections(",
			"func (program *Program) currentPullRequestConversationSections(",
			"func (program *Program) renderCurrentPullRequestConversationsTab(",
			"func (program *Program) browserConversationSectionAtCursor(",
		},
	}

	for relativePath, forbiddenPatterns := range forbiddenByFile {
		contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), relativePath))
		then_noError(t, actualErr)

		for _, forbiddenPattern := range forbiddenPatterns {
			if strings.Contains(string(contents), forbiddenPattern) {
				t.Fatalf("expected %q to move %q behind the browser detail snapshot", relativePath, forbiddenPattern)
			}
		}
	}
}
