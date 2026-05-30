package tui

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRefactorGuard_GivenCacheAndLinkRuntimeFiles_WhenScanning_ThenRuntimeConfigApplyStopsReplacingShellDepsInline(t *testing.T) {
	forbiddenPattern := regexp.MustCompile(strings.Join([]string{
		`program\.(?:pullRequestCache|notificationDoneStore|linkOpener)\s*=\s*[^=]`,
	}, "|"))

	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", forbiddenPattern, func(path string) bool {
		base := filepath.Base(path)
		return base == "cache_config.go" || base == "link_opener.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cache and links runtime config entrypoints to stop replacing shell dependencies and stores inline, actual %v", actualMatches)
	}
}
