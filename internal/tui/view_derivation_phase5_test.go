package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestViewDerivation_GivenRenderHelperSourceFiles_WhenInspecting_ThenTheyDoNotOwnSelectorCachingOrTerminalImageSync(t *testing.T) {
	forbiddenByFile := map[string][]string{
		"render.go": {"detailImageManager.Sync"},
		"render_support.go": {
			"currentPullRequestDetailDocumentCacheKey(",
			"pullRequestDetailDocumentForKey(",
			"cachePullRequestDetailDocument(",
		},
		"selectable_list_view.go": {"consumePendingListViewportPlacement("},
	}

	for relativePath, forbiddenPatterns := range forbiddenByFile {
		contents, actualErr := os.ReadFile(filepath.Join(given_guardPackageRoot(t), relativePath))
		then_noError(t, actualErr)

		for _, forbiddenPattern := range forbiddenPatterns {
			if strings.Contains(string(contents), forbiddenPattern) {
				t.Fatalf("expected %q to stay free of %q, actual source:\n%s", relativePath, forbiddenPattern, string(contents))
			}
		}
	}
}
