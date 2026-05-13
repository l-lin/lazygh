package tui

import (
	"strings"
	"testing"
)

func TestRefactorGuard_GivenProductionFiles_WhenScanning_ThenOnlyCoreFilesBranchOnReviewSessionActiveAndLegacySideFocusIsGone(t *testing.T) {
	actualMatches := given_forbiddenTextMatchesInGoFiles(t, ".", []string{"reviewSession.active", "legacySideFocus("})
	allowedReviewSessionActivePrefixes := []string{
		"mode_descriptor.go contains \"reviewSession.active\"",
		"review_session.go contains \"reviewSession.active\"",
	}

	remainingMatches := make([]string, 0)
	for _, match := range actualMatches {
		if strings.Contains(match, "_test.go") {
			continue
		}
		allowed := false
		for _, prefix := range allowedReviewSessionActivePrefixes {
			if strings.HasPrefix(match, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			remainingMatches = append(remainingMatches, match)
		}
	}
	if len(remainingMatches) != 0 {
		t.Fatalf("expected shared screen queries to replace non-core reviewSession branches and legacy side focus, actual %v", remainingMatches)
	}
}
