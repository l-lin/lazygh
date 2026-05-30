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

func TestRefactorGuard_GivenCacheConfigEntryPoint_WhenScanning_ThenItStopsOwningOpenAndCloseIO(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`persistcache\.Open\(`,
		`OpenNotificationDoneStore\(`,
		`\.Close\(`,
	}, "|")), func(path string) bool {
		return filepath.Base(path) == "cache_config.go"
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected cache_config.go to delegate cache open and close IO to an explicit runtime surface, actual %v", actualMatches)
	}
}

func TestRefactorGuard_GivenLinkOpenerRuntimeFiles_WhenScanning_ThenLinksConfigApplyStopsTypeAssertingAndMutatingSystemOpenersInPlace(t *testing.T) {
	actualMatches := given_regexpLineMatchesInGoFiles(t, ".", regexp.MustCompile(strings.Join([]string{
		`updateSystemLinkOpenerCommand\(`,
		`program\.linkOpener\.\(\*systemLinkOpener\)`,
	}, "|")), func(path string) bool {
		switch filepath.Base(path) {
		case "runtime_dependency_adapter.go", "update_runtime_config.go":
			return true
		default:
			return false
		}
	})
	if len(actualMatches) != 0 {
		t.Fatalf("expected links config apply to replace system link openers through one explicit seam, actual %v", actualMatches)
	}
}
