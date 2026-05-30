package tui

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenDetailImageWorkflowFiles_WhenScanning_ThenGitHubAuthTokenLookupStaysOnTheExplicitRuntimeSurface(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(`program\.authTokenProvider\.GetAuthToken\(`), func(path string) bool {
		base := filepath.Base(path)
		return base == "detail_image_loader.go" || base == "workflow_detail_image_commands.go"
	})

	if len(actualMatches) != 1 || !strings.Contains(actualMatches[0], "workflow_detail_image_commands.go") {
		t.Fatalf("expected exactly one direct auth token provider call on the detail-image workflow runtime surface, actual %v", actualMatches)
	}
}
