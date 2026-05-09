package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseProjectFiles_GivenMiseToml_WhenReadingTheTasks_ThenItDefinesGoReleaserValidationAndSnapshotTasks(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", "mise.toml"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"[tasks.release-check]",
		"ghcr.io/goreleaser/goreleaser-cross:v1.25.9",
		"check --config .goreleaser.yaml",
		"[tasks.release-snapshot]",
		"release --snapshot --clean --skip=publish --config .goreleaser.yaml",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected mise.toml to contain %q, actual %q", expected, actual)
		}
	}
}

func TestReleaseProjectFiles_GivenTheRepo_WhenInspectingTheGoReleaserConfig_ThenItDefinesCrossPlatformArchivesForGitHubAndCodeberg(t *testing.T) {
	contents, actualErr := os.ReadFile(filepath.Join("..", "..", ".goreleaser.yaml"))
	then_noError(t, actualErr)

	actual := string(contents)
	for _, expected := range []string{
		"version: 2",
		"project_name: lazygh",
		"gitea_urls:",
		"api: https://codeberg.org/api/v1",
		"download: https://codeberg.org",
		"CGO_ENABLED=1",
		"CC_darwin_amd64=o64-clang",
		"CC_darwin_arm64=oa64-clang",
		"CC_linux_amd64=x86_64-linux-gnu-gcc",
		"CC_linux_arm64=aarch64-linux-gnu-gcc",
		"CC_windows_amd64=x86_64-w64-mingw32-gcc",
		"ignore:",
		"goos: windows",
		"goarch: arm64",
		"goos:",
		"- linux",
		"- darwin",
		"- windows",
		"goarch:",
		"- amd64",
		"- arm64",
		"format_overrides:",
		"goos: windows",
		"formats: [zip]",
		"checksum:",
		"name_template: checksums.txt",
	} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("expected .goreleaser.yaml to contain %q, actual %q", expected, actual)
		}
	}
}
