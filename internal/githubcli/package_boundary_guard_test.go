package githubcli

import "testing"

func TestArchitectureGuard_GivenGithubcliFiles_WhenScanning_ThenTheyDoNotDependOnTUIOrCache(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGithubcliGoFiles(t, ".", []string{
		"github.com/l-lin/lazygh/internal/tui",
		"github.com/l-lin/lazygh/internal/cache",
	})

	if len(actualMatches) != 0 {
		t.Fatalf("expected githubcli to stay adapter-only and not depend on tui or cache, actual %v", actualMatches)
	}
}
